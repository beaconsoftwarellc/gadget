package net

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getF(expectedCalls int) (*int, F) {
	i := 0
	return &i, func() error {
		var err error
		if (i + 1) < expectedCalls {
			err = fmt.Errorf("%s %d", "error", i)
		}
		i++
		return err
	}
}

func TestBackoff(t *testing.T) {
	assert := assert.New(t)
	calls, f := getF(1)
	assert.NoError(Backoff(f))
	assert.Equal(1, *calls)
}

func TestBackoffExtendedErrorIsReturned(t *testing.T) {
	assert := assert.New(t)
	calls, f := getF(10)
	start := time.Now()
	minCycle := 1 * time.Millisecond
	assert.Error(BackoffExtended(f, 5, minCycle, 100*time.Millisecond))
	assert.Equal(5, *calls)
	assert.True(time.Since(start) > minCycle)
}

func TestBackoffExtendedBadMinimum(t *testing.T) {
	assert := assert.New(t)
	calls, f := getF(3)
	assert.Error(BackoffExtended(f, 2, 1*time.Microsecond, 10*time.Millisecond))
	assert.Equal(2, *calls)
}

func TestCalculateBackoff(t *testing.T) {
	assert := assert.New(t)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result1 := CalculateBackoff(r, 5, 10*time.Millisecond, 6*time.Hour)
	result2 := CalculateBackoff(r, 6, 10*time.Millisecond, 6*time.Hour)
	result3 := CalculateBackoff(r, 10, 10*time.Millisecond, 6*time.Hour)
	assert.True(result1 > (10*time.Second), "Expected result to be greater than 10 seconds, was: %v", result1)
	assert.True(result2 > result1, "Expected result2 to be greater than result1, was: %v", result2)
	assert.True(result3 < 6*time.Hour, "Expected result3 to be less than 6 hours, was: %v", result3)
}
