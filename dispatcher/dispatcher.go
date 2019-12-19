package dispatcher

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/beaconsoftwarellc/gadget/intutil"
	"github.com/beaconsoftwarellc/gadget/log"
	"github.com/beaconsoftwarellc/gadget/timeutil"
)

// Status is the how the dispatcher is currently functioning.
type Status int32

const (
	// Draining indicates that the dispatcher is stopping and waiting for all
	// pending tasks to exit
	Draining Status = 3
	// Running indicates that the dispatcher is receiving and executing tasks
	Running Status = 2
	// Stopping indicates that the dispatcher is stopping workers after their
	// current task finishes.
	Stopping Status = 1
	// Stopped indicates that the Dispatcher is not currently executing tasks.
	Stopped Status = 0
	// DefaultWaitBetweenScaleDowns is the time duration to wait between
	// attempts to scale down the worker pool. You do not want this to be too small
	// or a lot of time will be wasted scaling up and down.
	DefaultWaitBetweenScaleDowns = 5 * time.Minute
	// DefaultDispatchMissesBeforeDraining is the number of times the system
	// must fail to dispatch a message from the task channel prior to considering the
	// system fully drained
	DefaultDispatchMissesBeforeDraining = 5
)

// Dispatcher is responsible for dispatching work to be processed asynchronously.
type Dispatcher interface {
	// Run this dispatcher by starting up all it's workers and executing any
	// tasks in it's queue. The dispatcher will run asynchronously until Quit is
	// called. Run is reentrant.
	Run()
	// Dispatch is non-blocking and will accept tasks up to the maximum number of
	// buffered tasks even if the dispatcher is not currently running. Returns true
	// if the queue is not currently full and false otherwise. If the queue is full,
	// increase the number of workers or run the dispatcher.
	Dispatch(task Task) bool
	// Quit running and stop all workers.
	Quit(drain bool)
}

type dispatcher struct {
	// buffered up to the maximum amount of queued messages
	queue      chan *internalTask
	pool       chan Worker
	complete   chan *internalTask
	overflow   TaskStack
	workers    []Worker
	exit       chan bool
	exited     chan bool
	drain      chan bool
	minWorkers int
	maxWorkers int
	bufferSize int
	// > 0 indicates we are running
	running int32
	// > 0 indicates we are currently scaling
	scaling                    int32
	mutex                      sync.RWMutex
	waitBetweenScaleDowns      time.Duration
	consecutiveScaleDownMisses int
	etMux                      sync.Mutex
	executingTasks             map[string]*internalTask
}

// NewDispatcher to handle asynchronous processing of Tasks with the specified maximum number of workers.
// In order to use the dispatcher to process work, callers must implement the `Task` interface.
// Execution for the dispatcher is asynchronous but `Dispatcher.Run` must be called for any tasks to be worked.
// Example Usage:
//
//      type MyTask struct {}
//
//      func (task *MyTask) Execute() error {
//          // do work here
//          return nil
//      }
//      func main() {
//             maxBufferedMessages := 10
//			   minWorkers := 1
//			   maxWorkers := 10
//             myDispatcher := dispatcher.NewDispatcher(maxBufferedMessages, minWorkers, maxWorkers)
//             myDispatcher.Run()
//             myTask := &MyTask{}
//             myDispatcher.Dispatch(myTask)
//             myDispatcher.Quit()
//      }
func NewDispatcher(maxBufferedMessage int, minWorkers int, maxWorkers int) Dispatcher {
	d := &dispatcher{
		bufferSize: maxBufferedMessage,
		queue:      make(chan *internalTask, maxBufferedMessage),
		complete:   make(chan *internalTask, maxBufferedMessage),
		// don't set min below 0
		minWorkers:                 intutil.Maxv(0, minWorkers),
		overflow:                   NewTaskStack(),
		waitBetweenScaleDowns:      DefaultWaitBetweenScaleDowns,
		consecutiveScaleDownMisses: DefaultDispatchMissesBeforeDraining,
		executingTasks:             make(map[string]*internalTask),
	}
	// don't set max below min
	d.maxWorkers = intutil.Maxv(d.minWorkers, maxWorkers)
	return d
}

func (d *dispatcher) overflowPush(t *internalTask) {
	d.overflow.Push(t)
	if d.overflow.Size()%20 == 0 {
		d.etMux.Lock()
		ets := make([]string, len(d.executingTasks))
		i := 0
		for _, task := range d.executingTasks {
			ets[i] = fmt.Sprintf("%#v", task)
			i++
		}
		log.Errorf("Overflow activity at threshold, current tasks:\n%s", strings.Join(ets, "\n"))
		d.etMux.Unlock()
	}
}

func (d *dispatcher) Status() Status {
	return Status(atomic.LoadInt32(&d.running))
}

func (d *dispatcher) Resize(size int, start bool) {
	if atomic.LoadInt32(&d.scaling) == 0 {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		atomic.StoreInt32(&d.scaling, 1)
		if size > d.maxWorkers {
			size = d.maxWorkers
		}
		if size < d.minWorkers {
			size = d.minWorkers
		}
		if size == len(d.workers) {
			log.Warnf("cannot resize pool to it's current size %d", len(d.workers))
			return
		}
		if size < 0 {
			log.Warnf("pool size cannot be less than zero")
			return
		}
		log.Debugf("scaling worker pool %d -> %d", len(d.workers), size)

		// kill all the existing workers
		for _, w := range d.workers {
			w.Quit()
		}
		newPool := make(chan Worker, size)
		d.workers = make([]Worker, size)
		// add in the new workers
		for i := 0; i < len(d.workers); i++ {
			d.workers[i] = NewWorker(newPool, d.complete)
		}
		// no one is using the old pool anymore so close it
		if nil != d.pool {
			close(d.pool)
		}
		d.pool = newPool
		if start {
			for _, w := range d.workers {
				<-w.Run()
			}
		}
		log.Debugf("scaling complete")
		atomic.StoreInt32(&d.scaling, 0)
	}
}

