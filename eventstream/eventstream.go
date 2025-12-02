package eventstream

type Event interface {
	GetProject() string  // this should be static and with Resource identifies the stream
	GetResource() string // this should be static and with Action identifies the stream
	GetAction() string
	GetPayload() []byte
}

type EventStream[T Event] interface {
	// Get records from the underlying Kinesis Stream
	Get() <-chan T
	// Put an event record into the underlying Kinesis Stream
	Put(T) error
}
