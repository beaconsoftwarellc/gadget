package messagequeue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"

	"github.com/beaconsoftwarellc/gadget/v2/timeutil"
)

const (
	stateStopped                     = 0
	stateRunning                     = 1
	defaultChunkSize                 = 10
	defaultEntryExpiry time.Duration = 10 * time.Second
	minimumExpiry                    = time.Millisecond
	maximumExpiry                    = 24 * time.Hour
	minimumChunkSize                 = 1
	maximumChunkSize                 = 1024
)

// Chunker interface for 'chunking' entries on a buffer into slices of a desired
// size.
type Chunker[T any] interface {
	// Start the chunker
	Start() error
	// Stop the chunker and clean up any resources
	Stop() error
}

// ChunkerOptions for configuring a Chunker instance
type ChunkerOptions struct {
	// Size determines the desired chunk size for entries to be returned
	// from the buffer.
	Size uint
	// ElementExpiry determines the maximum time an entry will wait to be chunked
	// smaller durations may result in chunks that are less than the desired size.
	ElementExpiry time.Duration
}

// Valid returns an error if an options value is out of bounds.
func (o *ChunkerOptions) Valid() error {
	if o.Size < minimumChunkSize || o.Size > maximumChunkSize {
		return errors.New("ChunkerOptions.Size(%d) was out of bounds [%d, %d]",
			o.Size, minimumChunkSize, maximumChunkSize,
		)
	}
	if o.ElementExpiry < minimumExpiry || o.ElementExpiry > maximumExpiry {
		return errors.New("ChunkerOptions.ElementExpiry(%s) was out of bounds [%s,%s]",
			minimumExpiry.String(), minimumExpiry, maximumExpiry)
	}
	return nil
}

// NewChunkerOptions creates a ChunkerOptions with default values that can
// be used to create a NewChunker
func NewChunkerOptions() *ChunkerOptions {
	return &ChunkerOptions{
		Size:          defaultChunkSize,
		ElementExpiry: defaultEntryExpiry,
	}
}

type chunker[T any] struct {
	options *ChunkerOptions
	buffer  <-chan T
	control chan bool
	handler Handler[T]
	state   *atomic.Uint32
	mux     sync.Mutex
}

// Handler receives chunks that are created by the chunker
type Handler[T any] func(chunk []T)

// NewChunker with the passed buffer, handler and options. Chunker will create
// slices of a specified size from the passed buffer and pass them to the handler.
// Entries retrieved from the buffer are guaranteed to be delivered to
// the handler within the configured ChunkElementExpiry duration.
func NewChunker[T any](buffer <-chan T, handler Handler[T],
	options *ChunkerOptions) Chunker[T] {
	if nil == options {
		options = NewChunkerOptions()
	}
	return &chunker[T]{
		options: options,
		buffer:  buffer,
		// we need a control channel since we are not controlling the
		// buffer, at least two so that a premature buffer close does
		// not lock us on stop
		control: make(chan bool),
		handler: handler,
		state:   &atomic.Uint32{},
	}
}

func (c *chunker[T]) Start() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if !c.state.CompareAndSwap(stateStopped, stateRunning) {
		return errors.New("Chunker.Run called while not in state 'Stopped'")
	}
	// validate our options
	if err := c.options.Valid(); err != nil {
		return err
	}
	c.control = make(chan bool)
	go c.chunk()
	return nil
}

func (c *chunker[T]) Stop() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if !c.state.CompareAndSwap(stateRunning, stateStopped) {
		return errors.New("Chunker.Run called while not in state 'Running'")
	}
	// we only have one worker so we can just close the control channel
	close(c.control)
	return nil
}

func (c *chunker[T]) chunk() {
	var (
		staleBatch = false
		stop       = false
		batch      = make([]T, 0, c.options.Size)
		ticker     = timeutil.NewTicker(c.options.ElementExpiry).Start()
	)
	defer ticker.Stop()
	for !stop {
		select {
		case message, ok := <-c.buffer:
			if !ok {
				stop = true
				break
			}
			batch = append(batch, message)
			// we want to flush the buffer when the oldest entry is
			// c.options.ChunkElementExpiry old, so we should start
			// counting at len == 1
			if len(batch) == 1 {
				ticker.Reset()
			}
		case <-ticker.Channel():
			staleBatch = true
		case <-c.control:
			stop = true
		}
		// send if there is at least one entry and:
		// 1. a message in the batch is stale
		// 2. stop was called
		// 3. we are at our batch size
		if len(batch) > 0 && (staleBatch || stop ||
			len(batch) >= int(c.options.Size)) {
			c.handler(batch)
			batch = make([]T, 0, c.options.Size)
		}
		if stop {
			return
		}
	}
}
