package messagequeue

import "time"

const (
	defaultBufferSize       = 20
	defaultBatchSize        = 10
	defaultFailedBufferSize = 15
	defaultMaxMessageWait   = 10 * time.Second
)

// NewEnqueuerOptions with default values
func NewEnqueuerOptions() *EnqueuerOptions {
	return &EnqueuerOptions{
		BufferSize:       defaultBufferSize,
		BatchSize:        defaultBatchSize,
		FailedBufferSize: defaultFailedBufferSize,
		MaxMessageWait:   defaultMaxMessageWait,
	}
}

type EnqueuerOptions struct {
	// BufferSize to use for forming batches, must be greater than BatchSize
	BufferSize uint
	// BatchSize of messages to enqueue. Messages will be held until batch
	// size is reached or MaxMessageWait elapses before enqueueing.
	BatchSize uint
	// FailedBufferSize for messages that fail to enqueue. This should be at
	// least equal to BatchSize to avoid blocking.
	FailedBufferSize uint
	// MaxMessageWait before being enqueued. Messages wait for a full batch of
	// messages is ready prior to enqueueing. This duration prevents a message
	// from sitting too long before the batch is sent.
	MaxMessageWait time.Duration
}
