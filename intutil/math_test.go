package intutil

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
)

func TestMin(t *testing.T) {
	assert := assert1.New(t)
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
	assert := assert1.New(t)
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
	assert := assert1.New(t)
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
	assert := assert1.New(t)
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
		intput   T
		expected T
	}

	var testData = struct {
		inttype   []testcase[int]
		int32type []testcase[int32]
	}{
		inttype: []testcase[int]{
			{
				intput:   5,
				expected: 5,
			},
			{
				intput:   0,
				expected: 0,
			},
			{
				intput:   -1,
				expected: 1,
			},
		},
		int32type: []testcase[int32]{
			{
				intput:   5,
				expected: 5,
			},
			{
				intput:   0,
				expected: 0,
			},
			{
				intput:   -1,
				expected: 1,
			},
		},
	}

	t.Run("int", func(t *testing.T) {
		assert := assert1.New(t)
		for _, test := range testData.inttype {
			assert.Equal(test.expected, Abs(test.intput))
		}
	})

	t.Run("int32", func(t *testing.T) {
		assert := assert1.New(t)
		for _, test := range testData.int32type {
			assert.Equal(test.expected, Abs(test.intput))
		}
	})

}

func TestDecrementor_Decrement_Resets(t *testing.T) {
	assert := assert1.New(t)
	decrementor := NewDecrementor(2)
	assert.Equal(1, decrementor.Decrement())
	assert.Equal(0, decrementor.Decrement())
	assert.Equal(decrementor.GetInitialMax()-1, decrementor.Decrement())
}

func TestDecrementor_Decrement(t *testing.T) {
	decrementor := NewDecrementor(10000)
	assert := assert1.New(t)
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
