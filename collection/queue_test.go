package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	assert := assert.New(t)
	queue := NewQueue()
	assert.Equal(0, queue.Size())
	actual, err := queue.Peek()
	assert.Nil(actual)
	assert.Error(err, NewEmptyListError().Error())

	actual, err = queue.Pop()
	assert.Nil(actual)
	assert.Error(err, NewEmptyListError().Error())
	assert.Equal(0, queue.Size())

	queue.Push("go")
	queue.Push("is")
	queue.Push("super")
	queue.Push("fun")

	actual, err = queue.Peek()
	assert.NoError(err)
	assert.Equal("go", actual)
	assert.Equal(4, queue.Size())

	actual, err = queue.Pop()
	assert.NoError(err)
	assert.Equal("go", actual)
	assert.Equal(3, queue.Size())

	actual, err = queue.Pop()
	assert.NoError(err)
	assert.Equal("is", actual)
	assert.Equal(2, queue.Size())

	actual, err = queue.Pop()
	assert.NoError(err)
	assert.Equal("super", actual)
	assert.Equal(1, queue.Size())

	actual, err = queue.Pop()
	assert.NoError(err)
	assert.Equal("fun", actual)
	assert.Equal(0, queue.Size())

	actual, err = queue.Pop()
	assert.Error(err, NewEmptyListError().Error())
	assert.Nil(actual)
	assert.Equal(0, queue.Size())
}
