package collection

// Stack is an implementation of a stack (filo/lifo) datastructure
type Stack interface {
	// Size of the stack represented as a count of the elements in the stack.
	Size() int
	// Push a new data element onto the stack.
	Push(data interface{})
	// Pop the most recently pushed data element off the stack.
	Pop() (interface{}, error)
	// Peek returns the most recently pushed element without modifying the stack
	Peek() (interface{}, error)
}

// stack is a threadsafe implementation of a stack datastructure
type stack struct {
	list List
}

// NewStack that is empty.
func NewStack() Stack {
	return &stack{list: NewList()}
}

// Size of the stack represented as a count of the elements in the stack.
func (s *stack) Size() int { return s.list.Size() }

// Push a new data element onto the stack.
func (s *stack) Push(data interface{}) {
	s.list.InsertNext(nil, data)
}

// Pop the most recently pushed data element off the stack.
func (s *stack) Pop() (data interface{}, err error) {
	return s.list.RemoveNext(nil)
}

// Peek returns the most recently pushed element without modifying the stack
func (s stack) Peek() (interface{}, error) {
	if s.list.Size() == 0 {
		return nil, NewEmptyListError()
	}
	return s.list.Head().Data(), nil
}
