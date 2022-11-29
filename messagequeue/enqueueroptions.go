package messagequeue

import (
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

const (
	minBufferSize           = 2
	maxBufferSize           = 1024
	minFailedBufferSize     = 1
	maxFailedBufferSize     = maxBufferSize
	defaultBufferSize       = 20
	minBatchSize            = 1
	defaultBatchSize        = 10
	defaultFailedBufferSize = 15
	minMaxMessageWait       = time.Millisecond
	maxMaxMessageWait       = time.Hour
	defaultMaxMessageWait   = 10 * time.Second
)

// NewEnqueuerOptions with default values
func NewEnqueuerOptions() *EnqueuerOptions {
	return &EnqueuerOptions{
		Logger:           log.Global(),
		BufferSize:       defaultBufferSize,
		BatchSize:        defaultBatchSize,
		FailedBufferSize: defaultFailedBufferSize,
		MaxMessageWait:   defaultMaxMessageWait,
	}
}

type EnqueuerOptions struct {
	// Logger for the enqueuer to use
	Logger log.Logger
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
	// FailureHandler receives messages that failed to enqueue, optional.
	FailureHandler HandleFailedEnqueue
}

// Validate that the values contained in this Options are complete and within the
// bounds necessary for operation.
func (eo *EnqueuerOptions) Validate() error {
	// logger should not be nil
	if eo.Logger == nil {
		return errors.New("EnqueuerOptions.Logger cannot be nil")
	}
	// buffer size must be within sensible bounds
	if eo.BufferSize < minBufferSize || eo.BufferSize > maxBufferSize {
		return errors.New("EnqueuerOptions.BufferSize(%d) was out of bounds [%d, %d]",
			eo.BufferSize, minBufferSize, maxBufferSize)
	}
	// batch size must be more than 0 and less than BatchSize
	if eo.BatchSize < minBatchSize || eo.BatchSize >= eo.BufferSize {
		return errors.New("EnqueuerOptions.BatchSize(%d) was out of bounds [%d, %d)",
			eo.BatchSize, minBatchSize, eo.BufferSize)
	}
	if eo.MaxMessageWait < minMaxMessageWait || eo.MaxMessageWait > maxMaxMessageWait {
		return errors.New("EnqueuerOptions.MaxMessageWait(%s) was out of bounds [%s, %s)",
			eo.MaxMessageWait, minMaxMessageWait.String(), maxMaxMessageWait.String())
	}
	if eo.FailedBufferSize < minFailedBufferSize || eo.FailedBufferSize > maxFailedBufferSize {
		return errors.New("EnqueuerOptions.FailedBufferSize(%d) was out of bounds [%d, %d)",
			eo.FailedBufferSize, minFailedBufferSize, maxFailedBufferSize)
	}
	return nil
}
