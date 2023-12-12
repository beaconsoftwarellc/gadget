package net

import (
	"math"
	"math/rand"
	"time"
)

const (
	// DefaultMaxRetries is used as the maximum retries in the backoff when Backoff is called.
	DefaultMaxRetries = 5
	// DefaultMinimumCycle is used as the minimum cycle duration when Backoff is called.
	DefaultMinimumCycle = 100 * time.Millisecond
	// DefaultCeiling is used as the maximum wait time between executions when Backoff is called.
	DefaultCeiling = 15 * time.Second
)

// F is a function that will be called sequentially
type F func() error

// Backoff Executes the passed function with exponential back off up to a maximum a number of tries with the
// minimum amount of time per cycle using default values.
func Backoff(f F) error {
	return BackoffExtended(f, DefaultMaxRetries, DefaultMinimumCycle, DefaultCeiling)
}

// BackoffExtended Executes the passed function with exponential back off up to a maximum a number of tries with the
// minimum amount of time per cycle.
func BackoffExtended(f F, maxTries int, minimumCycle time.Duration, maxCycle time.Duration) error {
	var err error
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for n := 1; n < maxTries+1; n++ {
		err = f()
		if nil == err {
			break
		}
		wait := CalculateBackoff(r, n, minimumCycle, maxCycle)
		time.Sleep(wait)
	}
	return err
}

// CalculateBackoff returns a duration for exponential backoff
func CalculateBackoff(r *rand.Rand, attempt int, minimumCycle time.Duration, maxCycle time.Duration) time.Duration {
	min := uint(minimumCycle.Seconds())
	// if min cycle is 1 second or smaller bring it over 1 to make the exponentiation work
	if min <= 1 {
		min = 2
	}
	max := float64(maxCycle.Seconds())
	full := math.Min(math.Pow(float64(min), float64(attempt)), max)
	// exponentiation will be 90% of our calculated value with maxCycle as a ceiling
	exp := full * 0.9
	// r.Float64() is in [0.0, 1.0)
	jitter := r.Float64() * float64(full*0.1)
	// exponentiation plus jitter
	return time.Duration(exp+jitter) * time.Second
}
