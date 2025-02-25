package collection

import (
	"testing"

	_assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_require "github.com/stretchr/testify/require"
)

func TestListHead(t *testing.T) {
	assert := _assert.New(t)
	require := _require.New(t)

	// test list initialization
	list := NewList[string]()
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
	_, err := list.RemoveNext(nil)
	require.NoError(err)
	head = list.Head()
	assert.NotNil(head)
	actual = head.Data()
	assert.Equal(expected, actual)

	// test empty
	_, err = list.RemoveNext(nil)
	require.NoError(err)
	_, err = list.RemoveNext(nil)
	require.NoError(err)
	assert.Nil(list.Head())
}

func TestListIsHead(t *testing.T) {
	assert := _assert.New(t)
	list := NewList[string]()
	elm := list.InsertNext(nil, "foo")
	assert.True(list.IsHead(elm))
	elm1 := list.InsertNext(nil, "bar")
	assert.True(list.IsHead(elm1))
	assert.False(list.IsHead(elm))
	_, err := list.RemoveNext(nil)
	require.NoError(t, err)
	assert.True(list.IsHead(elm))
}

func TestListTail(t *testing.T) {
	assert := _assert.New(t)
	require := _require.New(t)
	var err error
	// test list initialization
	list := NewList[string]()
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
	_, err = list.RemoveNext(list.Head().Next())
	require.NoError(err)
	// fun go
	tail = list.Tail()
	assert.NotNil(tail)
	actual = tail.Data()
	assert.Equal(expected, actual)

	// test empty
	_, err = list.RemoveNext(nil)
	require.NoError(err)
	// go
	_, err = list.RemoveNext(nil)
	require.NoError(err)
	// {empty}
	assert.Nil(list.Tail())
}

func TestListIsTail(t *testing.T) {
	assert := _assert.New(t)
	require := _require.New(t)
	list := NewList[string]()
	elm := list.InsertNext(nil, "foo")
	assert.True(list.IsTail(elm))
	elm1 := list.InsertNext(elm, "bar")
	assert.False(list.IsTail(elm))
	assert.True(list.IsTail(elm1))
	_, err := list.RemoveNext(elm)
	require.NoError(err)
	assert.True(list.IsTail(elm))
}

func TestInsertNext(t *testing.T) {
	assert := _assert.New(t)
	list := NewList[string]()
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
	assert := _assert.New(t)
	require := _require.New(t)
	list := NewList[string]()
	data, err := list.RemoveNext(nil)
	assert.Empty(data)
	assert.EqualError(err, NewEmptyListError().Error())

	elm := list.InsertNext(nil, "foo")
	data, err = list.RemoveNext(elm)
	assert.Empty(data)
	assert.EqualError(err, NewNoElementError().Error())

	expected := "bar"
	list.InsertNext(list.Head(), expected)
	actual, err := list.RemoveNext(list.Head())
	require.NoError(err)
	assert.Equal(expected, actual)

	expected1 := list.Head().Data()
	actual, err = list.RemoveNext(nil)
	require.NoError(err)
	assert.Equal(expected1, actual)
	assert.Equal(0, list.Size())

	elm1 := list.InsertNext(nil, "bar")
	assert.NotNil(elm1)
	_, err = list.RemoveNext(elm)
	require.EqualError(err, NewNoElementError().Error())
	assert.Equal(1, list.Size())
}

func insert[T any](list List[T], ia []T, done chan bool) {
	for _, i := range ia {
		list.InsertNext(nil, i)
	}
	done <- true
}

func TestListConcurrentInsert(t *testing.T) {
	assert := _assert.New(t)
	var done = make(chan bool, 2)
	elms := []int{1, 2, 3}
	elms1 := []int{4, 5, 6}
	list := NewList[int]()
	go insert(list, elms, done)
	go insert(list, elms1, done)
	<-done
	<-done
	assert.Equal(6, list.Size())
}
