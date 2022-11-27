package messagequeue

import (
	"context"
	"time"
)

// MessageQueue for enqueueing and dequeueing messages
type MessageQueue interface {
	// Enqueue all the passed messages as a batch
	EnqueueBatch(context.Context, []*Message) ([]*EnqueueMessageResult, error)
	// Dequeue up to the passed count of messages waiting up to the passed
	// duration
	Dequeue(ctx context.Context, count int, wait time.Duration) ([]*Message, error)
	// Delete the passed message from the queue so that it is not processed by
	// other workers
	// TODO: [COR-553] Batch delete messages
	Delete(context.Context, *Message) error
}
