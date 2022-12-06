package messagequeue

import (
	"strconv"
	"testing"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/stretchr/testify/assert"
)

type Entry struct {
	Name string
}

func Test_NewChunker(t *testing.T) {
	assert := assert.New(t)
	actual := NewChunker[*Entry](nil)
	assert.NotNil(actual)
	handler := func(chunk []*Entry) {}
	buffer := make(chan *Entry, 2)
	assert.NoError(actual.Start(buffer, handler))
	assert.NoError(actual.Stop())
	close(buffer)
}

func Test_Chunker_StartStop(t *testing.T) {
	assert := assert.New(t)
	expiry := time.Hour
	batchSize := 3
	buffer := make(chan *Entry, batchSize*2)
	resultChannel := make(chan []*Entry, 2)
	handler := func(chunk []*Entry) {
		resultChannel <- chunk
	}
	options := NewChunkerOptions()
	options.ChunkSize = uint(batchSize)
	options.MaxElementWait = expiry
	c := NewChunker[*Entry](options)
	// stop on not started should error
	assert.EqualError(c.Stop(), "Chunker.Run called while not in state 'Running'")

	// start the chunker
	assert.NoError(c.Start(buffer, handler))

	// start on a started chunker should error
	assert.EqualError(c.Start(buffer, handler), "Chunker.Run called while not in state 'Stopped'")

	// add batch + 1
	for i := 0; i < batchSize+1; i++ {
		buffer <- &Entry{strconv.Itoa(i)}
	}

	actual := <-resultChannel
	assert.Equal(batchSize, len(actual))
	for i := 0; i < len(actual); i++ {
		assert.Equal(strconv.Itoa(i), actual[i].Name)
	}

	// stop and make sure we get our last entry
	assert.NoError(c.Stop())
	actual = <-resultChannel
	assert.Equal(1, len(actual))
	// the batchSize+1th element is the index batchSize
	assert.Equal(strconv.Itoa(batchSize), actual[0].Name)
	close(buffer)
	close(resultChannel)
}

func Test_Chunker_Expiry(t *testing.T) {
	assert := assert.New(t)
	expiry := time.Millisecond
	batchSize := 3
	buffer := make(chan *Entry, batchSize*2)
	resultChannel := make(chan []*Entry, 2)
	handler := func(chunk []*Entry) {
		resultChannel <- chunk
	}
	options := NewChunkerOptions()
	options.ChunkSize = uint(batchSize)
	options.MaxElementWait = expiry
	c := NewChunker[*Entry](options)
	assert.NoError(c.Start(buffer, handler))

	// add a entry we want to expire
	expected := generator.String(20)
	buffer <- &Entry{expected}
	actual := <-resultChannel
	assert.Equal(1, len(actual))
	assert.Equal(expected, actual[0].Name)
	assert.NoError(c.Stop())
	close(buffer)
	close(resultChannel)
}

func Test_Chunker_BufferCloseDoesNotPanic(t *testing.T) {
	assert := assert.New(t)
	buffer := make(chan *Entry, 2)

	handler := func(chunk []*Entry) {}
	c := NewChunker[*Entry](nil)
	assert.NoError(c.Start(buffer, handler))

	close(buffer)

	assert.NoError(c.Stop())
}
