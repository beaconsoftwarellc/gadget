package messagequeue

// EnqueueMessageResult is returned on for each message that is enqueued
type EnqueueMessageResult struct {
	// Message this result is for
	*Message
	// Success indicates whether the message was successfully enqueued
	Success bool
	// SenderFault when success is false, indicates that the enqueue failed due
	// to a malformed message
	SenderFault bool
	// Error that occurred when enqueueing the message
	Error string
}
