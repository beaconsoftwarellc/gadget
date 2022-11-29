package messagequeue

import (
	"context"
	"time"
)

// MessageQueue for processing tasks asynchronously. All tasks must be reentrant.
type MessageQueue interface {
	// BlockingEnqueue will enqueue the passed message in this message queue,
	// blocking the calling process until the operation completes.
	BlockingEnqueue(messages *Message) (*EnqueueMessageResult, error)
	// Enqueue the passed message in this message queue. This operation may be
	// buffered and errors enqueueing will not be available for immediate handling.
	Enqueue(messages *Message) error
	// Poll for messages, processing them using the passed ProcessMessage with
	// the passed number of workers
	Poll(processMessage ProcessMessage, workers int) error
	// Stop all active workers, drain queues, and free resources.
	Stop() error
}

// ProcessMessage returning a boolean indicating if the message was successfully
// processed.
type ProcessMessage func(context.Context, *Message, ExtendDeadline) (bool, error)

// HandleFailedEnqueue is called when a message fails to enqueue
type HandleFailedEnqueue func(context.Context, MessageQueue, *EnqueueMessageResult)

// ExtendDeadline for the passed message
type ExtendDeadline func(context.Context, *Message, time.Duration) error
