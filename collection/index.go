package collection

// GetIndexedValue type Y for the indexable object T
type GetIndexedValue[T, Y comparable] func(obj T) []Y

// Index for a field comparable field of type Y that is a member of objects of
// type T
type Index[T, Y comparable] interface {
	// Add the passed comparable to the index
	Add(...T)
	// Get indexed object by indexed value
	Get(Y) []T
	// Remove the passed comparable from the index
	Remove(T)
}

type index[T, Y comparable] struct {
	getter GetIndexedValue[T, Y]
	index  map[Y]Set[T]
}

// NewIndex for the field of type Y that is a member of type T
func NewIndex[T, Y comparable](getter GetIndexedValue[T, Y]) Index[T, Y] {
	idx := &index[T, Y]{
		getter: getter,
		index:  make(map[Y]Set[T]),
	}
	return idx
}

func (idx *index[T, Y]) Get(value Y) []T {
	values, ok := idx.index[value]
	if !ok {
		values = NewSet[T]()
	}
	return values.Elements()
}

func (idx *index[T, Y]) add(obj T) {
	for _, value := range idx.getter(obj) {
		values, ok := idx.index[value]
		if ok {
			values.Add(obj)
		} else {
			idx.index[value] = NewSet(obj)
		}
	}
}

func (idx *index[T, Y]) Add(objs ...T) {
	for _, obj := range objs {
		idx.add(obj)
	}
}

func (idx *index[T, Y]) Remove(obj T) {
	for _, value := range idx.getter(obj) {
		objs, ok := idx.index[value]
		if ok {
			objs.Remove(obj)
		}
	}
}
