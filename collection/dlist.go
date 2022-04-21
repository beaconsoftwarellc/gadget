package collection

import (
	"sync"

	"github.com/beaconsoftwarellc/gadget/errors"
)

// ListNonEmptyError is returned when a empty list is required for the operation.
type ListNonEmptyError struct{ trace []string }

func (err *ListNonEmptyError) Error() string {
	return "list is not empty"
}

// Trace returns the stack trace for the error
func (err *ListNonEmptyError) Trace() []string {
	return err.trace
}

// NewListNonEmptyError instantiates a ListNonEmptyError with a stack trace
func NewListNonEmptyError() errors.TracerError {
	return &ListNonEmptyError{trace: errors.GetStackTrace()}
}

// NoMemberError is returned when an element is passed that is not a member of a list.
type NoMemberError struct{ trace []string }

func (err *NoMemberError) Error() string {
	return "element is not a member of the list"
}

// Trace returns the stack trace for the error
func (err *NoMemberError) Trace() []string {
	return err.trace
}

// NewNoMemberError instantiates a NoMemberError with a stack trace
func NewNoMemberError() errors.TracerError {
	return &NoMemberError{trace: errors.GetStackTrace()}
}

// DListElement is the primary element for use inside of a DList
type DListElement[T any] struct {
	prev *DListElement[T]
	next *DListElement[T]
	data T
}

// Previous element in the Dlist
func (element DListElement[T]) Previous() *DListElement[T] { return element.prev }

// Next element in the DList.
func (element DListElement[T]) Next() *DListElement[T] { return element.next }

// Data in this DListElement[T]
func (element DListElement[T]) Data() T { return element.data }

// DList is an implementation of a doubly linked list data structure.
type DList[T any] interface {
	// Size of the this dlist as a count of the elements in it.
	Size() int
	// Head of the list.
	Head() *DListElement[T]
	// IsHead of the list.
	IsHead(element *DListElement[T]) bool
	// Tail of the list.
	Tail() *DListElement[T]
	// IsTail of the list.
	IsTail(element *DListElement[T]) bool
	// InsertNext inserts the passed data after the passed element. If the list is empty a 'nil' element is allowed
	// otherwise an error will be returned.
	InsertNext(element *DListElement[T], data T) (*DListElement[T], error)
	// InsertPrevious inserts the passed data before the passed element in the list. If the list is empty a 'nil' element
	// is allowed, otherwise an error will be returned.
	InsertPrevious(element *DListElement[T], data T) (*DListElement[T], error)
	// Remove the element from the list.
	Remove(element *DListElement[T]) (data T, err error)
}

// dlinkedList is a threadsafe implementation of a doubly linked list
type dlinkedList[T any] struct {
	mutex *sync.Mutex
	head  *DListElement[T]
	tail  *DListElement[T]
	size  int
}

// NewDList returns a new initialized empty DList
func NewDList[T any]() DList[T] {
	return &dlinkedList[T]{mutex: &sync.Mutex{}}
}

// Size of the this dlist as a count of the elements in it.
func (list dlinkedList[T]) Size() int { return list.size }

// Head of the list.
func (list dlinkedList[T]) Head() *DListElement[T] { return list.head }

// IsHead of the list.
func (list dlinkedList[T]) IsHead(element *DListElement[T]) bool { return element == list.head }

// Tail of the list.
func (list dlinkedList[T]) Tail() *DListElement[T] { return list.tail }

// IsTail of the list.
func (list dlinkedList[T]) IsTail(element *DListElement[T]) bool { return element == list.tail }

// InsertNext inserts the passed data after the passed element. If the list is empty a 'nil' element is allowed
// otherwise an error will be returned.
func (list *dlinkedList[T]) InsertNext(element *DListElement[T], data T) (newElement *DListElement[T], err error) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if 0 != list.Size() && nil == element {
		err = NewListNonEmptyError()
		return
	}
	newElement = &DListElement[T]{data: data}
	if 0 == list.Size() {
		list.head = newElement
		list.tail = newElement
	} else {
		newElement.next = element.next
		newElement.prev = element
		if nil == element.next {
			list.tail = newElement
		} else {
			element.next.prev = newElement
		}
		element.next = newElement
	}
	list.size++
	return
}

// InsertPrevious inserts the passed data before the passed element in the list. If the list is empty a 'nil' element
// is allowed, otherwise an error will be returned.
func (list *dlinkedList[T]) InsertPrevious(element *DListElement[T], data T) (newElement *DListElement[T], err error) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if 0 != list.Size() && nil == element {
		err = NewListNonEmptyError()
		return
	}
	newElement = &DListElement[T]{data: data}
	if 0 == list.Size() {
		list.head = newElement
		list.tail = newElement
	} else {
		newElement.next = element
		newElement.prev = element.prev
		if nil == element.prev {
			list.head = newElement
		} else {
			element.prev.next = newElement
		}
		element.prev = newElement
	}
	list.size++
	return
}

// Remove the element from the list.
func (list *dlinkedList[T]) Remove(element *DListElement[T]) (data T, err error) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if 0 == list.size {
		err = NewEmptyListError()
		return
	}

	if nil == element {
		err = NewNoElementError()
		return
	}

	if list.IsHead(element) {
		list.head = element.next
		if nil == list.head {
			list.tail = nil
		} else {
			element.next.prev = nil
		}
	} else if nil != element.prev {
		element.prev.next = element.next
		if nil == element.next {
			list.tail = element.prev
		} else {
			element.next.prev = element.prev
		}
	} else {
		// the element is not the head and has nil for it's prev so it is not a member
		err = NewNoMemberError()
		return
	}

	data = element.data
	element.next = nil
	element.prev = nil

	list.size--
	return
}
