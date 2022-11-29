package messagequeue

import (
	"fmt"
	"testing"
	"time"

	assert1 "github.com/stretchr/testify/assert"
)

func TestNewEnqueuerOptions(t *testing.T) {
	assert := assert1.New(t)
	actual := NewEnqueuerOptions()
	assert.Equal(uint(defaultBufferSize), actual.BufferSize)
	assert.Equal(uint(defaultBatchSize), actual.BatchSize)
	assert.Equal(uint(defaultFailedBufferSize), actual.FailedBufferSize)
	assert.Equal(defaultMaxMessageWait, actual.MaxMessageWait)
	assert.Nil(actual.FailureHandler)
	assert.NotNil(actual.Logger)
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
	actual.BufferSize = minBufferSize - 1
	expectedError := fmt.Sprintf("EnqueuerOptions.BufferSize(%d) was out of bounds [%d, %d]",
		actual.BufferSize, minBufferSize, maxBufferSize)
	assert.EqualError(actual.Validate(), expectedError)

	actual.BufferSize = maxBufferSize + 1
	expectedError = fmt.Sprintf("EnqueuerOptions.BufferSize(%d) was out of bounds [%d, %d]",
		actual.BufferSize, minBufferSize, maxBufferSize)
	assert.EqualError(actual.Validate(), expectedError)

	// batch
	actual = NewEnqueuerOptions()
	actual.BatchSize = minBatchSize - 1
	expectedError = fmt.Sprintf("EnqueuerOptions.BatchSize(%d) was out of bounds [%d, %d)",
		actual.BatchSize, minBatchSize, actual.BufferSize)
	assert.EqualError(actual.Validate(), expectedError)

	actual.BatchSize = actual.BufferSize + 1
	expectedError = fmt.Sprintf("EnqueuerOptions.BatchSize(%d) was out of bounds [%d, %d)",
		actual.BatchSize, minBatchSize, actual.BufferSize)
	assert.EqualError(actual.Validate(), expectedError)

	// fail buffer
	actual = NewEnqueuerOptions()

	actual.FailedBufferSize = minFailedBufferSize - 1
	expectedError = fmt.Sprintf("EnqueuerOptions.FailedBufferSize(%d) was out of bounds [%d, %d)",
		actual.FailedBufferSize, minFailedBufferSize, maxFailedBufferSize)
	assert.EqualError(actual.Validate(), expectedError)

	actual.FailedBufferSize = maxFailedBufferSize + 1
	expectedError = fmt.Sprintf("EnqueuerOptions.FailedBufferSize(%d) was out of bounds [%d, %d)",
		actual.FailedBufferSize, minFailedBufferSize, maxFailedBufferSize)
	assert.EqualError(actual.Validate(), expectedError)

	// wait
	actual = NewEnqueuerOptions()
	actual.MaxMessageWait = time.Microsecond
	expectedError = fmt.Sprintf("EnqueuerOptions.MaxMessageWait(%s) was out of bounds [%s, %s)",
		actual.MaxMessageWait, minMaxMessageWait.String(), maxMaxMessageWait.String())
	assert.EqualError(actual.Validate(), expectedError)
	actual.MaxMessageWait = 24 * time.Hour
	expectedError = fmt.Sprintf("EnqueuerOptions.MaxMessageWait(%s) was out of bounds [%s, %s)",
		actual.MaxMessageWait, minMaxMessageWait.String(), maxMaxMessageWait.String())
	assert.EqualError(actual.Validate(), expectedError)
}
