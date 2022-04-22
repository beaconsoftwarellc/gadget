package collection

import (
	"sync"
)

// Set data structure methods.
type Set[T comparable] interface {
	// Add the passed objects to the set
	Add(objs ...T) Set[T]
	// Remove the passed objects from the set
	Remove(objs ...T) Set[T]
	// Contains the passed object
	Contains(T) bool
	// Elements in this set
	Elements() []T
	// Size of this set (number of elements)
	Size() int
	// New set of the same type as this set
	New() Set[T]
}

// Union of sets a and b as a new set of the same type as a.
func Union[T comparable](a, b Set[T]) Set[T] {
	result := a.New()
	result.Add(a.Elements()...)
	result.Add(b.Elements()...)
	return result
}

// Disjunction of sets a and b as a new set of the same type as a.
func Disjunction[T comparable](a, b Set[T]) Set[T] {
	return Union(a, b).Remove(Intersection(a, b).Elements()...)
}

// Intersection of sets a and b as a new set of the same type as a.
func Intersection[T comparable](a, b Set[T]) Set[T] {
	result := a.New()
	for _, e := range b.Elements() {
		if a.Contains(e) {
			result.Add(e)
		}
	}
	for _, e := range a.Elements() {
		if b.Contains(e) {
			result.Add(e)
		}
	}
	return result
}

type mapSet[T comparable] struct {
	mutex sync.RWMutex
	// empty interface with nil stored is smaller than bool
	m map[T]interface{}
}

// NewSet instance.
func NewSet[T comparable](objs ...T) Set[T] {
	ms := &mapSet[T]{m: make(map[T]interface{}, len(objs))}
	return ms.Add(objs...)
}

func (ms *mapSet[T]) New() Set[T] {
	return NewSet[T]()
}

func (ms *mapSet[T]) Add(objs ...T) Set[T] {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	for _, o := range objs {
		ms.m[o] = nil
	}
	return ms
}

func (ms *mapSet[T]) Remove(objs ...T) Set[T] {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	for _, o := range objs {
		delete(ms.m, o)
	}
	return ms
}

func (ms *mapSet[T]) Contains(obj T) bool {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	_, ok := ms.m[obj]
	return ok
}

func (ms *mapSet[T]) Elements() []T {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	i := 0
	keys := make([]T, len(ms.m))
	for k := range ms.m {
		keys[i] = k
		i++
	}
	return keys
}

func (ms *mapSet[T]) Size() int {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return len(ms.m)
}
