package dispatcher

import "github.com/beaconsoftwarellc/gadget/collection"

// TaskStack is a stack for storing tasks.
type TaskStack interface {
	// Size of the stack represented as a count of the elements in the stack.
	Size() int
	// Push a new data element onto the stack.
	Push(data *internalTask)
	// Pop the most recently pushed data element off the stack.
	Pop() (*internalTask, error)
	// Peek returns the most recently pushed element without modifying the stack
	Peek() (*internalTask, error)
}

type taskStack struct {
	stack collection.Stack
}

// NewTaskStack that is empty and ready to use.
func NewTaskStack() TaskStack {
	return &taskStack{stack: collection.NewStack()}
}

func (s *taskStack) Size() int {
	return s.stack.Size()
}

func (s *taskStack) Push(data *internalTask) {
	s.stack.Push(data)
}

func (s *taskStack) Pop() (*internalTask, error) {
	return convert(s.stack.Pop)
}

func (s *taskStack) Peek() (*internalTask, error) {
	return convert(s.stack.Peek)
}

func convert(call func() (interface{}, error)) (*internalTask, error) {
	var data *internalTask
	i, err := call()
	if nil == err {
		data = i.(*internalTask)
	}
	return data, err
}
