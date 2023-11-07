package messagequeue

import (
	"sync"
	"testing"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	assert := assert1.New(t)

	actual := New(nil)
	assert.NotNil(actual)

	options := &EnqueuerOptions{}
	actual = New(options)
	assert.NotNil(actual)
}

func TestEnqueuerStart(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	messageQueue := NewMockMessageQueue(ctrl)

	options := &EnqueuerOptions{ChunkerOptions: &ChunkerOptions{}}
	eq := New(options)

	// fails on nil queue
	assert.EqualError(eq.Start(nil), "messageQueue cannot be nil")

	// fails on options validate
	assert.EqualError(eq.Start(messageQueue), "EnqueuerOptions.Logger cannot be nil")

	// fails on bad state
	options.Logger = log.Global()
	options.BufferSize = 2
	options.ChunkSize = 1
	options.MaxElementWait = 1 * time.Millisecond
	options.FailedBufferSize = defaultFailedBufferSize
	assert.NoError(eq.Start(messageQueue))
	assert.EqualError(eq.Start(messageQueue),
		"Enqueuer.Start called while not in state 'Stopped'")
	eq.Stop()
}

func TestEnqueueStop(t *testing.T) {
	assert := assert1.New(t)

	options := NewEnqueuerOptions()
	qr := New(options)
	// stop in bad state fails
	assert.EqualError(qr.Stop(), "Enqueuer.Stop called while not in state 'Running'")

	// test no error
	ctrl := gomock.NewController(t)
	messageQueue := NewMockMessageQueue(ctrl)
	assert.NoError(qr.Start(messageQueue))
	assert.NoError(qr.Stop())

	// test we can restart
	assert.NoError(qr.Start(messageQueue))
	assert.NoError(qr.Stop())
}

func TestEnqueuerEnqueue_Validation(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	messageQueue := NewMockMessageQueue(ctrl)
	options := NewEnqueuerOptions()
	options.BufferSize = 2
	options.ChunkSize = 1

	eq := New(options)
	// enqueue into stopped fails
	assert.EqualError(eq.Enqueue(&Message{}), "cannot enqueue into a stopped MessageQueue")

	// test enqueue error
	expected := &Message{
		Service: generator.String(5),
		Method:  generator.String(10),
	}
	expectedError := generator.String(32)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	options.FailureHandler = func(_ Enqueuer, actual *EnqueueMessageResult) {
		if assert.NotNil(actual) {
			assert.Equal(expected, actual.Message)
			assert.Equal(expectedError, actual.Error)
		}
		waitGroup.Done()
	}

	messageQueue.EXPECT().EnqueueBatch(gomock.Any(), []*Message{expected}).
		Return(nil, errors.New(expectedError))
	assert.NoError(eq.Start(messageQueue))
	assert.NoError(eq.Enqueue(expected))
	waitGroup.Wait()
	assert.NoError(eq.Stop())
}

func TestEnqueue_BatchError(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	messageQueue := NewMockMessageQueue(ctrl)
	options := NewEnqueuerOptions()
	options.BufferSize = 2
	options.ChunkSize = 1

	eq := New(options)

	expected := &Message{
		Service: generator.String(5),
		Method:  generator.String(10),
	}
	expectedError := generator.String(32)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	options.FailureHandler = func(_ Enqueuer, actual *EnqueueMessageResult) {
		if assert.NotNil(actual) {
			assert.Equal(expected, actual.Message)
			assert.Equal(expectedError, actual.Error)
		}
		waitGroup.Done()
	}

	messageQueue.EXPECT().EnqueueBatch(gomock.Any(), []*Message{expected}).
		Return(nil, errors.New(expectedError))
	assert.NoError(eq.Start(messageQueue))
	assert.NoError(eq.Enqueue(expected))
	waitGroup.Wait()
	assert.NoError(eq.Stop())
}

func TestEnqueue(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	messageQueue := NewMockMessageQueue(ctrl)
	options := NewEnqueuerOptions()
	options.BufferSize = 3
	options.ChunkSize = 2
	eq := New(options)

	expected := []*Message{
		{
			Service: generator.String(5),
			Method:  generator.String(10),
		},
		{
			Service: generator.String(5),
			Method:  generator.String(10),
		},
	}
	expectedError := generator.String(32)
	results := []*EnqueueMessageResult{
		{
			Message: expected[0],
			Success: true,
		},
		{
			Message: expected[1],
			Success: false,
			Error:   expectedError,
		},
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	options.FailureHandler = func(_ Enqueuer, actual *EnqueueMessageResult) {
		if assert.NotNil(actual) {
			assert.Equal(expected[1], actual.Message)
			assert.Equal(expectedError, actual.Error)
		}
		waitGroup.Done()
	}

	messageQueue.EXPECT().EnqueueBatch(gomock.Any(), expected).
		Return(results, nil)

	assert.NoError(eq.Start(messageQueue))
	assert.NoError(eq.Enqueue(expected[0]))
	assert.NoError(eq.Enqueue(expected[1]))
	waitGroup.Wait()
	assert.NoError(eq.Stop())
}
