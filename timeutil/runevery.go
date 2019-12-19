package timeutil

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	running int32 = 1 + iota
	stopped
)

// RunStop exposes a functions for running, stopping, and querying status of
// presumably a background job.
type RunStop interface {
	Run()
	Running() bool
	Stop()
}

type runEvery struct {
	mutex  sync.Mutex
	f      func()
	wait   time.Duration
	stop   chan bool
	status int32
}

func (re *runEvery) Run() {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	if re.Running() {
		return
	}
	go func() {
		timer := NewTicker(re.wait).Start()
		for {
			select {
			case <-timer.Channel():
				re.f()
				timer.Reset()
			case <-re.stop:
				timer.Stop()
				return
			}
		}
	}()
	atomic.StoreInt32(&re.status, running)
}

func (re *runEvery) Running() bool {
	return atomic.LoadInt32(&re.status) == running
}

func (re *runEvery) Stop() {
	select {
	case re.stop <- true:
		break
	default:
		break
	}
}

// RunEvery wait duration. Use returned channel to stop.
func RunEvery(f func(), wait time.Duration) RunStop {
	s := &runEvery{
		f:    f,
		wait: wait,
		stop: make(chan bool),
	}
	s.Run()
	return s
}
