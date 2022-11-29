package messagequeue

import (
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

const (
	defaultChunkSize                  = 10
	defaultEntryMaxWait time.Duration = 10 * time.Second
	minimumWait                       = time.Millisecond
	maximumWait                       = 24 * time.Hour
	minimumChunkSize                  = 1
	maximumChunkSize                  = 1024
)

// ChunkerOptions for configuring a Chunker instance
type ChunkerOptions struct {
	// ChunkSize determines the desired chunk size for entries to be returned
	// from the buffer.
	ChunkSize uint
	// MaxElementWait determines the maximum time an entry will wait to be chunked.
	// Smaller durations may result in chunks that are less than the desired size.
	MaxElementWait time.Duration
}

// NewChunkerOptions creates a ChunkerOptions with default values that can
// be used to create a NewChunker
func NewChunkerOptions() *ChunkerOptions {
	return &ChunkerOptions{
		ChunkSize:      defaultChunkSize,
		MaxElementWait: defaultEntryMaxWait,
	}
}

// Validate that the values contained in this ChunkerOptions are complete and
// within the bounds necessary for operation.
func (o *ChunkerOptions) Validate() error {
	if o.ChunkSize < minimumChunkSize || o.ChunkSize > maximumChunkSize {
		return errors.New("ChunkerOptions.ChunkSize(%d) was out of bounds [%d, %d]",
			o.ChunkSize, minimumChunkSize, maximumChunkSize,
		)
	}
	if o.MaxElementWait < minimumWait || o.MaxElementWait > maximumWait {
		return errors.New("ChunkerOptions.MaxElementWait(%s) was out of bounds [%s,%s]",
			o.MaxElementWait.String(), minimumWait, maximumWait)
	}
	return nil
}
