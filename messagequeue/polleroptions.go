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
	maximumTimeoutDequeueAfter       = time.Hour
	defaultTimeoutDequeueAfter       = 40 * time.Second
	minimumDequeueCount              = 1
	maximumDequeueCount              = 10
	defaultDequeueCount              = 10
)

type PollerOptions struct {
	// Logger to use for reporting errors
	Logger log.Logger
	// ConcurrentMessageHandlers that should be running at any given time
	ConcurrentMessageHandlers int
	// WaitForBatch the specified duration before prematurely returning with less
	// than the desired number of messages.
	WaitForBatch time.Duration
	// TimeoutDequeueAfter the specified duration and consider the dequeue attempt
	// to be in an error state.
	TimeoutDequeueAfter time.Duration
	// DequeueCount is the number of messages to attempt to dequeue per request.
	// maximum will vary by implementation
	DequeueCount int
}

// NewPollerOptions with valid values that can be used to initialize a new Poller
func NewPollerOptions() *PollerOptions {
	return &PollerOptions{
		Logger:                    log.Global(),
		ConcurrentMessageHandlers: defaultConcurrentMessageHandlers,
		WaitForBatch:              defaultWaitForBatch,
		TimeoutDequeueAfter:       defaultTimeoutDequeueAfter,
		DequeueCount:              defaultDequeueCount,
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
	if po.TimeoutDequeueAfter < po.WaitForBatch || po.TimeoutDequeueAfter > maximumTimeoutDequeueAfter {
		return errors.New("PollerOptions.TimeoutDequeueAfter(%s) was out of bounds (WaitForBatch(%s), %s]",
			po.TimeoutDequeueAfter, po.WaitForBatch, maximumTimeoutDequeueAfter)
	}
	if po.DequeueCount < minimumDequeueCount || po.DequeueCount > maximumDequeueCount {
		return errors.New("PollerOptions.DequeueCount(%d) was out of bounds [%d, %d]",
			po.DequeueCount, minimumDequeueCount, maximumDequeueCount)
	}
	return nil
}
