package messagequeue

import (
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

const (
	minimumBufferSize       = 2
	maximumBufferSize       = 1024
	defaultBufferSize       = 20
	minimumFailedBufferSize = 1
	maximumFailedBufferSize = maximumBufferSize
	defaultFailedBufferSize = 15
)

// NewEnqueuerOptions with default values
func NewEnqueuerOptions() *EnqueuerOptions {
	return &EnqueuerOptions{
		ChunkerOptions:   NewChunkerOptions(),
		Logger:           log.Global(),
		BufferSize:       defaultBufferSize,
		FailedBufferSize: defaultFailedBufferSize,
	}
}

type EnqueuerOptions struct {
	*ChunkerOptions
	// Logger for the enqueuer to use
	Logger log.Logger
	// BufferSize to use for forming batches, must be greater than BatchSize
	BufferSize uint
	// FailedBufferSize for messages that fail to enqueue. This should be at
	// least equal to BatchSize to avoid blocking.
	FailedBufferSize uint
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
	if eo.BufferSize < minimumBufferSize || eo.BufferSize > maximumBufferSize {
		return errors.Newf("EnqueuerOptions.BufferSize(%d) was out of bounds [%d, %d]",
			eo.BufferSize, minimumBufferSize, maximumBufferSize)
	}
	if eo.FailedBufferSize < minimumFailedBufferSize || eo.FailedBufferSize > maximumFailedBufferSize {
		return errors.Newf("EnqueuerOptions.FailedBufferSize(%d) was out of bounds [%d, %d)",
			eo.FailedBufferSize, minimumFailedBufferSize, maximumFailedBufferSize)
	}
	return eo.ChunkerOptions.Validate()
}
