package intutil

import (
	"math"
	"testing"

	_assert "github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
)

func TestMin(t *testing.T) {
	assert := _assert.New(t)
	var testData = []struct {
		a        int
		b        int
		expected int
	}{
		{1, 2, 1},
		{3, 2, 2},
		{3, 3, 3},
	}
	for _, data := range testData {
		assert.Equal(data.expected, Min(data.a, data.b))
	}
}

func TestMinv(t *testing.T) {
	assert := _assert.New(t)
	var testData = []struct {
		data     []int
		expected int
	}{
		{[]int{}, 0},
		{[]int{1, 2, 3, 4, 5}, 1},
		{[]int{5, 4, 3, 2, 1}, 1},
		{[]int{5, 5, 5, 5, 5}, 5},
	}
	for _, test := range testData {
		assert.Equal(test.expected, Minv(test.data...))
	}
}

func TestMax(t *testing.T) {
	assert := _assert.New(t)
	var testData = []struct {
		a        int
		b        int
		expected int
	}{
		{1, 2, 2},
		{3, 2, 3},
		{3, 3, 3},
	}
	for _, data := range testData {
		assert.Equal(data.expected, Max(data.a, data.b))
	}
}

func TestMaxv(t *testing.T) {
	assert := _assert.New(t)
	var testData = []struct {
		data     []int
		expected int
	}{
		{[]int{}, 0},
		{[]int{1, 2, 3, 4, 5}, 5},
		{[]int{5, 4, 3, 2, 1}, 5},
		{[]int{5, 5, 5, 5, 5}, 5},
	}
	for _, test := range testData {
		assert.Equal(test.expected, Maxv(test.data...))
	}
}

func TestAbs(t *testing.T) {
	type testcase[T constraints.Signed] struct {
		input    T
		expected T
	}

	var testData = struct {
		inttype   []testcase[int]
		int32type []testcase[int32]
	}{
		inttype: []testcase[int]{
			{
				input:    5,
				expected: 5,
			},
			{
				input:    0,
				expected: 0,
			},
			{
				input:    -1,
				expected: 1,
			},
		},
		int32type: []testcase[int32]{
			{
				input:    5,
				expected: 5,
			},
			{
				input:    0,
				expected: 0,
			},
			{
				input:    -1,
				expected: 1,
			},
		},
	}

	t.Run("int", func(t *testing.T) {
		assert := _assert.New(t)
		for _, test := range testData.inttype {
			assert.Equal(test.expected, Abs(test.input))
		}
	})

	t.Run("int32", func(t *testing.T) {
		assert := _assert.New(t)
		for _, test := range testData.int32type {
			assert.Equal(test.expected, Abs(test.input))
		}
	})

}

func TestDecrementor_Decrement_Resets(t *testing.T) {
	assert := _assert.New(t)
	decrementor := NewDecrementor(2)
	assert.Equal(1, decrementor.Decrement())
	assert.Equal(0, decrementor.Decrement())
	assert.Equal(decrementor.GetInitialMax()-1, decrementor.Decrement())
}

func TestDecrementor_Decrement(t *testing.T) {
	decrementor := NewDecrementor(10000)
	assert := _assert.New(t)
	ch := make(chan bool)
	f := func() {
		last := 0
		current := 0
		for i := 0; i < 1000; i++ {
			current = decrementor.Decrement()
			if i != 0 {
				assert.True(current < last, "expected %d < %d", current, last)
			}
			last = current
		}
		ch <- true
	}

	go f()
	go f()

	<-ch
	<-ch
}

func TestRange(t *testing.T) {
	type testcase[T constraints.Signed] struct {
		input    T
		expected T
		min      T
		max      T
	}

	var testData = struct {
		inttype   []testcase[int]
		int32type []testcase[int32]
	}{
		inttype: []testcase[int]{
			{
				input:    5,
				expected: 5,
				min:      0,
				max:      6,
			},
			{
				input:    0,
				expected: 1,
				min:      1,
				max:      6,
			},
			{
				input:    -1,
				expected: -10,
				max:      -10,
				min:      -100,
			},
		},
		int32type: []testcase[int32]{
			{
				input:    5,
				expected: 5,
				min:      0,
				max:      6,
			},
			{
				input:    0,
				expected: 1,
				min:      1,
				max:      6,
			},
		},
	}

	t.Run("int", func(t *testing.T) {
		assert := _assert.New(t)
		for _, test := range testData.inttype {
			assert.Equal(test.expected, Clamp(test.input, test.min, test.max))
		}
	})

	t.Run("int32", func(t *testing.T) {
		assert := _assert.New(t)
		for _, test := range testData.int32type {
			assert.Equal(test.expected, Clamp(test.input, test.min, test.max))
		}
	})
}

func TestClampCast(t *testing.T) {
	assert := _assert.New(t)

	t.Run("int64 -> int8", func(t *testing.T) {
		var actual int8
		actual = ClampCast[int64, int8](math.MaxInt8 + 1)
		assert.Equal(int8(math.MaxInt8), actual)

		actual = ClampCast[int64, int8](math.MinInt8 - 1)
		assert.Equal(int8(math.MinInt8), actual)
	})

	t.Run("int8 -> int64", func(t *testing.T) {
		var actual int64
		actual = ClampCast[int8, int64](math.MaxInt8)
		assert.Equal(int64(math.MaxInt8), actual)

		actual = ClampCast[int8, int64](math.MinInt8)
		assert.Equal(int64(math.MinInt8), actual)
	})

	t.Run("uint64 -> int64", func(t *testing.T) {
		var actual int64
		actual = ClampCast[uint64, int64](math.MaxUint64)
		assert.Equal(int64(math.MaxInt64), actual)

		actual = ClampCast[uint64, int64](0)
		assert.Equal(int64(0), actual)
	})

	t.Run("int64 -> uint64", func(t *testing.T) {
		var actual uint64
		actual = ClampCast[int64, uint64](math.MaxInt64)
		assert.Equal(uint64(math.MaxInt64), actual)

		actual = ClampCast[int64, uint64](math.MinInt64)
		assert.Equal(uint64(0), actual)
	})
}
