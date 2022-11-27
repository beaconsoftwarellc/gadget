package messagequeue

import (
	"context"
	"time"
)

// Implementation of a message queue
type Implementation interface {
	// Enqueue all the passed messages as a batch
	EnqueueBatch(context.Context, []*Message) ([]*EnqueueMessageResult, error)
	// Dequeue up to the passed count of messages waiting up to the passed
	// duration
	Dequeue(ctx context.Context, count int, wait time.Duration) ([]*Message, error)
	// Delete the passed message from the queue so that it is not processed by
	// other workers
	Delete(context.Context, *Message) error
}
