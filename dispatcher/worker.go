package dispatcher

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/beaconsoftwarellc/gadget/log"
)

// DefaultTaskTimeout is the timeout for tasks that do not define their own
// timeout.
const DefaultTaskTimeout = 10 * time.Second

type dummy struct{}

func (task *dummy) Execute() error {
	return nil
}

// Worker is used internally by the dispatched to execute tasks in ready
// go routines
type Worker interface {
	// Exec the passed message asynchronously
	Exec(*internalTask) bool
	// Run this worker so that it will perform work
	Run() <-chan bool
	// Quit this workers main event loop. This function block until
	// the current task has completed executing.
	Quit()
}

type worker struct {
	// > 0 indicates that this worker is 'running'
	running int32
	// the channel that this worker will add itself into
	pool chan Worker
	// mutex used for locking changes to the pool
	mux sync.Mutex
	// the channel this instance uses to receive messages
	tasks chan *internalTask
	// the channel used to tell this instance to exit it's main loop, see Quit below
	exit chan bool
	// the channel used to signal that this worker has exited it's main loop
	exited chan bool
	// channel where we put completed tasks.
	complete chan<- *internalTask
}

// NewWorker for the passed worker pool.
func NewWorker(pool chan Worker, complete chan<- *internalTask) Worker {
	worker := &worker{
		pool:     pool,
		running:  0,
		complete: complete,
	}
	return worker
}

func (w *worker) Running() bool {
	return (atomic.LoadInt32(&w.running)) > 0
}

func (w *worker) Exec(t *internalTask) bool {
	if !w.Running() {
		log.Warnf("attempt to exec a task on a stopped worker")
		return false
	}
	success := false
	if w.Running() {
		select {
		// non-blocking put into channel, neat!
		case w.tasks <- t:
			success = true
		default:
			// this is not an error, it can happen if the worker has been added
			// but is not yet listening
			log.Debugf("failed to add task to worker channel")
			success = false
		}
	}
	return success
}

func (w *worker) run(loaded chan bool) {
	exit := false
	notified := false
	for {
		// if we are not exiting put ourselves back into the pool to receive tasks
		if !exit {
			w.pool <- w
			if !notified {
				loaded <- true
				notified = true
			}
		}
		// This select cannot be non-blocking or the 'get' from tasks will
		// never succeed
		select {
		case task := <-w.tasks:
			log.Error(task.Execute())
			w.completeTask(task)
			if exit {
				w.exited <- true
				return
			}
		case <-w.exit:
			// set exit and run through one more cycle to make sure we do not have
			// a task in our channel
			exit = true
		}
	}
}

func (w *worker) completeTask(task *internalTask) {
	select {
	case w.complete <- task:
		return
	default:
		return
	}
}

func (w *worker) Run() <-chan bool {
	loaded := make(chan bool, 2)
	if w.Running() {
		log.Infof("run called on an already running worker")
		loaded <- true
		return loaded
	}
	w.tasks = make(chan *internalTask)
	w.exit = make(chan bool)
	w.exited = make(chan bool)
	atomic.StoreInt32(&w.running, 1)
	go w.run(loaded)
	return loaded
}

func (w *worker) Quit() {
	if w.Running() {
		atomic.StoreInt32(&w.running, 0)
		w.exit <- true
		// ensure we have a task in the channel to trigger the quit
		w.tasks <- newInternalTask(&dummy{})
		<-w.exited
		close(w.tasks)
		close(w.exit)
		close(w.exited)
	}
}
