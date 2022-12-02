package messagequeue

import (
	"sync"
	"sync/atomic"
)

// Work done by a worker, returns true if the worker should continue, or false
// if the worker should exit
type Work func() bool

func Stop() bool {
	return false
}

// Worker can be used to asynchronously call work added via AddWork until
// work returns false.
type Worker struct {
	work   chan *Work
	pool   chan<- *Worker
	closed atomic.Bool
}

// Add work to this worker. If the work returns false this workers
// routine will exit.
// This function will block.
func (w *Worker) Add(work *Work) {
	w.work <- work
}

// Exit this workers internal routing once current processing ends.
// This function will not block.
func (w *Worker) Exit() {
	if !w.closed.Load() {
		var stop Work = Stop
		w.work <- &stop
		close(w.work)
		w.closed.Store(true)
	}
}

// AddWorker to the passed pool
func AddWorker(wg *sync.WaitGroup, pool chan<- *Worker) {
	w := &Worker{
		pool: pool,
		work: make(chan *Work, 1),
	}
	wg.Add(1)
	w.pool <- w
	go func() {
		for work, ok := <-w.work; ok && (*work)(); work, ok = <-w.work {
			w.pool <- w
		}
		wg.Done()
		w.work = nil
	}()
}
