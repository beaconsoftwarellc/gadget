package messagequeue

import "time"

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
