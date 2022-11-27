package messagequeue

import (
	"context"
	"time"
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
	Poll(messageQueue MessageQueue) error
	// Stop polling for messages
	Stop() error
}
