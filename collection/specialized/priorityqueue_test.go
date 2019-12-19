package specialized

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/collection"
)

type MockPriority struct {
	priority int
}

func (mp *MockPriority) GetPriority() int {
	return mp.priority
}

func NewMockPriority(p int) *MockPriority {
	return &MockPriority{priority: p}
}

func TestPriorityQueue_Size(t *testing.T) {
	assert := assert.New(t)
	q := NewPriorityQueue()
	assert.Equal(0, q.Size())
	q.Push(NewMockPriority(1))
	assert.Equal(1, q.Size())
	q.Push(NewMockPriority(2))
	assert.Equal(2, q.Size())
	q.Push(NewMockPriority(3))
	assert.Equal(3, q.Size())
	q.Pop()
	assert.Equal(2, q.Size())
}

func TestPriorityQueue_Peek(t *testing.T) {
	assert := assert.New(t)
	q := NewPriorityQueue()
	expected := NewMockPriority(3)
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

func Test_priorityQueue_Pop(t *testing.T) {
	assert := assert.New(t)
	type fields struct {
		list collection.List
	}
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "Reverse Order",
			input:    []int{1, 2, 3, 4},
			expected: []int{4, 3, 2, 1},
		},
		{
			name:     "Correct Order",
			input:    []int{4, 3, 2, 1},
			expected: []int{4, 3, 2, 1},
		},
		{
			name:     "Interleaved",
			input:    []int{4, 2, 3, 1},
			expected: []int{4, 3, 2, 1},
		},
		{
			name:     "Dupes",
			input:    []int{4, 2, 2, 5, 5, 3, 1},
			expected: []int{5, 5, 4, 3, 2, 2, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewPriorityQueue()
			for i := 0; i < len(tt.input); i++ {
				q.Push(NewMockPriority(tt.input[i]))
			}
			actual := []int{}
			for p, ok := q.Pop(); ok; p, ok = q.Pop() {
				actual = append(actual, p.GetPriority())
			}
			assert.Equal(tt.expected, actual)
		})
	}
}
