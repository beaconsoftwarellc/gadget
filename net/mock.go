package net

import (
	"context"
	"net"
	"net/http"
	"time"

	"net/url"

	"github.com/beaconsoftwarellc/gadget/collection"
	"github.com/beaconsoftwarellc/gadget/errors"
)

// SimpleDoRequest allows for providing a function as a client.
type SimpleDoRequest struct {
	DoFunc    func(req *http.Request) (*http.Response, error)
	cookieJar map[string][]*http.Cookie
}

// Do implements DoHTTPRequest.Do
func (m *SimpleDoRequest) Do(req *http.Request) (*http.Response, errors.TracerError) {
	r, e := m.DoFunc(req)
	return r, errors.Wrap(e)
}

// DoWithContext implements DoHTTPRequest.DoWithContext
func (m *SimpleDoRequest) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, errors.TracerError) {
	return m.Do(req)
}

// AddCookieJar to http client to make them available to all future requests
func (m *SimpleDoRequest) AddCookieJar(jar http.CookieJar) {
	m.cookieJar = map[string][]*http.Cookie{}
}

// Cookies lists cookies in the jar
func (m *SimpleDoRequest) Cookies(url *url.URL) []*http.Cookie {
	urlStr := url.RawPath
	return m.cookieJar[urlStr]
}

// SetCookies adds cookies to the jar
func (m *SimpleDoRequest) SetCookies(url *url.URL, cookies []*http.Cookie) {
	urlStr := url.RawPath
	m.cookieJar[urlStr] = append(m.cookieJar[urlStr], cookies...)
}

// MockHTTPClient mocks the DoHTTPRequest interface
type MockHTTPClient struct {
	DoReturn  collection.Stack
	DoCalled  collection.Stack
	cookieJar map[string][]*http.Cookie
}

// Do returns the http.Response / error from the DoReturn stack and records the request on the DoCalled stack
func (client *MockHTTPClient) Do(req *http.Request) (*http.Response, errors.TracerError) {
	client.DoCalled.Push(req)
	top, err := client.DoReturn.Pop()
	if nil != err {
		panic("MockHTTPClient.Do called with no DoReturn set")
	}
	resp, ok := top.(*http.Response)
	if ok {
		var err errors.TracerError
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			err = NewBadStatusError(req.Method, req.URL.String(), resp.StatusCode)
		}
		return resp, err
	}

	err, ok = top.(error)
	if !ok {
		return nil, errors.New("MockHTTPClient.Do invalid return value :: %#v", top)
	}
	return nil, errors.Wrap(err)
}

// DoWithContext returns the http.Response / error from the DoReturn stack and records the request on the DoCalled stack
func (client *MockHTTPClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, errors.TracerError) {
	return client.Do(req)
}

// DoCalledPop pops the last http.Request from the DoCalled stack if it exists
func (client *MockHTTPClient) DoCalledPop() *http.Request {
	top, err := client.DoCalled.Pop()
	if nil != err {
		return nil
	}
	return top.(*http.Request)
}

// AddCookieJar to http client to make them available to all future requests
func (client *MockHTTPClient) AddCookieJar(jar http.CookieJar) {
	client.cookieJar = map[string][]*http.Cookie{}
}

// Cookies lists cookies in the jar
func (client *MockHTTPClient) Cookies(url *url.URL) []*http.Cookie {
	urlStr := url.RawPath
	return client.cookieJar[urlStr]
}

// SetCookies adds cookies to the jar
func (client *MockHTTPClient) SetCookies(url *url.URL, cookies []*http.Cookie) {
	urlStr := url.RawPath
	client.cookieJar[urlStr] = append(client.cookieJar[urlStr], cookies...)
}

// NewMockHTTPClient returns a mocked version of the DoHTTPRequest interface
func NewMockHTTPClient(returnStackItems ...interface{}) *MockHTTPClient {
	c := &MockHTTPClient{DoReturn: collection.NewStack(), DoCalled: collection.NewStack()}
	for _, item := range returnStackItems {
		c.DoReturn.Push(item)
	}
	return c
}

// MockListener is a mock implementation of the net.MockListener interface
type MockListener struct {
	Address       net.Addr
	GetConnection func() *MockConn
	AcceptError   error
}

// Accept returnes the MockConn and the AcceptError
func (listener *MockListener) Accept() (conn net.Conn, err error) {
	return listener.GetConnection(), listener.AcceptError
}

// Addr returns the listener's network address.
func (listener *MockListener) Addr() net.Addr { return listener.Address }

// Close is a no-op.
func (listener *MockListener) Close() error { return nil }

// MockConn is a mock implementation of the io.ReadWriteCloser interface
type MockConn struct {
	ID            int
	RAddress      net.Addr
	LAddress      net.Addr
	Closed        bool
	ReadComplete  bool
	ReadMultiple  bool
	Deadline      time.Time
	ReadDeadline  time.Time
	WriteDeadline time.Time
	ReadF         func(b []byte) (n int, err error)
	WriteF        func(b []byte) (n int, err error)
}

// Read mocks reading to a connection
func (conn *MockConn) Read(b []byte) (n int, err error) {
	if nil == conn {
		return 0, errors.New("cannot 'Read' on nil connection")
	}
	if conn.Closed {
		return 0, errors.New("attempting to 'Read' to a closed connection")
	}
	if conn.ReadComplete && !conn.ReadMultiple {
		return 0, nil
	}
	if nil == conn.ReadF {
		return 0, errors.New("conn.ReadF is nil")
	}
	conn.ReadComplete = true
	return conn.ReadF(b)
}

// Write mocks writing to a connection
func (conn *MockConn) Write(b []byte) (n int, err error) {
	if conn.Closed {
		return 0, errors.New("attempting to 'Write' to a closed connection")
	}
	if nil == conn.WriteF {
		return 0, errors.New("conn.WriteF is nil")
	}
	return conn.WriteF(b)
}

// Close sets the Closed attribute to true on the mock connection
func (conn *MockConn) Close() error {
	if nil != conn {
		conn.Closed = true
	}
	return nil
}

// LocalAddr returns a new mock Addr or the LAddress on the mock connection
func (conn *MockConn) LocalAddr() net.Addr {
	if nil == conn || nil == conn.LAddress {
		return &MockAddr{}
	}
	return conn.LAddress
}

// RemoteAddr returns a new mock Addr or the RAddress on the mock connection
func (conn *MockConn) RemoteAddr() net.Addr {
	if nil == conn || nil == conn.RAddress {
		return &MockAddr{}
	}
	return conn.RAddress
}

// SetDeadline is a no-op mock that returns nil
func (conn *MockConn) SetDeadline(t time.Time) error {
	conn.Deadline = t
	return nil
}

// SetReadDeadline is a no-op mock that returns nil
func (conn *MockConn) SetReadDeadline(t time.Time) error {
	if nil != conn {
		conn.ReadDeadline = t
	}
	return nil
}

// SetWriteDeadline is a no-op mock that returns nil
func (conn *MockConn) SetWriteDeadline(t time.Time) error {
	if nil != conn {
		conn.WriteDeadline = t
	}
	return nil
}

// MockAddr implements the net.MockAddr interface
type MockAddr struct {
	SNetwork string
	Address  string
}

// Network is the name of the network address
func (addr *MockAddr) Network() string { return addr.SNetwork }

// String form of the address
func (addr *MockAddr) String() string { return addr.Address }