// Dispatch is non-blocking and will accept buffer tasks for asynchronous execution
// by this dispatchers workers up to the maximum buffered tasks. If the maximum buffered
// tasks is reached and a non-blocking put cannot succeed on the task queue an overflow buffer
// will be used to store the additional tasks which will be loaded onto the queue as space becomes available,
// this will have negative performance implications.
// If overflow occurs regularly in production set max buffered tasks to a higher value
// when initializing the dispatcher.
// NOTE: If this dispatcher is currently draining you should not dispatch more tasks, as this will
// prevent the Quit function from exiting.
func (d *dispatcher) Dispatch(task Task) bool {
	if d.Status() == Draining {
		log.Error(fmt.Errorf("task added to dispatcher while draining: %+v", task))
	}
	return d.enqueue(newInternalTask(task), false)
}

func (d *dispatcher) enqueue(task *internalTask, suppressWarning bool) bool {
	select {
	case d.queue <- task:
		return true
	default:
		d.overflowPush(task)
		if !suppressWarning {
			log.Warnf("task added to dispatcher with a full queue, overflow is at %d ", d.overflow.Size())
		}
		return false
	}
}

func sendNonBlocking(value bool, ch chan bool) bool {
	select {
	case ch <- value:
		return true
	default:
		return false
	}
}

func (d *dispatcher) loadOverflow() bool {
	if d.overflow.Size() == 0 {
		return true
	}
	for t, e := d.overflow.Pop(); nil == e; t, e = d.overflow.Pop() {
		// we know we are probably going to hit overflow, so suppress the warning
		if !d.enqueue(t, true) {
			return false
		}
	}
	return true
}

func (d *dispatcher) Run() {
	// run while draining or stopping would cause all kinds of problems
	if d.Status() != Stopped {
		return
	}
	atomic.StoreInt32(&d.running, int32(Running))
	d.drain = make(chan bool, 2)
	d.exit = make(chan bool, 2)
	d.exited = make(chan bool)
	// resize to our min to get the correct amount of workers
	d.Resize(d.minWorkers, true)
	go d.run()
}

func (d *dispatcher) run() {
	// This CANNOT be non-blocking or we will just spin and use a ton of CPU
	var consecutiveMisses int
	var lastDispatch time.Time
	ticker := timeutil.NewTicker(d.waitBetweenScaleDowns).Start()
	defer ticker.Stop()
	for {
		select {
		case <-d.exit:
			d.exited <- true
			return
		case <-d.drain:
			consecutiveMisses++
			// try to load the overflow
			if d.overflow.Size() == 0 && consecutiveMisses > d.consecutiveScaleDownMisses {
				log.Infof("exiting as there are no more tasks")
				d.exited <- true
				return
			}
			// we want this to keep catching
			sendNonBlocking(true, d.drain)
		case task := <-d.complete:
			d.etMux.Lock()
			delete(d.executingTasks, task.ID)
			d.etMux.Unlock()
			consecutiveMisses = 0
			if d.overflow.Size() > 0 && !d.loadOverflow() {
				log.Infof("dispatcher overflow queue at %d messages after load", d.overflow.Size())
			}
		case task := <-d.queue:
			lastDispatch = time.Now()
			consecutiveMisses = 0
			d.dispatch(task)
		case <-ticker.Channel():
			// scale down if we have no overflow queue, we are not at our minimum number of workers
			// and it has been over a second since the last time we dispatched a message
			if d.overflow.Size() == 0 && len(d.workers) != d.minWorkers && time.Since(lastDispatch) > d.waitBetweenScaleDowns {
				log.Infof("dispatcher status: %d workers %d overflow", len(d.workers), d.overflow.Size())
				d.Resize(len(d.workers)/2, true)
			}
		}
	}
}

func (d *dispatcher) dispatch(t *internalTask) {
	select {
	// wait for a worker to become available
	case w := <-d.pool:
		// this should only return false when we somehow got a worker that
		// is not accepting requests
		// success = w.Exec(t)
		d.etMux.Lock()
		d.executingTasks[t.ID] = t
		d.etMux.Unlock()
		if !w.Exec(t) {
			// put it in overflow for now
			log.Warnf("worker exec failed, pushing task to overflow (%d tasks)", d.overflow.Size())
			d.overflowPush(t)
		}
	case <-d.exit:
		// add another true onto the channel so callers above us get the message
		// and then bail
		d.exit <- true
		return
	default:
		// no workers, try scaling up if we are not already at capacity
		if len(d.workers) != d.maxWorkers {
			log.Infof("No workers available scaling pool")
			d.Resize(2*len(d.workers), true)
		}
		// but either way push to overflow and try later
		d.overflowPush(t)

	}
}

func (d *dispatcher) Quit(drain bool) {
	// set the running to false
	if d.Status() == Running {
		if drain {
			atomic.StoreInt32(&d.running, int32(Draining))
			d.drain <- true
		} else {
			atomic.StoreInt32(&d.running, int32(Stopping))
			d.exit <- true
		}
		<-d.exited
		d.Resize(0, false)
		// if we set this prior to being done, a run command will break things.
		atomic.StoreInt32(&d.running, int32(Stopped))
	}
}
