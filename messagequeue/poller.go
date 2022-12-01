package messagequeue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// HandleMessage returning a boolean indicating if the message was successfully
// processed.
type HandleMessage func(context.Context, *Message, ExtendDeadline) (bool, error)

// ExtendDeadline for the passed message
type ExtendDeadline func(context.Context, *Message, time.Duration) error

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
	options       *PollerOptions
	handler       HandleMessage
	queue         MessageQueue
	pool          chan *worker
	status        atomic.Int32
	workerControl *int32
	mux           sync.Mutex
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
	if !p.status.CompareAndSwap(stateStopped, stateRunning) {
		return errors.New("Poller.Poll called on instance not in state stopped (%d)",
			stateStopped)
	}
	return nil
}

func (p *poller) Stop() error {
	p.mux.Lock()
	defer p.mux.Unlock()
	return nil
}
