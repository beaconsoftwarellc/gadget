package specialized

import (
	"github.com/beaconsoftwarellc/gadget/collection"
)

// HashPriority exposes both a priority function and Hash function.
type HashPriority interface {
	Priority
	GetHash() interface{}
}

// HashPriorityQueue prevents duplicate prioritized entries in a queue.
type HashPriorityQueue interface {
	// Size of this queue
	Size() int
	// Push the passed element onto this queue if it does not already exist in
	// the queue.
	Push(element HashPriority)
	// Pop the highest priority element off the queue
	Pop() (HashPriority, bool)
	// Peek at the highest priority element without modifying the queue
	Peek() (HashPriority, bool)
}

// NewHashPriorityQueue for queueing unique elements by priority.
func NewHashPriorityQueue() HashPriorityQueue {
	return &hashPriorityQueue{
		set:    collection.NewSet(),
		pqueue: NewPriorityQueue(),
	}
}

type hashPriorityWrapper struct {
	priority int
	element  HashPriority
}

func (hpw *hashPriorityWrapper) GetPriority() int {
	return hpw.priority
}

func (hpw *hashPriorityWrapper) GetHash() interface{} {
	return hpw.element.GetHash()
}

func newHashPriorityWrapper(element HashPriority) *hashPriorityWrapper {
	return &hashPriorityWrapper{
		priority: element.GetPriority(),
		element:  element,
	}
}

type hashPriorityQueue struct {
	set    collection.Set
	pqueue PriorityQueue
}

func (hpq *hashPriorityQueue) Size() int {
	return hpq.pqueue.Size()
}

func (hpq *hashPriorityQueue) Push(element HashPriority) {
	hash := element.GetHash()
	wrappedElement := newHashPriorityWrapper(element)
	if hpq.set.Contains(hash) {
		for elm := hpq.pqueue.(*priorityQueue).list.Head(); elm != nil; elm = elm.Next() {
			d := hpq.convert(elm.Data())
			if d.GetHash() == hash {
				hpq.pqueue.(*priorityQueue).list.Remove(elm)
				wrappedElement.priority = d.GetPriority()
			}
		}
	}
	hpq.pqueue.Push(wrappedElement)
	hpq.set.Add(hash)
}

func (hpq *hashPriorityQueue) convert(data interface{}) HashPriority {
	p, _ := data.(HashPriority)
	return p
}

func (hpq *hashPriorityQueue) nextElement(next func() (Priority, bool), remove bool) (HashPriority, bool) {
	p, ok := next()
	var hp *hashPriorityWrapper
	if !ok {
		return nil, false
	}

	hp, _ = p.(*hashPriorityWrapper)
	if remove {
		hpq.set.Remove(hp.GetHash())
	}
	return hp.element, true
}

func (hpq *hashPriorityQueue) Pop() (HashPriority, bool) {
	return hpq.nextElement(hpq.pqueue.Pop, true)
}

func (hpq *hashPriorityQueue) Peek() (HashPriority, bool) {
	return hpq.nextElement(hpq.pqueue.Peek, false)
}
