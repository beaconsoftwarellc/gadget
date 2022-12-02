package messagequeue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

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
	if nil == options {
		options = NewPollerOptions()
	}
	return &poller{
		options: options,
	}
}

type messageJob struct {
	p       *poller
	message *Message
}

func (mw *messageJob) Work() bool {
	return mw.p.handle(mw.message)
}

type poller struct {
	options *PollerOptions
	queue   MessageQueue
	handler HandleMessage
	pool    chan *Worker
	status  atomic.Uint32
	cancel  context.CancelFunc
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
	p.handler = handler
	// this is just so we don't panic if stop is called before the first poll
	p.cancel = func() {}
	p.queue = messageQueue
	p.pool = make(chan *Worker, p.options.ConcurrentMessageHandlers)
	for i := 0; i < p.options.ConcurrentMessageHandlers; i++ {
		AddWorker(&p.workers, p.pool)
	}
	go p.poll()
	return nil
}

func (p *poller) poll() {
	var (
		ctx      context.Context
		messages []*Message
		err      error
	)
	for p.status.Load() == statusRunning {
		ctx, p.cancel = context.WithTimeout(context.Background(),
			p.options.QueueOperationTimeout+p.options.WaitForBatch)
		defer p.cancel()
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
		worker.Add(&messageJob{p: p, message: message})
	}
}

func (p *poller) handle(message *Message) bool {
	var (
		ctx    = context.Background()
		cancel context.CancelFunc
	)
	if message.Deadline.After(time.Now()) {
		ctx, cancel = context.WithDeadline(ctx, message.Deadline)
		defer cancel()
	}
	if p.handler(ctx, message) {
		p.delete(message)
	}
	return p.status.Load() == statusRunning
}

func (p *poller) delete(message *Message) {
	ctx, cancel := context.WithTimeout(context.Background(),
		p.options.QueueOperationTimeout)
	defer cancel()
	p.queue.Delete(ctx, message)
}

func (p *poller) drain() {
	for i := 0; i < p.options.ConcurrentMessageHandlers; i++ {
		w := <-p.pool
		w.Exit()
	}
	p.workers.Wait()
}

func (p *poller) Stop() error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if !p.status.CompareAndSwap(statusRunning, statusDraining) {
		return errors.New("Poller.Stop called on instance not in state running (%d)",
			statusRunning)
	}
	p.cancel()
	// wait for any workers to exit, this will take up to Message.VisibilityTimeout
	// assuming one was provided and the Message Handlers are well behaved.
	// we could time this out as well and throw an error.
	p.drain()
	close(p.pool)
	p.status.Store(statusStopped)
	p.handler = nil
	p.queue = nil
	return nil
}
