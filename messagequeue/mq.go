package messagequeue

import "time"

type MessageQueue interface {
}

type Message struct {
	ID      string
	Trace   string
	Delay   time.Duration
	Service string
	Method  string
	Body    string
}

type messagequeue struct {
}

func New() (MessageQueue, error) {
	return &messagequeue{}, nil
}
