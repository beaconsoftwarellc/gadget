package messagequeue

import (
	"sync"
	"sync/atomic"

	"github.com/beaconsoftwarellc/gadget/v2/errors"

	"github.com/beaconsoftwarellc/gadget/v2/timeutil"
)

// Chunker interface for 'chunking' entries on a buffer into slices of a desired
// size.
type Chunker[T any] interface {
	// Start the chunker
	Start(buffer <-chan T, handler Handler[T]) error
	// Stop the chunker and clean up any resources
	Stop() error
}

type chunker[T any] struct {
	options *ChunkerOptions
	buffer  <-chan T
	// we need a control channel since we are not controlling the
	// buffer, at least two so that a premature buffer close does
	// not lock us on stop
	control chan bool
	handler Handler[T]
	state   *atomic.Uint32
	wait    sync.WaitGroup
	mux     sync.Mutex
}

// Handler receives chunks that are created by the chunker
type Handler[T any] func(chunk []T)

// NewChunker with the passed buffer, handler and options. Chunker will create
// slices of a specified size from the passed buffer and pass them to the handler.
// Entries retrieved from the buffer are guaranteed to be delivered to
// the handler within the configured ChunkElementExpiry duration.
func NewChunker[T any](options *ChunkerOptions) Chunker[T] {
	if nil == options {
		options = NewChunkerOptions()
	}
	return &chunker[T]{
		options: options,
		state:   &atomic.Uint32{},
	}
}

func (c *chunker[T]) Start(buffer <-chan T, handler Handler[T]) error {
	if nil == buffer {
		return errors.New("buffer cannot be nil")
	}
	if nil == handler {
		return errors.New("handler cannot be nil")
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	if !c.state.CompareAndSwap(statusStopped, statusRunning) {
		return errors.New("Chunker.Run called while not in state 'Stopped'")
	}
	// validate our options
	if err := c.options.Validate(); err != nil {
		return err
	}
	c.buffer = buffer
	c.handler = handler
	c.control = make(chan bool)
	go c.chunk()
	return nil
}

func (c *chunker[T]) Stop() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if !c.state.CompareAndSwap(statusRunning, statusStopped) {
		return errors.New("Chunker.Run called while not in state 'Running'")
	}
	// we only have one worker so we can just close the control channel
	close(c.control)
	c.wait.Wait()
	// release references to the buffer and handler so they can be gc'd
	c.buffer = nil
	c.handler = nil
	return nil
}

func (c *chunker[T]) chunk() {
	var (
		staleBatch = false
		stop       = false
		batch      = make([]T, 0, c.options.ChunkSize)
		ticker     = timeutil.NewTicker(c.options.MaxElementWait).Start()
	)
	c.wait.Add(1)
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
			len(batch) >= int(c.options.ChunkSize)) {
			c.handler(batch)
			batch = make([]T, 0, c.options.ChunkSize)
		}
		if stop {
			c.wait.Done()
			return
		}
	}
}
