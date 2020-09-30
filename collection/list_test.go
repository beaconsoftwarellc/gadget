package collection

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func TestListHead(t *testing.T) {
	assert := assert1.New(t)

	// test list initialization
	list := NewList()
	assert.Nil(list.Head())

	// test first insertion is the head
	expected := "fun"
	list.InsertNext(list.Head(), expected)
	head := list.Head()
	assert.NotNil(head)
	actual := head.Data()
	assert.Equal(expected, actual)

	// test insertion after head
	newData := "with"
	list.InsertNext(list.Head(), newData)
	head = list.Head()
	assert.NotNil(head)
	actual = head.Data()
	assert.Equal(expected, actual)

	// test insertion before the head
	expected = "go"
	list.InsertNext(nil, expected)
	head = list.Head()
	assert.NotNil(head)
	actual = head.Data()
	assert.Equal(expected, actual)

	// test head removal gets set back
	expected = "fun"
	list.RemoveNext(nil)
	head = list.Head()
	assert.NotNil(head)
	actual = head.Data()
	assert.Equal(expected, actual)

	// test empty
	list.RemoveNext(nil)
	list.RemoveNext(nil)
	assert.Nil(list.Head())
}

func TestListIsHead(t *testing.T) {
	assert := assert1.New(t)
	list := NewList()
	elm := list.InsertNext(nil, "foo")
	assert.True(list.IsHead(elm))
	elm1 := list.InsertNext(nil, "bar")
	assert.True(list.IsHead(elm1))
	assert.False(list.IsHead(elm))
	list.RemoveNext(nil)
	assert.True(list.IsHead(elm))
}

func TestListTail(t *testing.T) {
	assert := assert1.New(t)
	// test list initialization
	list := NewList()
	assert.Nil(list.Tail())

	// test first insertion is the tail
	expected := "fun"
	list.InsertNext(list.Tail(), expected)
	// fun
	tail := list.Tail()
	assert.NotNil(tail)
	actual := tail.Data()
	assert.Equal(expected, actual)

	// test insertion after tail
	expected = "with"
	list.InsertNext(list.Tail(), expected)
	// fun with
	tail = list.Tail()
	assert.NotNil(tail)
	actual = tail.Data()
	assert.Equal(expected, actual)

	// test insertion between head and tail does not affect tail
	newData := "go"
	list.InsertNext(list.Head(), newData)
	//  fun go with
	tail = list.Tail()
	assert.NotNil(tail)
	actual = tail.Data()
	assert.Equal(expected, actual)

	// test tail removal
	expected = "go"
	list.RemoveNext(list.Head().Next())
	// fun go
	tail = list.Tail()
	assert.NotNil(tail)
	actual = tail.Data()
	assert.Equal(expected, actual)

	// test empty
	list.RemoveNext(nil)
	// go
	list.RemoveNext(nil)
	// {empty}
	assert.Nil(list.Tail())
}

func TestListIsTail(t *testing.T) {
	assert := assert1.New(t)
	list := NewList()
	elm := list.InsertNext(nil, "foo")
	assert.True(list.IsTail(elm))
	elm1 := list.InsertNext(elm, "bar")
	assert.False(list.IsTail(elm))
	assert.True(list.IsTail(elm1))
	list.RemoveNext(elm)
	assert.True(list.IsTail(elm))
}

func TestInsertNext(t *testing.T) {
	assert := assert1.New(t)
	list := NewList()
	assert.Equal(0, list.Size())
	elm := list.InsertNext(nil, "fun")
	assert.Equal(1, list.Size())
	elm = list.InsertNext(elm, "with")
	assert.Equal(2, list.Size())
	list.InsertNext(elm, "go")
	assert.Equal(3, list.Size())
	assert.Equal("fun", list.Head().Data())
	assert.Equal("with", list.Head().Next().Data())
	assert.Equal("go", list.Tail().Data())
	assert.Equal(list.Head().Next().Next().Data(), list.Tail().Data())
}

func TestRemoveNext(t *testing.T) {
	assert := assert1.New(t)
	list := NewList()
	data, err := list.RemoveNext(nil)
	assert.Nil(data)
	assert.EqualError(err, NewEmptyListError().Error())

	elm := list.InsertNext(nil, "foo")
	data, err = list.RemoveNext(elm)
	assert.Nil(data)
	assert.EqualError(err, NewNoElementError().Error())

	expected := "bar"
	list.InsertNext(list.Head(), expected)
	actual, err := list.RemoveNext(list.Head())
	assert.NoError(err)
	assert.Equal(expected, actual)

	expected1 := list.Head().Data()
	actual, err = list.RemoveNext(nil)
	assert.NoError(err)
	assert.Equal(expected1, actual)
	assert.Equal(0, list.Size())

	elm1 := list.InsertNext(nil, "bar")
	assert.NotNil(elm1)
	list.RemoveNext(elm)
	assert.Equal(1, list.Size())
}

func insert(list List, ia []int, done chan bool) {
	for _, i := range ia {
		list.InsertNext(nil, i)
	}
	done <- true
}

func TestListConcurrentInsert(t *testing.T) {
	assert := assert1.New(t)
	var done = make(chan bool, 2)
	elms := []int{1, 2, 3}
	elms1 := []int{4, 5, 6}
	list := NewList()
	go insert(list, elms, done)
	go insert(list, elms1, done)
	<-done
	<-done
	assert.Equal(6, list.Size())
}
