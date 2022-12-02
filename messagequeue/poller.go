package messagequeue

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// HandleMessage returning a boolean indicating if the message was successfully
// processed. If this function returns true the message will be deleted from the
// queue, otherwise it will become available for other handlers after it's
// visibility timeout expires.
type HandleMessage func(context.Context, *Message) bool

// Poller retrieves batches of messages from a message queue and handles them
// using provided functions.
type Poller interface {
	// Poll for messages on the passed queue
	Poll(HandleMessage, MessageQueue) error
	// Stop polling for messages
	Stop() error
}

func NewPoller(options *PollerOptions) Poller {
	return &poller{
		options: options,
	}
}

type poller struct {
	options *PollerOptions
	queue   MessageQueue
	pool    chan *worker
	status  *atomic.Uint32
	workers sync.WaitGroup
	mux     sync.Mutex
}

func (p *poller) Poll(handler HandleMessage, messageQueue MessageQueue) error {
	if nil == handler {
		return errors.New("handler cannot be nil")
	}
	if nil == messageQueue {
		return errors.New("messageQueue cannot be nil")
	}
	p.mux.Lock()
	defer p.mux.Unlock()
	if !p.status.CompareAndSwap(statusStopped, statusRunning) {
		return errors.New("Poller.Poll called on instance not in state stopped (%d)",
			statusStopped)
	}
	p.queue = messageQueue
	p.pool = make(chan *worker, p.options.ConcurrentMessageHandlers)
	for i := 0; i < p.options.ConcurrentMessageHandlers; i++ {
		newWorker(handler).Run(&p.workers, p.status, p.pool)

	}
	go p.poll()
	return nil
}

func (p *poller) poll() {
	var (
		ctx      = context.Background()
		cancel   context.CancelFunc
		messages []*Message
		err      error
	)
	for p.status.Load() == statusRunning {
		if p.options.TimeoutDequeueAfter > 0 {
			ctx, cancel = context.WithTimeout(ctx, p.options.TimeoutDequeueAfter)
			defer cancel()
		}
		messages, err = p.queue.Dequeue(ctx, p.options.DequeueCount,
			p.options.WaitForBatch)
		if nil != err {
			p.options.Logger.Error(err)
		} else {
			p.handleMessages(messages)
		}
	}
}

func (p *poller) handleMessages(messages []*Message) {
	for _, message := range messages {
		worker, ok := <-p.pool
		if !ok {
			// this is early termination. The channel was closed so just exit.
			return
		}
		go worker.HandleMessage(message)
	}
}

func (p *poller) Stop() error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if !p.status.CompareAndSwap(statusRunning, statusDraining) {
		return errors.New("Poller.Stop called on instance not in state running (%d)",
			statusRunning)
	}
	p.workers.Wait()
	close(p.pool)
	p.status.Store(statusStopped)
	return nil
}
