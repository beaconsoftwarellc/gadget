package messagequeue

import (
	"fmt"
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func TestNewEnqueuerOptions(t *testing.T) {
	assert := assert1.New(t)
	actual := NewEnqueuerOptions()
	assert.NotNil(actual)
	assert.NoError(actual.Validate())
}

func TestEnqueuerOptions_Validate(t *testing.T) {
	assert := assert1.New(t)
	actual := NewEnqueuerOptions()
	assert.NoError(actual.Validate())

	// logger
	actual = NewEnqueuerOptions()
	actual.Logger = nil
	assert.EqualError(actual.Validate(), "EnqueuerOptions.Logger cannot be nil")

	// buffer
	actual = NewEnqueuerOptions()
	actual.BufferSize = minimumBufferSize - 1
	expectedError := fmt.Sprintf("EnqueuerOptions.BufferSize(%d) was out of bounds [%d, %d]",
		actual.BufferSize, minimumBufferSize, maximumBufferSize)
	assert.EqualError(actual.Validate(), expectedError)

	actual.BufferSize = maximumBufferSize + 1
	expectedError = fmt.Sprintf("EnqueuerOptions.BufferSize(%d) was out of bounds [%d, %d]",
		actual.BufferSize, minimumBufferSize, maximumBufferSize)
	assert.EqualError(actual.Validate(), expectedError)

	// size is validated
	actual = NewEnqueuerOptions()
	actual.ChunkSize = minimumChunkSize - 1
	expected := fmt.Sprintf("ChunkerOptions.ChunkSize(%d) was out of bounds [%d, %d]",
		actual.ChunkSize, minimumChunkSize, maximumChunkSize)
	assert.EqualError(actual.Validate(), expected)

	actual.ChunkSize = maximumChunkSize + 1
	expected = fmt.Sprintf("ChunkerOptions.ChunkSize(%d) was out of bounds [%d, %d]",
		actual.ChunkSize, minimumChunkSize, maximumChunkSize)
	assert.EqualError(actual.Validate(), expected)

	// element expiry is validated
	actual = NewEnqueuerOptions()
	actual.MaxElementWait = minimumWait - 1
	expected = fmt.Sprintf("ChunkerOptions.MaxElementWait(%s) was out of bounds [%s,%s]",
		actual.MaxElementWait.String(), minimumWait, maximumWait)
	assert.EqualError(actual.Validate(), expected)

	actual.MaxElementWait = maximumWait + 2
	expected = fmt.Sprintf("ChunkerOptions.MaxElementWait(%s) was out of bounds [%s,%s]",
		actual.MaxElementWait.String(), minimumWait, maximumWait)
	assert.EqualError(actual.Validate(), expected)

	// fail buffer
	actual = NewEnqueuerOptions()
	actual.FailedBufferSize = minimumFailedBufferSize - 1
	expectedError = fmt.Sprintf("EnqueuerOptions.FailedBufferSize(%d) was out of bounds [%d, %d)",
		actual.FailedBufferSize, minimumFailedBufferSize, maximumFailedBufferSize)
	assert.EqualError(actual.Validate(), expectedError)

	actual.FailedBufferSize = maximumFailedBufferSize + 1
	expectedError = fmt.Sprintf("EnqueuerOptions.FailedBufferSize(%d) was out of bounds [%d, %d)",
		actual.FailedBufferSize, minimumFailedBufferSize, maximumFailedBufferSize)
	assert.EqualError(actual.Validate(), expectedError)
}
