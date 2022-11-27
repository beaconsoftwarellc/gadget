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
func New(failHandler HandleFailedEnqueue, options *EnqueuerOptions) Enqueuer {
	return &enqueuer{
		options:       options,
		failedHandler: failHandler,
	}
}

type enqueuer struct {
	status        atomic.Uint32
	messageQueue  MessageQueue
	options       *EnqueuerOptions
	failedHandler HandleFailedEnqueue
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
		return errors.New("MessageQueue cannot be nil")
	}
	qr.mux.Lock()
	defer qr.mux.Unlock()
	if !qr.status.CompareAndSwap(statusStopped, statusRunning) {
		return errors.New("MessageQueue.Start called while not in state 'Stopped'")
	}
	qr.messageQueue = messageQueue
	qr.buffer = make(chan *Message)
	qr.failed = make(chan *EnqueueMessageResult, int(qr.options.FailedBufferSize))
	options := NewChunkerOptions()
	options.ElementExpiry = qr.options.MaxMessageWait
	options.Size = qr.options.BatchSize
	qr.chunker = NewChunker(qr.buffer, qr.sendBatch, options)
	go qr.handleFailed()
	return qr.chunker.Start()
}

func (qr *enqueuer) Stop() error {
	qr.mux.Lock()
	defer qr.mux.Unlock()
	if !qr.status.CompareAndSwap(statusRunning, statusDraining) {
		return errors.New("MessageQueue.Stop called while not in state 'Running'")
	}
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
		qr.failedHandler(qr, emr)
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
		result = make([]*EnqueueMessageResult, len(batch))
		for _, m := range batch {
			result = append(result, &EnqueueMessageResult{
				Message: m, Success: false, Error: err.Error()})
		}
	}
	for _, r := range result {
		qr.failed <- r
	}
}
