package net

import (
	"net"
	"sync"
	"time"

	"github.com/beaconsoftwarellc/gadget/dispatcher"
	"github.com/beaconsoftwarellc/gadget/log"
	"github.com/beaconsoftwarellc/gadget/timeutil"
)

// Sensible Defaults
const (
	DefaultMinWorkers    = 300
	DefaultBufferedTasks = 10000
	DefaultIdleTimeout   = 24 * time.Hour
)

// TCPServer serves as a base struct for all TCP service instances that listen on a Port.
type TCPServer struct {
	MaxConsecutiveErrors int
	MaxWorkers           int
	mutex                sync.Mutex
	implementation       GetListenerGetTask
	Dispatcher           dispatcher.Dispatcher
	idleUpdate           chan time.Duration
	idleTimeout          time.Duration
	idleTicker           timeutil.Ticker
	idleMutex            sync.RWMutex
	onIdle               func()
}

// NewTCPServer that will exit listen on the number of MaxConsecutiveErrors specified and use the number of
// workers specified to asynchronously process incoming connections.
func NewTCPServer(maxConsecutiveErrors, maxWorkers int, getListenerGetTask GetListenerGetTask) *TCPServer {
	return &TCPServer{
		MaxConsecutiveErrors: maxConsecutiveErrors,
		MaxWorkers:           maxWorkers,
		implementation:       getListenerGetTask,
		Dispatcher: dispatcher.NewDispatcher(DefaultBufferedTasks,
			DefaultMinWorkers, maxWorkers),
		idleUpdate:  make(chan time.Duration, 5),
		idleTimeout: DefaultIdleTimeout,
		idleTicker:  timeutil.NewTicker(DefaultIdleTimeout),
	}
}

// GetListenerGetTask provides functions for getting a listener and a task from a connection.
type GetListenerGetTask interface {
	// GetListener for accepting connections.
	GetListener() (net.Listener, error)
	// GetTask to be executed in response to an inbound connection.
	GetTask(conn net.Conn) (dispatcher.Task, error)
}

// SetIdleTimeout and on idle handler for this server
func (server *TCPServer) SetIdleTimeout(timeout time.Duration, onIdle func()) {
	server.idleUpdate <- timeout
	server.idleMutex.Lock()
	defer server.idleMutex.Unlock()
	server.onIdle = onIdle
}

func (server *TCPServer) callOnIdle() {
	server.idleMutex.RLock()
	defer server.idleMutex.RUnlock()
	server.onIdle()
}

func (server *TCPServer) listen(listener net.Listener, connections chan net.Conn, errors chan error, done chan bool) {
	for {
		select {
		// this will cause the listener to quit after the next accept or when the listener is closed, but only if
		// we are not blocking on accept, which should timeout after .
		case <-done:
			return
		default:
			conn, err := listener.Accept()
			if nil != err {
				errors <- err
			} else {
				connections <- conn
			}
		}
	}
}

// Dispatch a task using this TCPServer's worker pool.
func (server *TCPServer) Dispatch(task dispatcher.Task) {
	server.Dispatcher.Dispatch(task)
}

// Listen for incoming connection, wrap them using GetTask and execute them asynchronously.
func (server *TCPServer) Listen() (chan bool, error) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	done := make(chan bool)
	// give errors a buffer
	errors := make(chan error, 10)
	connections := make(chan net.Conn)
	consecutiveFailures := 0
	listener, err := server.implementation.GetListener()
	if nil != err {
		return nil, err
	}
	server.Dispatcher.Run()
	go func() {
		go server.listen(listener, connections, errors, done)
		server.idleTicker = timeutil.NewTicker(server.idleTimeout).Start()
		for {
			server.idleTicker.Reset()
			select {
			case timeout := <-server.idleUpdate:
				server.idleTimeout = timeout
				server.idleTicker.SetPeriod(server.idleTimeout)
			case <-server.idleTicker.Channel():
				log.Infof("tcp server was idle for %s", server.idleTimeout)
				server.callOnIdle()
				server.idleTicker.Stop()
			case <-done:
				server.Dispatcher.Quit(true)
				listener.Close()
				return
			case conn := <-connections:
				task, err := server.implementation.GetTask(conn)
				if nil != err {
					errors <- err
				} else {
					consecutiveFailures = 0
					server.Dispatch(task)
				}
			case err := <-errors:
				log.Errorf("error encountered listening %s %#v", err, err)
				consecutiveFailures++
				if consecutiveFailures > server.MaxConsecutiveErrors {
					log.Infof("Maximum consecutive errors threshold exceeded.")
					done <- true
				}
			}
		}
	}()
	return done, nil
}
