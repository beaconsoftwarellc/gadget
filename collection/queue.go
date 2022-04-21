package collection

// Queue is an implementation of a queue (fifo) data structure
type Queue[T any] interface {
	// Size of the queue represented as a count of the elements in the queue.
	Size() int
	// Push a new data element onto the queue.
	Push(data T)
	// Pop the most recently pushed data element off the queue.
	Pop() (T, error)
	// Peek returns the most recently pushed element without modifying the queue
	Peek() (T, error)
}

type queue[T any] struct {
	list List[T]
}

// NewQueue that is empty.
func NewQueue[T any]() Queue[T] {
	return &queue[T]{list: NewList[T]()}
}

func (q *queue[T]) Size() int {
	return q.list.Size()
}

func (q *queue[T]) Push(data T) {
	q.list.InsertNext(q.list.Tail(), data)
}

func (q *queue[T]) Pop() (T, error) {
	return q.list.RemoveNext(nil)
}

func (q *queue[T]) Peek() (T, error) {
	if q.list.Size() == 0 {
		var ret T
		return ret, NewEmptyListError()
	}
	return q.list.Head().Data(), nil
}
