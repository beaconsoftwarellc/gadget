package collection

// Indexable provides the methods necessary for using an object in the indexing structures.
type Indexable interface {
	GetID() string
	GetField(string) interface{}
}

// Index for a field on an Indexable
type Index interface {
	// Add the passed indexable to the index
	Add(obj Indexable)
	// Update the passed indexable in the index
	Update(obj Indexable)
	// Remove the passed indexable from the index (removed by ID)
	Remove(obj Indexable)
	// LookupValue in the index and return the ID's that correspond to it
	LookupValue(fieldValue interface{}) []string
}

type index struct {
	name        string
	valuesToIDs map[interface{}]Set
}

// NewIndex for the passed field name, objects added to this index will be
// indexed by the value in this field.
func NewIndex(fieldName string) Index {
	return &index{name: fieldName, valuesToIDs: make(map[interface{}]Set)}
}

func (idx *index) Add(obj Indexable) {
	fieldValue := obj.GetField(idx.name)
	ids, ok := idx.valuesToIDs[fieldValue]
	if !ok {
		ids = NewSet(obj.GetID())
	} else {
		ids.Add(obj.GetID())
	}
	idx.valuesToIDs[fieldValue] = ids
}

func (idx *index) Update(obj Indexable) {
	fieldValue := obj.GetField(idx.name)
	id := obj.GetID()
	s, ok := idx.valuesToIDs[fieldValue]
	if !ok {
		idx.Add(obj)
	}
	if !ok || !s.Contains(id) {
		for fv, set := range idx.valuesToIDs {
			if fv == fieldValue {
				set.Add(id)
			} else {
				set.Remove(id)
			}
		}
	}
}

func (idx *index) Remove(obj Indexable) {
	id := obj.GetID()
	for _, set := range idx.valuesToIDs {
		set.Remove(id)
	}
}

func stringify(objs []interface{}) []string {
	i := 0
	strs := make([]string, len(objs))
	for _, o := range objs {
		str, ok := o.(string)
		if ok {
			strs[i] = str
		} else {
			strs[i] = ""
		}
		i++
	}
	return strs
}

func (idx *index) LookupValue(fieldValue interface{}) []string {
	set, ok := idx.valuesToIDs[fieldValue]
	var v []string
	if ok {
		v = stringify(set.Elements())
	} else {
		v = make([]string, 0)
	}
	return v
}

// indexer of arbitrary fields on Indexable structs.
type indexer struct {
	objects map[string]Indexable
	indices map[string]Index
}

// Indexer exposes methods that allow for indexing arbitrary fields on an object collection.
type Indexer interface {
	// Index the field name for all objects added to this indexer
	Index(fieldName string)
	// Contains the passed id
	Contains(id string) bool
	// Add the passed indexable to all the indices
	Add(obj Indexable)
	// Remove the passed indexable from all indices
	Remove(obj Indexable)
	// Values in the indexes that have the fieldName assigned the fieldValue, returns
	// false if there is no index supporting the passed field name.
	Values(fieldName string, fieldValue interface{}) ([]Indexable, bool)
	// Get the indexable for the passed ID form this indexer.
	Get(id string) (Indexable, bool)
	// Count the number of distinct records in this indexer
	Count() int
	// Iterate all values in this indexer
	Iterate() []Indexable
}

// NewIndexer empty indexer.
func NewIndexer() Indexer {
	return &indexer{objects: make(map[string]Indexable), indices: make(map[string]Index)}
}

// Index for the passed field name.
func (indxr *indexer) Index(fieldName string) {
	_, ok := indxr.indices[fieldName]
	if !ok {
		index := NewIndex(fieldName)
		indxr.indices[fieldName] = index
		for _, obj := range indxr.objects {
			index.Add(obj)
		}
	}
}

// Iterate all values in this indexer
func (indxr *indexer) Iterate() []Indexable {
	objects := make([]Indexable, len(indxr.objects))
	i := 0
	for _, v := range indxr.objects {
		objects[i] = v
		i++
	}
	return objects
}

// Add the indexable object to this Indexers's indices.
func (indxr *indexer) Add(obj Indexable) {
	_, exists := indxr.objects[obj.GetID()]
	indxr.objects[obj.GetID()] = obj
	for _, index := range indxr.indices {
		if exists {
			index.Update(obj)
		} else {
			index.Add(obj)
		}
	}
}

// Remove the indexable object from this Indexers's indices.
func (indxr *indexer) Remove(obj Indexable) {
	id := obj.GetID()
	_, exists := indxr.objects[id]
	delete(indxr.objects, id)
	if exists {
		for _, index := range indxr.indices {
			index.Remove(obj)
		}
	}
}

func (indxr *indexer) Contains(id string) bool {
	_, ok := indxr.objects[id]
	return ok
}

// Get the passed indexable by id.
func (indxr *indexer) Get(id string) (Indexable, bool) {
	obj, ok := indxr.objects[id]
	return obj, ok
}

func (indxr *indexer) Count() int {
	return len(indxr.objects)
}

// Values in the index for the passed field name with the passed field value.
func (indxr *indexer) Values(fieldName string, fieldValue interface{}) ([]Indexable, bool) {
	index, ok := indxr.indices[fieldName]
	var values []Indexable
	if ok {
		ids := index.LookupValue(fieldValue)
		values = make([]Indexable, len(ids))
		for i, id := range ids {
			values[i] = indxr.objects[id]
		}
	}
	return values, ok
}
