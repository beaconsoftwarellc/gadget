package timeutil

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/log"
)

func Test_RunEvery(t *testing.T) {
	assert := assert.New(t)
	var value int32
	ptr := &value
	f := func() {
		log.Errorf("foo")
		value = atomic.AddInt32(ptr, 1)
	}
	re := RunEvery(f, 10*time.Millisecond)
	re.Run()
	time.Sleep(100 * time.Millisecond)
	re.Stop()
	actual := atomic.LoadInt32(ptr)
	// make sure it ran
	assert.True(actual > 0)
	time.Sleep(100 * time.Millisecond)
	// make sure it stopped
	assert.Equal(actual, atomic.LoadInt32(ptr))
}
