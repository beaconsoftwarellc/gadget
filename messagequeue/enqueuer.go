package messagequeue

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

const (
	statusStopped  uint32 = 0
	statusRunning  uint32 = 1
	statusDraining uint32 = 2
)

// HandleFailedEnqueue is called when a message fails to enqueue
type HandleFailedEnqueue func(Enqueuer, *EnqueueMessageResult)

// Enqueuer for inserting tasks into a MessageQueue
type Enqueuer interface {
	// Start validates configurations and creates the resources necessary to
	// handle enqueuing messages.
	Start(messageQueue MessageQueue) error
	// Enqueue the passed message. This operation may be
	// buffered and errors enqueueing will not be available for immediate handling.
	Enqueue(messages *Message) error
	// Stop all active workers, drain queues, and free resources.
	Stop() error
}

// New enqueuer for inserting messages into a MessageQueue
func New(options *EnqueuerOptions) Enqueuer {
	if nil == options {
		options = NewEnqueuerOptions()
	}
	return &enqueuer{
		options: options,
	}
}

type enqueuer struct {
	status       atomic.Uint32
	messageQueue MessageQueue
	options      *EnqueuerOptions
	// buffer used by the chunker
	buffer chan *Message
	// buffer where we send messages that fail to enqueue that will
	// be fed to the failedHandler
	failed chan *EnqueueMessageResult
	// chunker has the logic for batching our enqueue calls
	chunker Chunker[*Message]
	mux     sync.Mutex
}

func (qr *enqueuer) Start(messageQueue MessageQueue) error {
	if nil == messageQueue {
		return errors.New("messageQueue cannot be nil")
	}
	qr.mux.Lock()
	var err error
	if !qr.status.CompareAndSwap(statusStopped, statusRunning) {
		qr.mux.Unlock()
		return errors.New("Enqueuer.Start called while not in state 'Stopped'")
	}
	if err = qr.options.Validate(); nil != err {
		qr.status.Store(statusStopped)
		qr.mux.Unlock()
		return err
	}
	qr.messageQueue = messageQueue
	qr.buffer = make(chan *Message)
	options := NewChunkerOptions()
	options.ElementExpiry = qr.options.MaxMessageWait
	options.Size = qr.options.BatchSize
	qr.chunker = NewChunker(qr.buffer, qr.sendBatch, options)
	err = qr.chunker.Start()
	qr.failed = make(chan *EnqueueMessageResult, int(qr.options.FailedBufferSize))
	go qr.handleFailed()
	qr.mux.Unlock()
	if nil != err {
		qr.Stop()
	}
	// set the failure handler to just log if one is not defined
	if nil == qr.options.FailureHandler {
		qr.options.FailureHandler = qr.logFailed
	}
	return err
}

func (qr *enqueuer) Stop() error {
	qr.mux.Lock()
	if !qr.status.CompareAndSwap(statusRunning, statusStopped) {
		qr.mux.Unlock()
		return errors.New("Enqueuer.Stop called while not in state 'Running'")
	}
	defer qr.mux.Unlock()
	qr.chunker.Stop()
	close(qr.buffer)
	close(qr.failed)
	qr.messageQueue = nil
	return nil
}

func (qr *enqueuer) Running() bool {
	return qr.status.Load() == statusRunning
}

func (qr *enqueuer) Enqueue(message *Message) error {
	if !qr.Running() {
		return errors.New("cannot enqueue into a stopped MessageQueue")
	}
	// this will block if the buffer is full
	qr.buffer <- message
	return nil
}

func (qr *enqueuer) logFailed(nqr Enqueuer, result *EnqueueMessageResult) {
	qr.options.Logger.Errorf("failed to enqueue message(%s, %s): %s",
		result.Service, result.Method, result.Error)
}

func (qr *enqueuer) handleFailed() {
	var (
		emr *EnqueueMessageResult
		ok  bool
	)
	for qr.Running() {
		emr, ok = <-qr.failed
		if !ok {
			return
		}
		qr.options.FailureHandler(qr, emr)
	}
}

func (qr *enqueuer) sendBatch(batch []*Message) {
	if len(batch) == 0 {
		return
	}
	var (
		err    error
		result []*EnqueueMessageResult
	)
	result, err = qr.messageQueue.EnqueueBatch(context.Background(), batch)
	if nil != err {
		// the whole batch failed so call the handler
		// with all emr's for all the messages
		result = make([]*EnqueueMessageResult, 0, len(batch))
		for _, m := range batch {
			result = append(result, &EnqueueMessageResult{
				Message: m, Success: false, Error: err.Error()})
		}
	}
	for _, r := range result {
		if !r.Success {
			qr.failed <- r
		}
	}
}
