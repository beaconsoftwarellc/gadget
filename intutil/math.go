package intutil

import (
	"sync/atomic"

	"golang.org/x/exp/constraints"
)

// Constants
const (
	MaxUint = ^uint(0)
	MinUint = 0
	MaxInt  = int(MaxUint >> 1)
	MinInt  = -MaxInt - 1
)

// Abs value of an int
func Abs[T constraints.Signed](a T) T {
	if a < 0 {
		return -a
	}

	return a
}

// Min returns the smaller of a or b
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Minv returns the smallest int passed in the variadic argument.
func Minv[T constraints.Ordered](ints ...T) T {
	var min T
	if len(ints) == 0 {
		return min
	}

	min = ints[0]

	for _, i := range ints {
		min = Min(min, i)
	}

	return min
}

// Max returns the larger of a or b
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Maxv returns the largest int passed in the variadic argument.
func Maxv[T constraints.Ordered](ints ...T) T {
	var min T
	if len(ints) == 0 {
		return min
	}

	min = ints[0]

	for _, i := range ints {
		min = Max(min, i)
	}

	return min
}

// Decrementor defines a class that allows multiple threads/processes to act upon to decrement a counter
type Decrementor struct {
	initialMax int64
	max        *int64
}

// NewDecrementor will return a reference to a Decrementor object
func NewDecrementor(initialMax int64) *Decrementor {
	return &Decrementor{initialMax: initialMax, max: &initialMax}
}

// Decrement will decrement by 1 the initial max the class was instantiated with
func (decrementor *Decrementor) Decrement() int {
	// constant priority makes it a lifo whereas a decreasing priority will
	// give us a fifo
	p := atomic.LoadInt64(decrementor.max)
	// this will probably never happen in production, but just to be safe
	if p == 0 {
		atomic.StoreInt64(decrementor.max, decrementor.initialMax)
	}
	return int(atomic.AddInt64(decrementor.max, -1))
}

// GetInitialMax will return the inital maximum value that was set upon instantiation
func (decrementor *Decrementor) GetInitialMax() int {
	return int(decrementor.initialMax)
}

// Clamp limits value to a given minimum and maximum
func Clamp[T constraints.Ordered](value, min, max T) T {
	return Min(Max(value, min), max)
}

// Anonymize converts an array of ints to an array of anonymous interfaces
func Anonymize(arr []int) []interface{} {
	ia := make([]interface{}, len(arr))
	for i, s := range arr {
		ia[i] = s
	}
	return ia
}
