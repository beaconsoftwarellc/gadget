package specialized

import "github.com/beaconsoftwarellc/gadget/collection"

// NewRequeueingQueue for a queue that never loses elements, they are just
// added back onto the end of the queue on pop. Useful for when you don't want
// to keep track of an array and an index on an object.
func NewRequeueingQueue() collection.Stack {
	return &requeueingQueue{list: collection.NewDList()}
}

type requeueingQueue struct {
	list collection.DList
}

// Size of the queue represented as a count of the elements in the queue.
func (q *requeueingQueue) Size() int { return q.list.Size() }

// Push a new data element onto the queue.
func (q *requeueingQueue) Push(data interface{}) {
	q.list.InsertPrevious(q.list.Head(), data)
}

// Pop the most recently pushed data element off the queue and put it at the end of the queue.
func (q *requeueingQueue) Pop() (data interface{}, err error) {
	head := q.list.Head()
	if nil == head {
		return nil, collection.NewEmptyListError()
	}
	q.list.InsertNext(q.list.Tail(), head.Data())
	return q.list.Remove(head)
}

// Peek returns the most recently pushed element without modifying the queue
func (q requeueingQueue) Peek() (interface{}, error) {
	if q.list.Size() == 0 {
		return nil, collection.NewEmptyListError()
	}
	return q.list.Head().Data(), nil
}
