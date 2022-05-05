package collection

// GetPivotKeys type Y for the object T
type GetPivotKeys[T, Y comparable] func(obj T) []Y

// Pivot for a field comparable field of type Y that is a member of objects of
// type T
type Pivot[T, Y comparable] interface {
	// Add the passed comparables to the pivot
	Add(...T)
	// Get comparables who share a pivot key
	Get(Y) []T
	// Remove the passed comparable from the index
	Remove(T)
	// Len is the cardinality of the set of keys Y
	Len() int
}

type pivot[T, Y comparable] struct {
	getter GetPivotKeys[T, Y]
	index  map[Y]Set[T]
}

// NewPivot for the field of type Y that is a member of type T
func NewPivot[T, Y comparable](getter GetPivotKeys[T, Y], init ...T) Pivot[T, Y] {
	p := &pivot[T, Y]{
		getter: getter,
		index:  make(map[Y]Set[T]),
	}
	p.Add(init...)
	return p
}

func (p *pivot[T, Y]) Get(value Y) []T {
	values, ok := p.index[value]
	if !ok {
		values = NewSet[T]()
	}
	return values.Elements()
}

func (p *pivot[T, Y]) add(obj T) {
	for _, value := range p.getter(obj) {
		values, ok := p.index[value]
		if ok {
			values.Add(obj)
		} else {
			p.index[value] = NewSet(obj)
		}
	}
}

func (p *pivot[T, Y]) Add(objs ...T) {
	for _, obj := range objs {
		p.add(obj)
	}
}

func (p *pivot[T, Y]) Remove(obj T) {
	for _, value := range p.getter(obj) {
		objs, ok := p.index[value]
		if !ok {
			continue
		}
		objs.Remove(obj)
		if objs.Size() == 0 {
			delete(p.index, value)
		}
	}
}

func (p *pivot[T, Y]) Len() int {
	return len(p.index)
}
