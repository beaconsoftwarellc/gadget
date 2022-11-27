package messagequeue

import (
	"context"
	"time"
)

type MessageQueue interface {
	// BatchEnqueue the passed messages in the message queue, only an error
	// affecting the entire batch will be returned, otherwise errors are
	// returned per message.
	BatchEnqueue(messages []*Message) ([]*EnqueueMessageResult, error)
	// Poll for messages processing them using the passed ProcessMessage with
	// the passed number of workers
	Poll(processMessage ProcessMessage, workers int) error
	// Stop polling for messages
	Stop()
}

// Message that can be enqueued in a MessageQueue
type Message struct {
	// ID uniquely identifies this message
	ID string
	// External field used by the sdk
	External string
	// Trace field for telemetry
	Trace string
	// Delay before this message becomes visible after being enqueued
	Delay time.Duration
	// Service this message is for
	Service string
	// Method that should be invoked to process this message
	Method string
	// Body can contain any structured (JSON, XML) or unstructured text
	// limitations are determined by the implementation
	Body string
	// Deadline for processing this message
	Deadline time.Time
}

// EnqueueMessageResult is returned on for each message that is enqueued
type EnqueueMessageResult struct {
	// Message this result is for
	*Message
	// Success indicates whether the message was successfully enqueued
	Success bool
	// SenderFault when success is false, indicates that the enqueue failed due
	// to a malformed message
	SenderFault bool
	// Error that occurred when enqueueing the message
	Error string
}

// ProcessMessage returning a boolean indicating if the message was successfully
// processed.
type ProcessMessage func(context.Context, *Message, ExtendDeadline) (bool, error)

// ExtendDeadline for the passed message
type ExtendDeadline func(context.Context, *Message, time.Duration) error

// New message queue for asynchronous processing
func New() (MessageQueue, error) {
	return nil, nil
}
