package messagequeue

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/log"
)

const stop = 1

func newWorker(handler HandleMessage, extend ExtendDeadline) *worker {
	return &worker{
		handler: handler,
		extend:  extend,
	}
}

type worker struct {
	handler HandleMessage
	extend  ExtendDeadline
	pool    chan<- *worker
	ctrl    *int32
}

func (w *worker) Run(ctrl *int32, pool chan<- *worker) {
	w.pool = pool
	w.ctrl = ctrl
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
	w.handler(ctx, message, w.extend)
	if atomic.LoadInt32(w.ctrl) != stop {
		select {
		case w.pool <- w:
		default:
			log.Warnf("could not add worker to pool as the operation would block")
		}
	}
}
