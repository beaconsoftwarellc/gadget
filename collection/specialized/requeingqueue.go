package specialized

import "github.com/beaconsoftwarellc/gadget/v2/collection"

// NewRequeueingQueue for a queue that never loses elements, they are just
// added back onto the end of the queue on pop. Useful for when you don't want
// to keep track of an array and an index on an object.
func NewRequeueingQueue[T any]() collection.Stack[T] {
	return &requeueingQueue[T]{list: collection.NewDList[T]()}
}

type requeueingQueue[T any] struct {
	list collection.DList[T]
}

// Size of the queue represented as a count of the elements in the queue.
func (q *requeueingQueue[T]) Size() int { return q.list.Size() }

// Push a new data element onto the queue.
func (q *requeueingQueue[T]) Push(data T) {
	q.list.InsertPrevious(q.list.Head(), data)
}

// Pop the most recently pushed data element off the queue and put it at the end of the queue.
func (q *requeueingQueue[T]) Pop() (data T, err error) {
	head := q.list.Head()
	if nil == head {
		var ret T
		return ret, collection.NewEmptyListError()
	}
	q.list.InsertNext(q.list.Tail(), head.Data())
	return q.list.Remove(head)
}

// Peek returns the most recently pushed element without modifying the queue
func (q requeueingQueue[T]) Peek() (T, error) {
	if q.list.Size() == 0 {
		var ret T
		return ret, collection.NewEmptyListError()
	}
	return q.list.Head().Data(), nil
}
