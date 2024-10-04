package intutil

import (
	"math"
	"sync/atomic"

	"golang.org/x/exp/constraints"
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

func getMinMax[T constraints.Integer](i T) (int64, uint64) {
	switch any(i).(type) {
	case int:
		return math.MinInt, math.MaxInt
	case int8:
		return math.MinInt8, math.MaxInt8
	case int16:
		return math.MinInt16, math.MaxInt16
	case int32:
		return math.MinInt32, math.MaxInt32
	case int64:
		return math.MinInt64, math.MaxInt64
	case uint:
		return 0, math.MaxUint
	case uint8:
		return 0, math.MaxUint8
	case uint16:
		return 0, math.MaxUint16
	case uint32:
		return 0, math.MaxUint32
	case uint64:
		return 0, math.MaxUint64
	default:
		return 0, 0
	}
}

// Clamp limits value to a given minimum and maximum
func Clamp[T constraints.Integer](value, a, b T) T {
	var (
		lower, upper = a, b
	)
	if lower > upper {
		lower, upper = upper, lower
	}
	return Min(Max(value, lower), upper)
}

// ClampCast will clamp a value to the range of V
func ClampCast[T, V constraints.Integer](value T) V {
	var (
		min, max = getMinMax(V(0))
	)
	if int64(value) < min {
		return V(min)
	}
	if value > 0 && uint64(value) > max {
		return V(max)
	}
	// otherwise it fits
	return V(value)
}
