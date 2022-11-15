package messagequeue

import (
	"context"
	"time"
)

type MessageQueue interface {
	// Enqueue the passed messages in the message queue
	Enqueue(messages ...*Message) error
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
}

// ProcessMessage returning a boolean indicating if the message was successfully
// processed.
type ProcessMessage func(context.Context, *Message) (bool, error)

// New message queue for asynchronous processing
func New() (MessageQueue, error) {
	return nil, nil
}
