package messagequeue

import (
	"sync"
	"sync/atomic"
)

// Job done by a worker, returns true if the worker should continue, or false
// if the worker should exit
type Job interface {
	Work() bool
}

type stop struct{}

func (s *stop) Work() bool {
	return false
}

// Worker can be used to asynchronously call work added via AddWork until
// work returns false.
type Worker struct {
	jobs   chan Job
	pool   chan<- *Worker
	closed atomic.Bool
}

// Add work to this worker. If the work returns false this workers
// routine will exit.
// This function will block.
func (w *Worker) Add(job Job) {
	w.jobs <- job
}

// Exit this workers internal routing once current processing ends.
// This function will not block.
func (w *Worker) Exit() {
	if !w.closed.Load() {
		w.jobs <- &stop{}
		close(w.jobs)
		w.closed.Store(true)
	}
}

// AddWorker to the passed pool
func AddWorker(wg *sync.WaitGroup, pool chan<- *Worker) {
	w := &Worker{
		pool: pool,
		jobs: make(chan Job, 1),
	}
	wg.Add(1)
	w.pool <- w
	go func() {
		for job, ok := <-w.jobs; ok && job.Work(); job, ok = <-w.jobs {
			w.pool <- w
		}
		wg.Done()
		w.jobs = nil
	}()
}
