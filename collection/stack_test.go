package collection

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	assert := assert1.New(t)
	stack := NewStack()
	assert.Equal(0, stack.Size())
	actual, err := stack.Peek()
	assert.Nil(actual)
	assert.Error(err, NewEmptyListError().Error())

	actual, err = stack.Pop()
	assert.Nil(actual)
	assert.Error(err, NewEmptyListError().Error())
	assert.Equal(0, stack.Size())

	stack.Push("fun")
	stack.Push("super")
	stack.Push("is")
	stack.Push("go")

	actual, err = stack.Peek()
	assert.NoError(err)
	assert.Equal("go", actual)
	assert.Equal(4, stack.Size())

	actual, err = stack.Pop()
	assert.NoError(err)
	assert.Equal("go", actual)
	assert.Equal(3, stack.Size())

	actual, err = stack.Pop()
	assert.NoError(err)
	assert.Equal("is", actual)
	assert.Equal(2, stack.Size())

	actual, err = stack.Pop()
	assert.NoError(err)
	assert.Equal("super", actual)
	assert.Equal(1, stack.Size())

	actual, err = stack.Pop()
	assert.NoError(err)
	assert.Equal("fun", actual)
	assert.Equal(0, stack.Size())

	actual, err = stack.Pop()
	assert.Error(err, NewEmptyListError().Error())
	assert.Nil(actual)
	assert.Equal(0, stack.Size())
}

func TestStringStack(t *testing.T) {
	assert := assert1.New(t)
	stack := NewStringStack()
	assert.Equal(0, stack.Size())
	actual, err := stack.Peek()
	assert.Equal("", actual)
	assert.Error(err, NewEmptyListError().Error())

	actual, err = stack.Pop()
	assert.Equal("", actual)
	assert.Error(err, NewEmptyListError().Error())
	assert.Equal(0, stack.Size())

	stack.Push("fun")
	stack.Push("super")
	stack.Push("is")
	stack.Push("go")

	actual, err = stack.Peek()
	assert.NoError(err)
	assert.Equal("go", actual)
	assert.Equal(4, stack.Size())

	actual, err = stack.Pop()
	assert.NoError(err)
	assert.Equal("go", actual)
	assert.Equal(3, stack.Size())

	actual, err = stack.Pop()
	assert.NoError(err)
	assert.Equal("is", actual)
	assert.Equal(2, stack.Size())

	actual, err = stack.Pop()
	assert.NoError(err)
	assert.Equal("super", actual)
	assert.Equal(1, stack.Size())

	actual, err = stack.Pop()
	assert.NoError(err)
	assert.Equal("fun", actual)
	assert.Equal(0, stack.Size())

	actual, err = stack.Pop()
	assert.Error(err, NewEmptyListError().Error())
	assert.Equal("", actual)
	assert.Equal(0, stack.Size())
}
