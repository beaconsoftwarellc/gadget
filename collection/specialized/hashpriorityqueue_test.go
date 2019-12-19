package specialized

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/collection"
	"github.com/beaconsoftwarellc/gadget/generator"
)

type MockHashPriority struct {
	priority int
	hash     string
}

func (mp *MockHashPriority) GetPriority() int {
	return mp.priority
}

func (mp *MockHashPriority) GetHash() interface{} {
	return mp.hash
}

func NewMockHashPriority(p int, s string) *MockHashPriority {
	return &MockHashPriority{priority: p, hash: s}
}

func TestHashPriorityQueue_Size(t *testing.T) {
	assert := assert.New(t)
	q := NewHashPriorityQueue()
	assert.Equal(0, q.Size())
	sameHash := generator.String(20)
	q.Push(NewMockHashPriority(2, sameHash))
	assert.Equal(1, q.Size())
	q.Push(NewMockHashPriority(2, sameHash))
	assert.Equal(1, q.Size())
	q.Push(NewMockHashPriority(3, generator.String(20)))
	assert.Equal(2, q.Size())
	q.Pop()
	assert.Equal(1, q.Size())
}

func TestHashPriorityQueue_Peek(t *testing.T) {
	assert := assert.New(t)
	q := NewHashPriorityQueue()
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

func TestHashPriorityQueue(t *testing.T) {
	assert := assert.New(t)
	type fields struct {
		list collection.List
	}
	tests := []struct {
		name      string
		pinput    []int
		hinput    []string
		hexpected []interface{}
	}{
		{
			name:      "Reverse Order",
			pinput:    []int{1, 2, 3, 4},
			hinput:    []string{"d", "c", "b", "a"},
			hexpected: []interface{}{"a", "b", "c", "d"},
		},
		{
			name:      "Correct Order",
			pinput:    []int{4, 3, 2, 1},
			hinput:    []string{"a", "b", "c", "d"},
			hexpected: []interface{}{"a", "b", "c", "d"},
		},
		{
			name:      "Dupes",
			pinput:    []int{4, 2, 2, 5, 5, 3, 1},
			hinput:    []string{"a", "b", "c", "d", "d", "f", "g"},
			hexpected: []interface{}{"d", "a", "f", "c", "b", "g"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewHashPriorityQueue()
			for i := 0; i < len(tt.pinput); i++ {
				q.Push(NewMockHashPriority(tt.pinput[i], tt.hinput[i]))
			}
			actual := []interface{}{}
			for p, ok := q.Pop(); ok; p, ok = q.Pop() {
				actual = append(actual, p.GetHash())
			}
			assert.Equal(tt.hexpected, actual)
		})
	}
}
