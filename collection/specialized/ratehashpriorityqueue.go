package specialized

import (
	"sync/atomic"
	"time"

	"github.com/beaconsoftwarellc/gadget/timeutil"
)

// RateHashPriorityQueue allows prioritized unique elements to be emitted
// at a specific rate.
type RateHashPriorityQueue interface {
	HashPriorityQueue
	// NoLimitPop the highest priority element off the queue ignoring the rate limit
	// for the purpose of batching commands
	NoLimitPop() (HashPriority, bool)
	// Channel that can be used instead of Pop
	Channel() <-chan HashPriority
	// Stop this RateHashPriorityQueue so that it can be garbage collected.
	Stop()
}

// NewRateHashPriorityQueue that will return at max 'n' elements per 'rate' duration.
func NewRateHashPriorityQueue(n int, rate time.Duration) RateHashPriorityQueue {
	q := &rhpQueue{
		rate:    rate,
		queue:   NewHashPriorityQueue(),
		channel: make(chan HashPriority, n),
		stop:    make(chan bool),
	}
	go q.run()
	return q
}

// rhpQueue implementes the RateHashPriorityQueue interface
type rhpQueue struct {
	rate    time.Duration
	queue   HashPriorityQueue
	size    int32
	channel chan HashPriority
	stop    chan bool
}

func (q *rhpQueue) run() {
	ticker := timeutil.NewTicker(q.rate).Start()
	for {
		select {
		case <-ticker.Channel():
			if elm, ok := q.queue.Pop(); ok {
				// if this blocks either the rate has been reached
				// or no one is listening. either way we want to block this thread
				// on it
				q.channel <- elm
				atomic.AddInt32(&q.size, -1)
			}
			ticker.Reset()
		case <-q.stop:
			ticker.Stop()
			close(q.channel)
			return
		}
	}
}

func (q *rhpQueue) Size() int {
	return int(atomic.LoadInt32(&q.size))
}

func (q *rhpQueue) Push(element HashPriority) {
	q.queue.Push(element)
	atomic.StoreInt32(&q.size, int32(q.queue.Size()))
}

func (q *rhpQueue) Pop() (HashPriority, bool) {
	return <-q.channel, true
}

func (q *rhpQueue) Channel() <-chan HashPriority {
	return q.channel
}

func (q *rhpQueue) Peek() (HashPriority, bool) {
	return q.queue.Peek()
}

func (q *rhpQueue) NoLimitPop() (HashPriority, bool) {
	select {
	case elm := <-q.channel:
		return elm, true
	default:
		if elm, ok := q.queue.Pop(); ok {
			atomic.StoreInt32(&q.size, int32(q.queue.Size()))
			return elm, true
		}
		return nil, false
	}
}

func (q *rhpQueue) Stop() {
	// non-blocking so that this is reentrant
	select {
	case q.stop <- true:
		return
	default:
		return
	}
}
