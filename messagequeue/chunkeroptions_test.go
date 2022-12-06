package messagequeue

import (
	"fmt"
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func TestNewChunkerOptions(t *testing.T) {
	assert := assert1.New(t)
	options := NewChunkerOptions()
	assert.NotNil(options)
	assert.NoError(options.Validate())
}

func TestChunkerOptions_Valid(t *testing.T) {
	assert := assert1.New(t)

	options := NewChunkerOptions()

	// default should be valid
	assert.NoError(options.Validate())

	// size is validated
	options.ChunkSize = minimumChunkSize - 1
	expected := fmt.Sprintf("ChunkerOptions.ChunkSize(%d) was out of bounds [%d, %d]",
		options.ChunkSize, minimumChunkSize, maximumChunkSize)
	assert.EqualError(options.Validate(), expected)

	options.ChunkSize = maximumChunkSize + 1
	expected = fmt.Sprintf("ChunkerOptions.ChunkSize(%d) was out of bounds [%d, %d]",
		options.ChunkSize, minimumChunkSize, maximumChunkSize)
	assert.EqualError(options.Validate(), expected)

	options.ChunkSize = defaultChunkSize
	// element expiry is validated
	options.MaxElementWait = minimumWait - 1
	expected = fmt.Sprintf("ChunkerOptions.MaxElementWait(%s) was out of bounds [%s,%s]",
		options.MaxElementWait.String(), minimumWait, maximumWait)
	assert.EqualError(options.Validate(), expected)

	options.MaxElementWait = maximumWait + 2
	expected = fmt.Sprintf("ChunkerOptions.MaxElementWait(%s) was out of bounds [%s,%s]",
		options.MaxElementWait.String(), minimumWait, maximumWait)
	assert.EqualError(options.Validate(), expected)
}
