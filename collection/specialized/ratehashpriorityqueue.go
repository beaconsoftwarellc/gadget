package specialized

import (
	"sync/atomic"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/timeutil"
)

// RateHashPriorityQueue allows prioritized unique elements to be emitted
// at a specific rate.
type RateHashPriorityQueue[T comparable] interface {
	HashPriorityQueue[T]
	// NoLimitPop the highest priority element off the queue ignoring the rate limit
	// for the purpose of batching commands
	NoLimitPop() (HashPriority[T], bool)
	// Channel that can be used instead of Pop
	Channel() <-chan HashPriority[T]
	// Stop this RateHashPriorityQueue so that it can be garbage collected.
	Stop()
}

// NewRateHashPriorityQueue that will return at max 'n' elements per 'rate' duration.
func NewRateHashPriorityQueue[T comparable](n int, rate time.Duration) RateHashPriorityQueue[T] {
	q := &rhpQueue[T]{
		rate:    rate,
		queue:   NewHashPriorityQueue[T](),
		channel: make(chan HashPriority[T], n),
		stop:    make(chan bool),
	}
	go q.run()
	return q
}

// rhpQueue implementes the RateHashPriorityQueue interface
type rhpQueue[T comparable] struct {
	rate    time.Duration
	queue   HashPriorityQueue[T]
	size    int32
	channel chan HashPriority[T]
	stop    chan bool
}

func (q *rhpQueue[T]) run() {
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

func (q *rhpQueue[T]) Size() int {
	return int(atomic.LoadInt32(&q.size))
}

func (q *rhpQueue[T]) Push(element HashPriority[T]) {
	q.queue.Push(element)
	atomic.StoreInt32(&q.size, int32(q.queue.Size()))
}

func (q *rhpQueue[T]) Pop() (HashPriority[T], bool) {
	return <-q.channel, true
}

func (q *rhpQueue[T]) Channel() <-chan HashPriority[T] {
	return q.channel
}

func (q *rhpQueue[T]) Peek() (HashPriority[T], bool) {
	return q.queue.Peek()
}

func (q *rhpQueue[T]) NoLimitPop() (HashPriority[T], bool) {
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

func (q *rhpQueue[T]) Stop() {
	// non-blocking so that this is reentrant
	select {
	case q.stop <- true:
		return
	default:
		return
	}
}
