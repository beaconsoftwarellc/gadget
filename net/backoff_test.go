package net

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	assert1 "github.com/stretchr/testify/assert"
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
	assert := assert1.New(t)
	calls, f := getF(1)
	assert.NoError(Backoff(f))
	assert.Equal(1, *calls)
}

func TestBackoffExtendedErrorIsReturned(t *testing.T) {
	assert := assert1.New(t)
	calls, f := getF(10)
	start := time.Now()
	minCycle := 1 * time.Millisecond
	assert.Error(BackoffExtended(f, 5, minCycle, 100*time.Millisecond))
	assert.Equal(5, *calls)
	assert.True(time.Since(start) > minCycle)
}

func TestBackoffExtendedBadMinimum(t *testing.T) {
	assert := assert1.New(t)
	calls, f := getF(3)
	assert.Error(BackoffExtended(f, 2, 1*time.Microsecond, 10*time.Millisecond))
	assert.Equal(2, *calls)
}

func TestCalculateBackoff(t *testing.T) {
	assert := assert1.New(t)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := CalculateBackoff(r, 1, time.Minute, time.Hour, time.Minute)
	assert.GreaterOrEqual(5*time.Minute, result)
	result = CalculateBackoff(r, 5, time.Minute, time.Hour, time.Minute)
	assert.GreaterOrEqual(40*time.Minute, result)
	result = CalculateBackoff(r, 10, time.Minute, time.Hour, time.Minute)
	assert.GreaterOrEqual(time.Hour, result)
}

func TestCalculateBackoffSecond(t *testing.T) {
	assert := assert1.New(t)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result1 := CalculateBackoff(r, 1, 1*time.Second, 10*time.Minute,
		time.Second)
	assert.GreaterOrEqual(2*time.Second, result1)
	result2 := CalculateBackoff(r, 2, 1*time.Second, 10*time.Minute,
		time.Second)
	assert.GreaterOrEqual(5*time.Second, result2)
	result3 := CalculateBackoff(r, 3, 1*time.Second, 10*time.Minute,
		time.Second)
	assert.GreaterOrEqual(10*time.Second, result3)
	result4 := CalculateBackoff(r, 4, 1*time.Second, 10*time.Minute,
		time.Second)
	assert.GreaterOrEqual(20*time.Second, result4)
}
