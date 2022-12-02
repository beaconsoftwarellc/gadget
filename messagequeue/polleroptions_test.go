package messagequeue

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func TestNewPollerOptions(t *testing.T) {
	assert := assert1.New(t)
	actual := NewPollerOptions()
	assert.NotNil(actual)
	assert.NoError(actual.Validate())
}

func TestPollerOptions_Validate(t *testing.T) {
	assert := assert1.New(t)
	actual := NewPollerOptions()

	// no error
	assert.NoError(actual.Validate())

	// logger
	actual.Logger = nil
	assert.EqualError(actual.Validate(), "PollerOptions.Logger cannot be nil")
	actual = NewPollerOptions()

	// ConcurrentMessageHandlers
	actual.ConcurrentMessageHandlers = minimumConcurrentMessageHandlers - 1
	assert.EqualError(actual.Validate(), "PollerOptions.ConcurrentMessageHandlers(0) was out of bounds [1, -)")
	actual = NewPollerOptions()

	// WaitForBatch
	actual.WaitForBatch = minimumWaitForBatch - 1
	assert.EqualError(actual.Validate(), "PollerOptions.WaitForBatch(999.999999ms) was out of bounds [1s, 1h0m0s]")
	actual.WaitForBatch = maximumWaitForBatch + 1
	assert.EqualError(actual.Validate(), "PollerOptions.WaitForBatch(1h0m0.000000001s) was out of bounds [1s, 1h0m0s]")
	actual = NewPollerOptions()

	// TimeoutDequeueAfter
	actual.TimeoutDequeueAfter = actual.WaitForBatch - 1
	assert.EqualError(actual.Validate(), "PollerOptions.TimeoutDequeueAfter(29.999999999s) was out of bounds (WaitForBatch(30s), 1h0m0s]")
	actual.TimeoutDequeueAfter = maximumTimeoutDequeueAfter + 1
	assert.EqualError(actual.Validate(), "PollerOptions.TimeoutDequeueAfter(1h0m0.000000001s) was out of bounds (WaitForBatch(30s), 1h0m0s]")
	actual = NewPollerOptions()

	// DequeueCount
	actual.DequeueCount = minimumDequeueCount - 1
	assert.EqualError(actual.Validate(), "PollerOptions.DequeueCount(0) was out of bounds [1, 10]")
	actual.DequeueCount = maximumDequeueCount + 1
	assert.EqualError(actual.Validate(), "PollerOptions.DequeueCount(11) was out of bounds [1, 10]")
	actual = NewPollerOptions()

}
