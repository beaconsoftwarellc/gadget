package timeutil

import (
	"sync"
	"time"
)

// Ticker sends the time, after a period, on it's channel once
// started.
type Ticker interface {
	// Start this ticker sending time on it's channel.
	Start() Ticker
	// Reset this ticker by creating a new channel and sending time on it.
	Reset()
	// Stop this ticker and release underlying resources.
	Stop()
	// SetPeriod of this ticker to the passed duration. The ticker must be restarted
	// after this change.
	SetPeriod(period time.Duration)
	// Channel that is used to send the 'ticks'
	Channel() <-chan time.Time
}

// NewTicker returns a new Ticker containing a channel that will send the
// time with a period specified by the duration argument once Started.
// It adjusts the intervals or drops ticks to make up for slow receivers.
// The duration d must be greater than zero; if not, NewTicker will panic.
// Stop the ticker to release associated resources.
func NewTicker(period time.Duration) Ticker {
	return &ticker{period: period}
}

type ticker struct {
	sync.Mutex
	ticker  *time.Ticker
	period  time.Duration
	channel <-chan time.Time
}

func (t *ticker) Start() Ticker {
	t.Lock()
	t.ticker = time.NewTicker(t.period)
	t.channel = t.ticker.C
	t.Unlock()
	return t
}

func (t *ticker) Reset() {
	t.Stop()
	t.Start()
}

func (t *ticker) SetPeriod(period time.Duration) {
	t.Lock()
	t.stop()
	t.period = period
	t.Unlock()
}

func (t *ticker) Stop() {
	t.Lock()
	t.stop()
	t.Unlock()
}

func (t *ticker) stop() {
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
		t.channel = nil
	}
}

func (t *ticker) Channel() <-chan time.Time {
	return t.channel
}
