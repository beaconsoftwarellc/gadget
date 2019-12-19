package collection

// StringSet is self explanatory
type StringSet interface {
	// Add the passed objects to the set
	Add(objs ...string) StringSet
	// Remove the passed objects from the set
	Remove(objs ...string) StringSet
	// Contains the passed object
	Contains(string) bool
	// Elements in this set
	Elements() []string
	// Size of this set (number of elements)
	Size() int
	// New set of the same type as this set
	New() StringSet
}

// can't use stringutil.anonymize here because stringutil uses collections
func anonymize(strs ...string) []interface{} {
	objs := make([]interface{}, len(strs))
	for i, s := range strs {
		objs[i] = s
	}
	return objs
}

// NewStringSet containing the passed strings.
func NewStringSet(strs ...string) StringSet {
	return &stringSet{set: NewSet(anonymize(strs...)...)}
}

type stringSet struct {
	set Set
}

func (strSet *stringSet) Add(objs ...string) StringSet {
	strSet.set = strSet.set.Add(anonymize(objs...)...)
	return strSet
}

func (strSet *stringSet) Remove(objs ...string) StringSet {
	strSet.set = strSet.set.Remove(anonymize(objs...)...)
	return strSet
}

func (strSet *stringSet) Contains(s string) bool {
	return strSet.set.Contains(s)
}

func (strSet *stringSet) Elements() []string {
	elements := strSet.set.Elements()
	sa := make([]string, len(elements))
	for i, obj := range elements {
		s, _ := obj.(string)
		sa[i] = s
	}
	return sa
}

func (strSet *stringSet) Size() int {
	return strSet.set.Size()
}

func (strSet *stringSet) New() StringSet {
	return NewStringSet()
}
