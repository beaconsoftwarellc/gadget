package messagequeue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/golang/mock/gomock"
	assert1 "github.com/stretchr/testify/assert"
)

type waitQueue struct {
	dq       []*Message
	notFirst bool
	deleted  *Message
}

func (wq *waitQueue) EnqueueBatch(context.Context, []*Message) ([]*EnqueueMessageResult, error) {
	return nil, nil
}

func (wq *waitQueue) Dequeue(ctx context.Context, count int, wait time.Duration) ([]*Message, error) {
	if !wq.notFirst {
		wq.notFirst = true
		return wq.dq, nil
	}
	<-ctx.Done()
	return nil, nil
}

func (wq *waitQueue) Delete(ctx context.Context, m *Message) error {
	wq.deleted = m
	return nil
}

func TestNewPoller(t *testing.T) {
	assert := assert1.New(t)
	actual := NewPoller(nil)
	assert.NotNil(actual)
}

func TestPoller_Poll_Validation(t *testing.T) {
	assert := assert1.New(t)
	poller := NewPoller(nil)
	handler := func(context.Context, *Message) bool { return false }
	messageQueue := &waitQueue{}
	assert.EqualError(poller.Poll(nil, messageQueue), "handler cannot be nil")
	assert.EqualError(poller.Poll(handler, nil), "messageQueue cannot be nil")
	assert.NoError(poller.Poll(handler, messageQueue))
	assert.EqualError(poller.Poll(handler, messageQueue),
		"Poller.Poll called on instance not in state stopped (0)")
	assert.NoError(poller.Stop())
}

func TestPoller_Poll(t *testing.T) {
	assert := assert1.New(t)
	controller := gomock.NewController(t)
	messageQueue := NewMockMessageQueue(controller)
	successfulMessage := &Message{ID: generator.String(5)}
	unsuccessfulMessage := &Message{ID: generator.String(10)}
	wg := sync.WaitGroup{}
	wg.Add(3)
	handler := func(_ context.Context, m *Message) bool {
		wg.Done()
		return m.ID == successfulMessage.ID
	}
	options := NewPollerOptions()
	// we want the first call to match
	firstCall := messageQueue.EXPECT().Dequeue(gomock.Any(),
		options.DequeueCount, options.WaitForBatch).Return(
		[]*Message{successfulMessage, unsuccessfulMessage}, nil)
	// and then just return an empty response every 100ms so that we don't turn
	// our computers into space heaters.
	// messageQueue.EXPECT().Dequeue(gomock.Any(), options.DequeueCount, options.WaitForBatch).
	// 	Return(nil, nil).After(firstCall).Do(func(interface{}, interface{}, interface{}) {
	// 	time.Sleep(100 * time.Millisecond)
	// })
	messageQueue.EXPECT().Dequeue(gomock.Any(),
		options.DequeueCount, options.WaitForBatch).Return(nil, nil).
		After(firstCall).AnyTimes()

	messageQueue.EXPECT().Delete(gomock.Any(), successfulMessage).Return(nil).Do(func(interface{}, interface{}) {
		wg.Done()
	})
	poller := NewPoller(options)
	assert.NoError(poller.Poll(handler, messageQueue))
	wg.Wait()
	assert.NoError(poller.Stop())
}

func TestPoller_Stop(t *testing.T) {

}
