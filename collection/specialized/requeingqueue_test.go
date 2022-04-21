package specialized

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/v2/collection"
)

func TestReuqueingQueue(t *testing.T) {
	assert := assert1.New(t)
	rqq := NewRequeueingQueue[string]()
	assert.Equal(0, rqq.Size())
	actual, err := rqq.Peek()
	assert.Empty(actual)
	assert.Error(err, collection.NewEmptyListError().Error())

	actual, err = rqq.Pop()
	assert.Empty(actual)
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
