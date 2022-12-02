package messagequeue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

func newWorker(handler HandleMessage) *worker {
	return &worker{
		handler: handler,
	}
}

type worker struct {
	handler HandleMessage
	pool    chan<- *worker
	ctrl    *atomic.Uint32
	wg      *sync.WaitGroup
}

func (w *worker) Run(wg *sync.WaitGroup, ctrl *atomic.Uint32, pool chan<- *worker) {
	w.wg.Add(1)
	w.pool = pool
	w.ctrl = ctrl
	w.wg = wg
	w.pool <- w
}

func (w *worker) HandleMessage(message *Message) {
	var (
		ctx    = context.Background()
		cancel context.CancelFunc
	)
	if message.Deadline.After(time.Now()) {
		ctx, cancel = context.WithDeadline(ctx, message.Deadline)
		defer cancel()
	}
	w.handler(ctx, message)
	if w.ctrl.Load() != statusRunning {
		select {
		case w.pool <- w:
		default:
			// the pool was inappropriately sized or we got an extra worker
			// don't block, just exit.
			w.wg.Done()
			return
		}
	} else {
		w.wg.Done()
	}
}
