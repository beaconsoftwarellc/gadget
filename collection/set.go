package collection

import "sync"

// Set data structure methods.
type Set interface {
	// Add the passed objects to the set
	Add(objs ...interface{}) Set
	// Remove the passed objects from the set
	Remove(objs ...interface{}) Set
	// Contains the passed object
	Contains(interface{}) bool
	// Elements in this set
	Elements() []interface{}
	// Size of this set (number of elements)
	Size() int
	// New set of the same type as this set
	New() Set
}

// Union of sets a and b as a new set of the same type as a.
func Union(a Set, b Set) Set {
	result := a.New()
	result.Add(a.Elements()...)
	result.Add(b.Elements()...)
	return result
}

// Disjunction of sets a and b as a new set of the same type as a.
func Disjunction(a Set, b Set) Set {
	return Union(a, b).Remove(Intersection(a, b).Elements()...)
}

// Intersection of sets a and b as a new set of the same type as a.
func Intersection(a Set, b Set) Set {
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

type mapSet struct {
	mutex sync.RWMutex
	// empty interface with nil stored is smaller than bool
	m map[interface{}]interface{}
}

// NewSet instance.
func NewSet(objs ...interface{}) Set {
	ms := &mapSet{m: make(map[interface{}]interface{}, len(objs))}
	return ms.Add(objs...)
}

func (ms *mapSet) New() Set {
	return NewSet()
}

func (ms *mapSet) Add(objs ...interface{}) Set {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	for _, o := range objs {
		ms.m[o] = nil
	}
	return ms
}

func (ms *mapSet) Remove(objs ...interface{}) Set {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	for _, o := range objs {
		delete(ms.m, o)
	}
	return ms
}

func (ms *mapSet) Contains(obj interface{}) bool {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	_, ok := ms.m[obj]
	return ok
}

func (ms *mapSet) Elements() []interface{} {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	i := 0
	keys := make([]interface{}, len(ms.m))
	for k := range ms.m {
		keys[i] = k
		i++
	}
	return keys
}

func (ms *mapSet) Size() int {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return len(ms.m)
}
