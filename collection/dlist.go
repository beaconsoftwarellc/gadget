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
type DListElement struct {
	prev *DListElement
	next *DListElement
	data interface{}
}

// Previous element in the Dlist
func (element DListElement) Previous() *DListElement { return element.prev }

// Next element in the DList.
func (element DListElement) Next() *DListElement { return element.next }

// Data in this DListElement
func (element DListElement) Data() interface{} { return element.data }

// DList is an implementation of a doubly linked list data structure.
type DList interface {
	// Size of the this dlist as a count of the elements in it.
	Size() int
	// Head of the list.
	Head() *DListElement
	// IsHead of the list.
	IsHead(element *DListElement) bool
	// Tail of the list.
	Tail() *DListElement
	// IsTail of the list.
	IsTail(element *DListElement) bool
	// InsertNext inserts the passed data after the passed element. If the list is empty a 'nil' element is allowed
	// otherwise an error will be returned.
	InsertNext(element *DListElement, data interface{}) (*DListElement, error)
	// InsertPrevious inserts the passed data before the passed element in the list. If the list is empty a 'nil' element
	// is allowed, otherwise an error will be returned.
	InsertPrevious(element *DListElement, data interface{}) (*DListElement, error)
	// Remove the element from the list.
	Remove(element *DListElement) (data interface{}, err error)
}

// dlinkedList is a threadsafe implementation of a doubly linked list
type dlinkedList struct {
	mutex *sync.Mutex
	head  *DListElement
	tail  *DListElement
	size  int
}

// NewDList returns a new initialized empty DList
func NewDList() DList {
	return &dlinkedList{mutex: &sync.Mutex{}}
}

// Size of the this dlist as a count of the elements in it.
func (list dlinkedList) Size() int { return list.size }

// Head of the list.
func (list dlinkedList) Head() *DListElement { return list.head }

// IsHead of the list.
func (list dlinkedList) IsHead(element *DListElement) bool { return element == list.head }

// Tail of the list.
func (list dlinkedList) Tail() *DListElement { return list.tail }

// IsTail of the list.
func (list dlinkedList) IsTail(element *DListElement) bool { return element == list.tail }

// InsertNext inserts the passed data after the passed element. If the list is empty a 'nil' element is allowed
// otherwise an error will be returned.
func (list *dlinkedList) InsertNext(element *DListElement, data interface{}) (newElement *DListElement, err error) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if 0 != list.Size() && nil == element {
		err = NewListNonEmptyError()
		return
	}
	newElement = &DListElement{data: data}
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
func (list *dlinkedList) InsertPrevious(element *DListElement, data interface{}) (newElement *DListElement, err error) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if 0 != list.Size() && nil == element {
		err = NewListNonEmptyError()
		return
	}
	newElement = &DListElement{data: data}
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
func (list *dlinkedList) Remove(element *DListElement) (data interface{}, err error) {
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

	data = element.data

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
		data = nil
		return
	}

	element.next = nil
	element.prev = nil

	list.size--
	return
}
