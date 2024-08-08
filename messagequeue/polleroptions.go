package messagequeue

import (
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

const (
	minimumConcurrentMessageHandlers = 1
	defaultConcurrentMessageHandlers = 10
	minimumWaitForBatch              = time.Second
	maximumWaitForBatch              = time.Hour
	defaultWaitForBatch              = 30 * time.Second
	minimumDequeueCount              = 1
	maximumDequeueCount              = 10
	defaultDequeueCount              = 10
	minimumQueueOperationTimeout     = time.Second
	maximumQueueOperationTimeout     = time.Hour
	defaultQueueOperationTimeout     = 5 * time.Second
	minimumVisibilityTimeout         = 0 * time.Second
	maximumVisibilityTimeout         = 12 * time.Hour
	defaultVisibilityTimeout         = 30 * time.Second
)

type PollerOptions struct {
	// Logger to use for reporting errors
	Logger log.Logger
	// ConcurrentMessageHandlers that should be running at any given time
	ConcurrentMessageHandlers int
	// WaitForBatch the specified duration before prematurely returning with less
	// than the desired number of messages.
	WaitForBatch time.Duration
	// DequeueCount is the number of messages to attempt to dequeue per request.
	// maximum will vary by implementation
	DequeueCount int
	// QueueOperationTimeout
	QueueOperationTimeout time.Duration
	// VisibilityTimeout is the amount of time a message is hidden from other
	// consumers after it has been received by a message queue client.
	VisibilityTimeout time.Duration
}

// NewPollerOptions with valid values that can be used to initialize a new Poller
func NewPollerOptions() *PollerOptions {
	return &PollerOptions{
		Logger:                    log.Global(),
		ConcurrentMessageHandlers: defaultConcurrentMessageHandlers,
		WaitForBatch:              defaultWaitForBatch,
		QueueOperationTimeout:     defaultQueueOperationTimeout,
		DequeueCount:              defaultDequeueCount,
		VisibilityTimeout:         defaultVisibilityTimeout,
	}
}

// Validate that the values contained in this Options are complete and within the
// bounds necessary for operation.
func (po *PollerOptions) Validate() error {
	// logger should not be nil
	if po.Logger == nil {
		return errors.New("PollerOptions.Logger cannot be nil")
	}
	if po.ConcurrentMessageHandlers < minimumConcurrentMessageHandlers {
		return errors.New("PollerOptions.ConcurrentMessageHandlers(%d) was out of bounds [%d, -)",
			po.ConcurrentMessageHandlers, minimumConcurrentMessageHandlers)
	}
	if po.WaitForBatch < minimumWaitForBatch || po.WaitForBatch > maximumWaitForBatch {
		return errors.New("PollerOptions.WaitForBatch(%s) was out of bounds [%s, %s]",
			po.WaitForBatch, minimumWaitForBatch, maximumWaitForBatch)
	}
	if po.QueueOperationTimeout < minimumQueueOperationTimeout ||
		po.QueueOperationTimeout > maximumQueueOperationTimeout {
		return errors.New("PollerOptions.QueueOperationTimeout(%s) was out of bounds [%s, %s]",
			po.QueueOperationTimeout, minimumQueueOperationTimeout, maximumQueueOperationTimeout)
	}
	if po.DequeueCount < minimumDequeueCount || po.DequeueCount > maximumDequeueCount {
		return errors.New("PollerOptions.DequeueCount(%d) was out of bounds [%d, %d]",
			po.DequeueCount, minimumDequeueCount, maximumDequeueCount)
	}
	if po.VisibilityTimeout < minimumVisibilityTimeout || po.VisibilityTimeout > maximumVisibilityTimeout {
		return errors.New("PollerOptions.VisibilityTimeout(%s) was out of bounds [%s, %s]",
			po.VisibilityTimeout, minimumVisibilityTimeout, maximumVisibilityTimeout)
	}
	return nil
}
