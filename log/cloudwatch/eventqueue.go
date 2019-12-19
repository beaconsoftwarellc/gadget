package cloudwatch

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/beaconsoftwarellc/gadget/collection"
)

// EventQueue for buffering cloud watch events.
type EventQueue interface {
	// Size of this queue
	Size() int
	// Push the passed event onto this queue
	Push(event *cloudwatchlogs.InputLogEvent)
	// Pop the last event pushed off this queue
	Pop() (*cloudwatchlogs.InputLogEvent, error)
	// Peek at the next event that will be popped.
	Peek() (*cloudwatchlogs.InputLogEvent, error)
}

type eventQueue struct {
	queue collection.Queue
}

// NewEventQueue that is empty.
func NewEventQueue() EventQueue {
	return &eventQueue{queue: collection.NewQueue()}
}

func (eq *eventQueue) Size() int {
	return eq.queue.Size()
}

func (eq *eventQueue) Push(event *cloudwatchlogs.InputLogEvent) {
	eq.queue.Push(event)
}

func (eq *eventQueue) Pop() (*cloudwatchlogs.InputLogEvent, error) {
	obj, err := eq.queue.Pop()
	if nil != err {
		return nil, err
	}
	event, _ := obj.(*cloudwatchlogs.InputLogEvent)
	return event, nil
}

func (eq *eventQueue) Peek() (*cloudwatchlogs.InputLogEvent, error) {
	obj, err := eq.queue.Peek()
	if nil != err {
		return nil, err
	}
	event, _ := obj.(*cloudwatchlogs.InputLogEvent)
	return event, nil
}
