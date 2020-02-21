package collection

// StringQueue is a queue implementation that can only store strings.
type StringQueue interface {
	// Size of the queue represented as a count of the elements in the queue.
	Size() int
	// Push a new data element onto the queue.
	Push(data string)
	// Pop the most recently pushed data element off the queue.
	Pop() (string, error)
	// Peek returns the most recently pushed element without modifying the queue
	Peek() (string, error)
}

type stringQueue struct {
	list List
}

// NewQueue that is empty.
func NewStringQueue() StringQueue {
	return &stringQueue{list: NewList()}
}

func (q *stringQueue) Size() int {
	return q.list.Size()
}

func (q *stringQueue) Push(data string) {
	q.list.InsertNext(q.list.Tail(), data)
}

func (q *stringQueue) Pop() (string, error) {
	element, err := q.list.RemoveNext(nil)
	if nil != err {
		return "", err
	}
	s := element.(string)
	return s, nil
}

func (q *stringQueue) Peek() (string, error) {
	if q.list.Size() == 0 {
		return "", NewEmptyListError()
	}
	element := q.list.Head().Data()
	s := element.(string)
	return s, nil
}