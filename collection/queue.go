package collection

// Queue is an implementation of a queue (fifo) data structure
type Queue interface {
	// Size of the queue represented as a count of the elements in the queue.
	Size() int
	// Push a new data element onto the queue.
	Push(data interface{})
	// Pop the most recently pushed data element off the queue.
	Pop() (interface{}, error)
	// Peek returns the most recently pushed element without modifying the queue
	Peek() (interface{}, error)
}

type queue struct {
	list List
}

// NewQueue that is empty.
func NewQueue() Queue {
	return &queue{list: NewList()}
}

func (q *queue) Size() int {
	return q.list.Size()
}

func (q *queue) Push(data interface{}) {
	q.list.InsertNext(q.list.Tail(), data)
}

func (q *queue) Pop() (interface{}, error) {
	return q.list.RemoveNext(nil)
}

func (q *queue) Peek() (interface{}, error) {
	if q.list.Size() == 0 {
		return nil, NewEmptyListError()
	}
	return q.list.Head().Data(), nil
}
