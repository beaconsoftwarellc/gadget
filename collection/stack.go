package collection

// Stack is an implementation of a stack (filo/lifo) datastructure
type Stack[T any] interface {
	// Size of the stack represented as a count of the elements in the stack.
	Size() int
	// Push a new data element onto the stack.
	Push(data T)
	// Pop the most recently pushed data element off the stack.
	Pop() (T, error)
	// Peek returns the most recently pushed element without modifying the stack
	Peek() (T, error)
}

// stack is a threadsafe implementation of a stack datastructure
type stack[T any] struct {
	list List[T]
}

// NewStack that is empty.
func NewStack[T any]() Stack[T] {

	return &stack[T]{list: NewList[T]()}
}

// Size of the stack represented as a count of the elements in the stack.
func (s *stack[T]) Size() int { return s.list.Size() }

// Push a new data element onto the stack.
func (s *stack[T]) Push(data T) {
	s.list.InsertNext(nil, data)
}

// Pop the most recently pushed data element off the stack.
func (s *stack[T]) Pop() (data T, err error) {
	return s.list.RemoveNext(nil)
}

// Peek returns the most recently pushed element without modifying the stack
func (s *stack[T]) Peek() (T, error) {
	if s.list.Size() == 0 {
		var ret T
		return ret, NewEmptyListError()
	}
	return s.list.Head().Data(), nil
}
