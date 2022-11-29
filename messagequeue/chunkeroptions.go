package messagequeue

import (
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// ChunkerOptions for configuring a Chunker instance
type ChunkerOptions struct {
	// Size determines the desired chunk size for entries to be returned
	// from the buffer.
	Size uint
	// ElementExpiry determines the maximum time an entry will wait to be chunked
	// smaller durations may result in chunks that are less than the desired size.
	ElementExpiry time.Duration
}

// NewChunkerOptions creates a ChunkerOptions with default values that can
// be used to create a NewChunker
func NewChunkerOptions() *ChunkerOptions {
	return &ChunkerOptions{
		Size:          defaultChunkSize,
		ElementExpiry: defaultEntryExpiry,
	}
}

// Validate that the values contained in this ChunkerOptions are complete and
// within the bounds necessary for operation.
func (o *ChunkerOptions) Validate() error {
	if o.Size < minimumChunkSize || o.Size > maximumChunkSize {
		return errors.New("ChunkerOptions.Size(%d) was out of bounds [%d, %d]",
			o.Size, minimumChunkSize, maximumChunkSize,
		)
	}
	if o.ElementExpiry < minimumExpiry || o.ElementExpiry > maximumExpiry {
		return errors.New("ChunkerOptions.ElementExpiry(%s) was out of bounds [%s,%s]",
			o.ElementExpiry.String(), minimumExpiry, maximumExpiry)
	}
	return nil
}
