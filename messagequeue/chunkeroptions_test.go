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
	options.Size = minimumChunkSize - 1
	expected := fmt.Sprintf("ChunkerOptions.Size(%d) was out of bounds [%d, %d]",
		options.Size, minimumChunkSize, maximumChunkSize)
	assert.EqualError(options.Validate(), expected)

	options.Size = maximumChunkSize + 1
	expected = fmt.Sprintf("ChunkerOptions.Size(%d) was out of bounds [%d, %d]",
		options.Size, minimumChunkSize, maximumChunkSize)
	assert.EqualError(options.Validate(), expected)

	options.Size = defaultChunkSize
	// element expiry is validated
	options.ElementExpiry = minimumExpiry - 1
	expected = fmt.Sprintf("ChunkerOptions.ElementExpiry(%s) was out of bounds [%s,%s]",
		options.ElementExpiry.String(), minimumExpiry, maximumExpiry)
	assert.EqualError(options.Validate(), expected)

	options.ElementExpiry = maximumExpiry + 2
	expected = fmt.Sprintf("ChunkerOptions.ElementExpiry(%s) was out of bounds [%s,%s]",
		options.ElementExpiry.String(), minimumExpiry, maximumExpiry)
	assert.EqualError(options.Validate(), expected)
}
