package specialized

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/collection"
)

func TestReuqueingQueue(t *testing.T) {
	assert := assert.New(t)
	rqq := NewRequeueingQueue()
	assert.Equal(0, rqq.Size())
	actual, err := rqq.Peek()
	assert.Nil(actual)
	assert.Error(err, collection.NewEmptyListError().Error())

	actual, err = rqq.Pop()
	assert.Nil(actual)
	assert.Error(err, collection.NewEmptyListError().Error())
	assert.Equal(0, rqq.Size())

	rqq.Push("fun")
	rqq.Push("super")
	rqq.Push("is")
	rqq.Push("go")

	actual, err = rqq.Peek()
	assert.NoError(err)
	assert.Equal("go", actual)
	assert.Equal(4, rqq.Size())

	actual, err = rqq.Pop()
	assert.NoError(err)
	assert.Equal("go", actual)
	assert.Equal(4, rqq.Size())

	actual, err = rqq.Pop()
	assert.NoError(err)
	assert.Equal("is", actual)
	assert.Equal(4, rqq.Size())

	actual, err = rqq.Pop()
	assert.NoError(err)
	assert.Equal("super", actual)
	assert.Equal(4, rqq.Size())

	actual, err = rqq.Pop()
	assert.NoError(err)
	assert.Equal("fun", actual)
	assert.Equal(4, rqq.Size())

	actual, err = rqq.Pop()
	assert.NoError(err)
	assert.Equal("go", actual)
	assert.Equal(4, rqq.Size())
}
