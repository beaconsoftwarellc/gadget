package specialized

import (
	"testing"
	"time"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
)

func TestRateHashPriorityQueue_Size(t *testing.T) {
	assert := assert1.New(t)
	q := NewRateHashPriorityQueue[string](1, time.Microsecond)
	assert.Equal(0, q.Size())
	sameHash := generator.String(20)
	q.Push(NewMockHashPriority(2, sameHash))
	assert.Equal(1, q.Size())

	q.Push(NewMockHashPriority(2, sameHash))
	assert.Equal(1, q.Size())

	q.Push(NewMockHashPriority(3, generator.String(20)))
	assert.Greater(3, q.Size())
	_, ok := q.Pop()
	assert.True(ok)
	assert.Greater(2, q.Size())
}

func TestRateHashPriorityQueue_Peek(t *testing.T) {
	assert := assert1.New(t)
	// this needs to be slow enough that it does not elapse before size is called,
	// but fast enough that Pop does not take too long to return
	q := NewRateHashPriorityQueue[string](1, 10*time.Millisecond)
	expected := NewMockHashPriority(3, generator.String(20))
	q.Push(expected)
	actual, ok := q.Peek()
	assert.True(ok)
	assert.Equal(expected, actual)
	assert.Equal(1, q.Size())
	q.Pop()
	actual, ok = q.Peek()
	assert.Nil(actual)
	assert.False(ok)
	assert.Equal(0, q.Size())
}

func TestRateHashPriorityQueue_Stop(t *testing.T) {
	assert := assert1.New(t)
	obj := NewRateHashPriorityQueue[string](1, time.Microsecond)
	q, ok := obj.(*rhpQueue[string])
	assert.True(ok)
	// making sure Stop is reentrant and does not block forever.
	q.Stop()
	q.Stop()
}

func TestRateHashPriorityQueue_Channel(t *testing.T) {
	assert := assert1.New(t)
	q := NewRateHashPriorityQueue[string](1, 1*time.Microsecond)
	expected := NewMockHashPriority(3, generator.String(20))
	q.Push(expected)
	var actual HashPriority[string]
	select {
	case actual = <-q.Channel():
		// noop
	case <-time.After(2 * time.Millisecond):
		assert.Fail("should have gotten an element")
	}
	assert.Equal(expected, actual)
}

func TestRateHashPriorityQueue_Pop(t *testing.T) {
	assert := assert1.New(t)
	// we have to get all the elements in before this time elapses once
	q := NewRateHashPriorityQueue[string](1, 55*time.Millisecond)
	for i := 0; i < 10; i++ {
		q.Push(NewMockHashPriority(i, generator.String(20)))
	}
	for j := 9; j >= 0; j-- {
		start := time.Now()
		elm, ok := q.Pop()
		assert.True(ok)
		// we want to make sure we waited at least 50ms. Don't make this dead on
		// with the rate limit as the timers are not millisecond accurate.
		assert.True(time.Now().Sub(start) > 50*time.Millisecond)
		assert.Equal(j, elm.GetPriority())
	}
}

func TestRateHashPriorityQueue_NoLimitPop(t *testing.T) {
	assert := assert1.New(t)
	// we have to get all the elements in before this time elapses once
	q := NewRateHashPriorityQueue[string](1, 50*time.Millisecond)
	for i := 0; i < 10; i++ {
		q.Push(NewMockHashPriority(i, generator.String(20)))
	}
	for j := 9; j >= 0; j-- {
		start := time.Now()
		elm, ok := q.NoLimitPop()
		assert.True(time.Now().Sub(start) < time.Millisecond)
		assert.True(ok)
		assert.Equal(j, elm.GetPriority())
	}
}
