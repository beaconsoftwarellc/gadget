package net

import (
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/dispatcher"
	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/generator"
)

type MockGetListenerGetTask struct {
	listener         *MockListener
	task             dispatcher.Task
	getTaskError     error
	getListenerError error
}

// GetListener for accepting connections.
func (mglgt *MockGetListenerGetTask) GetListener() (net.Listener, error) {
	return mglgt.listener, mglgt.getListenerError
}

// GetTask to be executed in response to an inbound connection.
func (mglgt *MockGetListenerGetTask) GetTask(conn net.Conn) (dispatcher.Task, error) {
	return mglgt.task, mglgt.getTaskError
}

type MockTask struct {
	err error
}

func (mt *MockTask) Execute() error {
	return mt.err
}

func TestNewTCPServer(t *testing.T) {
	assert := assert.New(t)
	mglgt := &MockGetListenerGetTask{}
	server := NewTCPServer(2, 10, mglgt)
	assert.NotNil(server)
}

func TestListen(t *testing.T) {
	assert := assert.New(t)
	mglgt := &MockGetListenerGetTask{}
	atLeastOnce := false
	mglgt.task = &MockTask{}
	mglgt.listener = &MockListener{
		GetConnection: func() *MockConn {
			atLeastOnce = true
			connection := &MockConn{}
			return connection
		},
	}
	server := NewTCPServer(2, 10, mglgt)
	done, err := server.Listen()
	if assert.NoError(err) {
		time.Sleep(10 * time.Millisecond)
		assert.True(atLeastOnce)
		done <- true
	}
}

func TestListenFailsAfterMaxErrors(t *testing.T) {
	assert := assert.New(t)
	mglgt := &MockGetListenerGetTask{}
	mglgt.task = &MockTask{}
	mglgt.listener = &MockListener{
		GetConnection: func() *MockConn {
			return nil
		},
		AcceptError: errors.New(generator.String(20)),
	}
	server := NewTCPServer(2, 10, mglgt)
	done, err := server.Listen()
	if assert.NoError(err) {
		<-done
	}
}

func TestListenFailsAfterMaxErrorsCreateTask(t *testing.T) {
	assert := assert.New(t)
	mglgt := &MockGetListenerGetTask{}
	mglgt.task = &MockTask{}
	mglgt.listener = &MockListener{
		GetConnection: func() *MockConn {
			return &MockConn{}
		},
	}
	mglgt.getTaskError = errors.New(generator.String(20))
	server := NewTCPServer(2, 10, mglgt)
	done, err := server.Listen()
	if assert.NoError(err) {
		<-done
	}
}

func Test_server_OnIdle(t *testing.T) {
	assert := assert.New(t)
	mglgt := &MockGetListenerGetTask{}
	mglgt.task = &MockTask{}
	mglgt.listener = &MockListener{
		GetConnection: func() *MockConn {
			time.Sleep(100 * time.Millisecond)
			return &MockConn{}
		},
	}
	server := NewTCPServer(2, 10, mglgt)
	done, err := server.Listen()
	assert.NoError(err)
	idleCalled := int32(0)
	onIdle := func() {
		atomic.StoreInt32(&idleCalled, 1)
	}
	server.SetIdleTimeout(time.Millisecond, onIdle)
	time.Sleep(10 * time.Millisecond)
	done <- true
	assert.Equal(int32(1), atomic.LoadInt32(&idleCalled))
}
