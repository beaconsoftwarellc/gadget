package specialized

import "github.com/beaconsoftwarellc/gadget/collection"

// Priority is for use in collections that require elements to resolve
// their own priority.
type Priority interface {
	// GetPriority of this element as an int where lower priority < higher priority
	GetPriority() int
}

// PriorityQueue for queueing elements that implement priority and returning
// elements in the order of highest to lowest priority.
type PriorityQueue interface {
	// Size of this queue
	Size() int
	// Push the passed element onto this queue
	Push(element Priority)
	// Pop the highest priority element off the queue
	Pop() (Priority, bool)
	// Peek at the highest priority element without modifying the queue
	Peek() (Priority, bool)
}

type priorityQueue struct {
	list collection.DList
}

// NewPriorityQueue for queueing elements according to their priority
func NewPriorityQueue() PriorityQueue {
	return &priorityQueue{list: collection.NewDList()}
}

func (q *priorityQueue) Size() int {
	return q.list.Size()
}

func (q *priorityQueue) Push(p Priority) {
	var e *collection.DListElement
	for elm := q.list.Head(); elm != nil; elm = elm.Next() {
		d := q.convert(elm.Data())
		if d.GetPriority() > p.GetPriority() {
			e = elm
		}
	}
	if nil == e {
		q.list.InsertPrevious(q.list.Head(), p)
	} else {
		q.list.InsertNext(e, p)
	}
}

func (q *priorityQueue) nextElement(remove bool) (Priority, bool) {
	var p Priority
	success := false
	if q.list.Head() != nil {
		success = true
		p = q.convert(q.list.Head().Data())
		if remove {
			q.list.Remove(q.list.Head())
		}
	}
	return p, success
}

func (q *priorityQueue) Pop() (Priority, bool) {
	return q.nextElement(true)
}

func (q *priorityQueue) Peek() (Priority, bool) {
	return q.nextElement(false)
}

func (q *priorityQueue) convert(data interface{}) Priority {
	p, _ := data.(Priority)
	return p
}
