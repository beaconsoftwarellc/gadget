package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDListHead(t *testing.T) {
	assert := assert.New(t)

	// test list initialization
	list := NewDList()
	assert.Nil(list.Head())
	assert.Equal(0, list.Size())

	// test first insertion is the head
	expected := "fun"
	_, err := list.InsertNext(list.Head(), expected)
	assert.NoError(err)
	head := list.Head()
	assert.NotNil(head)
	actual := head.Data()
	assert.Equal(expected, actual)
	assert.Nil(head.Next())
	assert.Nil(head.Previous())
	assert.Equal(1, list.Size())

	// test insertion after head
	newData := "with"
	_, err = list.InsertNext(list.Head(), newData)
	assert.NoError(err)
	head = list.Head()
	assert.NotNil(head)
	actual = head.Data()
	assert.Equal(expected, actual)
	assert.Equal(newData, head.Next().Data())
	assert.Equal(2, list.Size())

	// test insertion before the head
	expected = "go"
	_, err = list.InsertPrevious(list.Head(), expected)
	assert.NoError(err)
	head = list.Head()
	assert.NotNil(head)
	actual = head.Data()
	assert.Equal(expected, actual)

	// test head removal gets set back
	expected = "fun"
	list.Remove(list.Head())
	head = list.Head()
	assert.NotNil(head)
	actual = head.Data()
	assert.Equal(expected, actual)

	// test empty
	list.Remove(list.Head())
	list.Remove(list.Head())
	assert.Nil(list.Head())
}

func TestDListIsHead(t *testing.T) {
	assert := assert.New(t)
	list := NewDList()
	elm, err := list.InsertNext(nil, "foo")
	assert.NoError(err)
	assert.True(list.IsHead(elm))
	elm1, err := list.InsertPrevious(list.Head(), "bar")
	assert.NoError(err)
	assert.True(list.IsHead(elm1))
	assert.False(list.IsHead(elm))
	list.Remove(list.Head())
	assert.True(list.IsHead(elm))
}

func TestDListTail(t *testing.T) {
	assert := assert.New(t)
	// test list initialization
	list := NewDList()
	assert.Nil(list.Tail())

	// test first insertion is the tail
	expected := "fun"
	_, err := list.InsertNext(list.Tail(), expected)
	assert.NoError(err)
	// fun
	tail := list.Tail()
	assert.NotNil(tail)
	actual := tail.Data()
	assert.Equal(expected, actual)

	// test insertion after tail
	expected = "with"
	_, err = list.InsertNext(list.Tail(), expected)
	assert.NoError(err)
	// fun with
	tail = list.Tail()
	assert.NotNil(tail)
	actual = tail.Data()
	assert.Equal(expected, actual)

	// test insertion between head and tail does not affect tail
	newData := "go"
	_, err = list.InsertNext(list.Head(), newData)
	assert.NoError(err)
	//  fun go with
	tail = list.Tail()
	assert.NotNil(tail)
	actual = tail.Data()
	assert.Equal(expected, actual)

	// test tail removal
	expected = "go"
	list.Remove(list.Head().Next().Next())
	// fun go
	tail = list.Tail()
	assert.NotNil(tail)
	actual = tail.Data()
	assert.Equal(expected, actual)

	// test empty
	list.Remove(list.Head())
	// go
	list.Remove(list.Head())
	// {empty}
	assert.Nil(list.Tail())
}

func TestDListIsTail(t *testing.T) {
	assert := assert.New(t)
	list := NewDList()
	elm, err := list.InsertNext(nil, "foo")
	assert.NoError(err)
	assert.True(list.IsTail(elm))
	elm1, err := list.InsertNext(elm, "bar")
	assert.NoError(err)
	assert.False(list.IsTail(elm))
	assert.True(list.IsTail(elm1))
	list.Remove(elm.Next())
	assert.True(list.IsTail(elm))
}

func TestDListInsertNext(t *testing.T) {
	assert := assert.New(t)
	list := NewDList()
	assert.Equal(0, list.Size())
	elm, err := list.InsertNext(nil, "fun")
	assert.NoError(err)
	assert.Equal(1, list.Size())
	_, err = list.InsertNext(nil, "invalid")
	assert.EqualError(err, NewListNonEmptyError().Error())
	elm, err = list.InsertNext(elm, "with")
	assert.NoError(err)
	assert.Equal(2, list.Size())
	list.InsertNext(elm, "go")
	assert.Equal(3, list.Size())
	assert.Equal("fun", list.Head().Data())
	assert.Equal("with", list.Head().Next().Data())
	assert.Equal("go", list.Tail().Data())
	assert.Equal(list.Head().Next().Next().Data(), list.Tail().Data())
}

func TestDListInsertPrevious(t *testing.T) {
	assert := assert.New(t)
	list := NewDList()
	assert.Equal(0, list.Size())
	elm, err := list.InsertPrevious(nil, "fun")
	assert.NoError(err)
	assert.NotNil(elm)
	assert.Equal(1, list.Size())

	_, err = list.InsertPrevious(nil, "invalid")
	assert.EqualError(err, NewListNonEmptyError().Error())
	assert.Equal(1, list.Size())

	elm1, err := list.InsertPrevious(elm, "is")
	assert.NoError(err)
	assert.NotNil(elm1)
	assert.Equal(2, list.Size())
	assert.True(list.IsHead(elm1))

	elm2, err := list.InsertPrevious(elm1, "go")
	assert.NoError(err)
	assert.NotNil(elm1)
	assert.Equal(3, list.Size())
	assert.True(list.IsHead(elm2))

	elm3, err := list.InsertPrevious(elm, "super")
	assert.NoError(err)
	assert.NotNil(elm3)
	assert.Equal(4, list.Size())
	assert.False(list.IsHead(elm3))
	assert.True(list.IsHead(elm2))
	assert.Equal(elm3, elm.Previous())

	assert.Equal(list.Head().Data(), "go")
	assert.Equal(list.Head().Next().Data(), "is")
	assert.Equal(list.Head().Next().Next().Data(), "super")
	assert.Equal(list.Head().Next().Next().Next().Data(), "fun")
}

func TestDListRemove(t *testing.T) {
	assert := assert.New(t)
	list := NewDList()
	data := "foo"
	data1 := "bar"
	data2 := "baz"

	el, _ := list.InsertNext(nil, data)
	assert.NotNil(el)

	el1, _ := list.InsertNext(el, data1)
	assert.NotNil(el1)

	el2, _ := list.InsertNext(el1, data2)
	assert.NotNil(el2)

	n, err := list.Remove(nil)
	assert.Nil(n)
	assert.EqualError(err, NewNoElementError().Error())

	actual, err := list.Remove(el1)
	assert.NoError(err)
	assert.Equal(data1, actual)
	assert.Equal(2, list.Size())

	actual, err = list.Remove(el)
	assert.NoError(err)
	assert.Equal(data, actual)
	assert.Equal(1, list.Size())

	actual, err = list.Remove(el1)
	assert.EqualError(err, NewNoMemberError().Error())
	assert.Nil(actual)
	assert.Equal(1, list.Size())

	actual, err = list.Remove(el2)
	assert.NoError(err)
	assert.Equal(data2, actual)
	assert.Equal(0, list.Size())

	actual, err = list.Remove(el1)
	assert.EqualError(err, NewEmptyListError().Error())
	assert.Nil(actual)
}
