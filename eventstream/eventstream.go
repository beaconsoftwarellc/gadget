package eventstream

type Event interface {
	// GetPartitionKey used to determine which shard this Event belongs
	// to. This key is md5'd and then bucketed into the hash space of the
	// number of shards. Any events who sequencing is important should be
	// on the same shard and should use the same partition key. e.g. all
	// events for actions on a resource should have the resource id as the
	// partition key.
	//
	// PartitionsKey does NOT have to be unique per [Event].
	GetPartitionKey() string
	// GetPayload as an immutable blob of byte data.
	GetPayload() ([]byte, error)
}

type EventStream[T Event] interface {
	// Get records from the underlying Kinesis Stream
	Get() <-chan T
	// Put an event record into the underlying Kinesis Stream
	Put(T) error
}

func New[T Event](project, resource string) EventStream[T] {
	return nil
}
