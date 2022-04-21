package collection

import (
	"sync"

	"github.com/beaconsoftwarellc/gadget/errors"
)

// EmptyListError is returned when a operation is called on an empty list that requires at least one element in the list.
type EmptyListError struct{ trace []string }

func (err *EmptyListError) Error() string {
	return "empty list"
}

// Trace returns the stack trace for the error
func (err *EmptyListError) Trace() []string {
	return err.trace
}

// NewEmptyListError instantiates a EmptyListError with a stack trace
func NewEmptyListError() errors.TracerError {
	return &EmptyListError{trace: errors.GetStackTrace()}
}

// NoElementError is returned when an operation that requires an element is called when no element is present at that
// position.
type NoElementError struct{ trace []string }

func (err *NoElementError) Error() string {
	return "no element present or element was nil"
}

// Trace returns the stack trace for the error
func (err *NoElementError) Trace() []string {
	return err.trace
}

// NewNoElementError instantiates a NoElementError with a stack trace
func NewNoElementError() errors.TracerError {
	return &NoElementError{trace: errors.GetStackTrace()}
}

// ListElement is a singly linked node in a list.
type ListElement[T any] struct {
	next *ListElement[T]
	data T
}

// Next returns the element the follows this element.
func (listElement ListElement[T]) Next() (element *ListElement[T]) {
	return listElement.next
}

// Data returns the data contained in this element.
func (listElement ListElement[T]) Data() T {
	return listElement.data
}

// List is a singly linked list implementation
type List[T any] interface {
	// Head (first element) of the list.
	Head() *ListElement[T]
	// Tail (last element) of the list.
	Tail() (element *ListElement[T])
	// IsHead returns a boolean indicating whether the passed element is the first element in the list.
	IsHead(element *ListElement[T]) bool
	// IsTail returns a boolean indicating if the passed element is the last element in the list.
	IsTail(element *ListElement[T]) bool
	// Size of the list (number of elements).
	Size() int
	// InsertNext data into the list after the passed element. If the element is nil, the data will be inserted at
	// the head of the list.
	InsertNext(element *ListElement[T], data T) *ListElement[T]
	// RemoveNext element from the list and return it's data. If passed element is 'nil' the head will be removed
	// from the list.
	RemoveNext(element *ListElement[T]) (data T, err error)
}

// linkedList is an implementation of a thread safe singly linked list.
type linkedList[T any] struct {
	mutex sync.RWMutex
	size  int
	head  *ListElement[T]
	tail  *ListElement[T]
}

// // NewList returns a new initialized list.
func NewList[T any]() List[T] {
	return &linkedList[T]{head: nil, tail: nil}
}

// Head (first element) of the list.
func (list *linkedList[T]) Head() *ListElement[T] {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.head
}

// Tail (last element) of the list.
func (list *linkedList[T]) Tail() (element *ListElement[T]) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.tail
}

// IsHead returns a boolean indicating whether the passed element is the first element in the list.
func (list *linkedList[T]) IsHead(element *ListElement[T]) bool {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return element == list.head
}

// IsTail returns a boolean indicating if the passed element is the last element in the list.
func (list *linkedList[T]) IsTail(element *ListElement[T]) bool {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return element == list.tail
}

// Size of the list (number of elements).
func (list *linkedList[T]) Size() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.size
}

// InsertNext data into the list after the passed element. If the element is nil, the data will be inserted at
// the head of the list.
func (list *linkedList[T]) InsertNext(element *ListElement[T], data T) *ListElement[T] {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	newElement := &ListElement[T]{data: data}
	if nil == element {
		if 0 == list.size {
			list.tail = newElement
		}
		newElement.next = list.head
		list.head = newElement
	} else {
		if nil == element.next {
			list.tail = newElement
		}
		newElement.next = element.next
		element.next = newElement
	}
	list.size++
	return newElement
}

// RemoveNext element from the list and return it's data. If passed element is 'nil' the head will be removed
// from the list.
func (list *linkedList[T]) RemoveNext(element *ListElement[T]) (data T, err error) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	if 0 == list.size {
		return data, NewEmptyListError()
	}

	if nil == element {
		data = list.head.data
		list.head = list.head.next
		if 1 == list.size {
			list.tail = nil
		}
	} else {
		if nil == element.next {
			return data, NewNoElementError()
		}
		data = element.next.data
		element.next = element.next.next
		if nil == element.next {
			list.tail = element
		}
		element.next = nil
	}

	list.size--
	return
}
