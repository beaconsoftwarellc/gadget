package messagequeue

import "time"

type MessageQueue interface {
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

type messagequeue struct {
}

// New message queue for asynchronous processing
func New() (MessageQueue, error) {
	return &messagequeue{}, nil
}
