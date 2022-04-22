package specialized

import (
	"github.com/beaconsoftwarellc/gadget/v2/collection"
)

// HashPriority exposes both a priority function and Hash function.
type HashPriority[T comparable] interface {
	Priority
	GetHash() T
}

// HashPriorityQueue prevents duplicate prioritized entries in a queue.
type HashPriorityQueue[T comparable] interface {
	// Size of this queue
	Size() int
	// Push the passed element onto this queue if it does not already exist in
	// the queue.
	Push(element HashPriority[T])
	// Pop the highest priority element off the queue
	Pop() (HashPriority[T], bool)
	// Peek at the highest priority element without modifying the queue
	Peek() (HashPriority[T], bool)
}

// NewHashPriorityQueue for queueing unique elements by priority.
func NewHashPriorityQueue[T comparable]() HashPriorityQueue[T] {
	return &hashPriorityQueue[T]{
		set:    collection.NewSet[T](),
		pqueue: NewPriorityQueue(),
	}
}

type hashPriorityWrapper[T comparable] struct {
	priority int
	element  HashPriority[T]
}

func (hpw *hashPriorityWrapper[T]) GetPriority() int {
	return hpw.priority
}

func (hpw *hashPriorityWrapper[T]) GetHash() T {
	return hpw.element.GetHash()
}

func newHashPriorityWrapper[T comparable](element HashPriority[T]) *hashPriorityWrapper[T] {
	return &hashPriorityWrapper[T]{
		priority: element.GetPriority(),
		element:  element,
	}
}

type hashPriorityQueue[T comparable] struct {
	set    collection.Set[T]
	pqueue PriorityQueue
}

func (hpq *hashPriorityQueue[T]) Size() int {
	return hpq.pqueue.Size()
}

func (hpq *hashPriorityQueue[T]) Push(element HashPriority[T]) {
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

func (hpq *hashPriorityQueue[T]) convert(data interface{}) HashPriority[T] {
	p, _ := data.(HashPriority[T])
	return p
}

func (hpq *hashPriorityQueue[T]) nextElement(next func() (Priority, bool), remove bool) (HashPriority[T], bool) {
	p, ok := next()
	var hp *hashPriorityWrapper[T]
	if !ok {
		return nil, false
	}

	hp, _ = p.(*hashPriorityWrapper[T])
	if remove {
		hpq.set.Remove(hp.GetHash())
	}
	return hp.element, true
}

func (hpq *hashPriorityQueue[T]) Pop() (HashPriority[T], bool) {
	return hpq.nextElement(hpq.pqueue.Pop, true)
}

func (hpq *hashPriorityQueue[T]) Peek() (HashPriority[T], bool) {
	return hpq.nextElement(hpq.pqueue.Peek, false)
}
