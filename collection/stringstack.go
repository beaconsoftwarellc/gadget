package collection

// StringStack is a stack for storing strings.
type StringStack interface {
	// Size of the stack represented as a count of the elements in the stack.
	Size() int
	// Push a new data element onto the stack.
	Push(data string)
	// Pop the most recently pushed data element off the stack.
	Pop() (string, error)
	// Peek returns the most recently pushed element without modifying the stack
	Peek() (string, error)
}

type stringStack struct {
	stack Stack
}

// NewStringStack that is empty and ready to use.
func NewStringStack() StringStack {
	return &stringStack{stack: NewStack()}
}

// NewStringStackFromStack allows for specifying the base of the string stack.
func NewStringStackFromStack(stack Stack) StringStack {
	return &stringStack{stack: stack}
}

func (s *stringStack) Size() int {
	return s.stack.Size()
}

func (s *stringStack) Push(data string) {
	s.stack.Push(data)
}

func (s *stringStack) Pop() (string, error) {
	return convert(s.stack.Pop)
}

func (s *stringStack) Peek() (string, error) {
	return convert(s.stack.Peek)
}

func convert(call func() (interface{}, error)) (string, error) {
	var data string
	i, err := call()
	if nil == err {
		data = i.(string)
	}
	return data, err
}
