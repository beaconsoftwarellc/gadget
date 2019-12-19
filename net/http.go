package net

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/log"
)

const (
	// HeaderAuthorization is the HTTP header for Authorization
	HeaderAuthorization = "Authorization"
	// HeaderContentType is the HTTP header Content Type
	HeaderContentType = "Content-Type"
	// HeaderAccept is the HTTP header Accept
	HeaderAccept = "Accept"
	// HeaderHost specifies the domain name of the server
	HeaderHost = "Host"
	// HeaderKeepAlive allows the sender to hint about how the connection may be used to set a timeout and max # of requests
	HeaderKeepAlive = "Keep-Alive"
	// HeaderConnection controls whether or not the network connection stays open after the transaction finishes
	HeaderConnection = "Connection"
	// HeaderUserAgent allows network peers to identify the application type, os, software vendor or version of requesting software user agent
	HeaderUserAgent = "User-Agent"
	// HeaderCacheControl specifies directives for caching mechanisms in both requests and responses
	HeaderCacheControl = "Cache-Control"

	// MIMEAppJSON is the HTTP value application/json for ContentType
	MIMEAppJSON = "application/json"
	// MIMEAppFormURLEncoded is the HTTP value application/x-www-form-urlencoded for ContentType
	MIMEAppFormURLEncoded = "application/x-www-form-urlencoded"
	// CacheControlNone is the Cache-Control value for disabling cacheing
	CacheControlNone = "no-cache"
)

// BadStatusError is returned when a request results in a non-successful status code ![200-299]
type BadStatusError struct {
	Method string
	URL    string
	Status int
	trace  []string
}

// NewBadStatusError for a http request
func NewBadStatusError(method string, url string, status int) errors.TracerError {
	return &BadStatusError{Method: method, URL: url, Status: status, trace: errors.GetStackTrace()}
}

func (err *BadStatusError) Error() string {
	return fmt.Sprintf("bad status: %s %s %d", err.Method, err.URL, err.Status)
}

// Trace for this error.
func (err *BadStatusError) Trace() []string {
	return err.trace
}

// DoHTTPRequest provides an interface for an HTTP Client.
type DoHTTPRequest interface {
	// Do the request by sending the payload to the remote server and returning the response and any errors
	Do(*http.Request) (*http.Response, errors.TracerError)
	// DoWithContext the request by sending the payload to the remote server and returning the response and any errors
	// cancelling the request at the transport level when the context returns on it's 'Done' channel.
	DoWithContext(context.Context, *http.Request) (*http.Response, errors.TracerError)
	// AddCookieJar to http client to make cookies available to future requests
	AddCookieJar(http.CookieJar)
	// Cookies lists cookies in the jar
	Cookies(url *url.URL) []*http.Cookie
	// SetCookies adds cookies to the jar
	SetCookies(url *url.URL, cookies []*http.Cookie)
}

// NewHTTPRedirectClient is the default net/http client with headers being set on redirect
func NewHTTPRedirectClient(timeout time.Duration) DoHTTPRequest {
	transport := &http.Transport{}
	return &httpRedirectClient{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				if len(via) == 0 {
					return nil
				}
				for attr, val := range via[0].Header {
					if _, ok := req.Header[attr]; !ok {
						req.Header[attr] = val
					}
				}
				return nil
			},
		},
		transport: transport,
	}
}

type httpRedirectClient struct {
	client    *http.Client
	transport *http.Transport
}

// Do the request by sending the payload to the remote server and returning the response and any errors
func (client *httpRedirectClient) Do(req *http.Request) (*http.Response, errors.TracerError) {
	log.Debugf("sending request to %s", req.URL.String())
	now := time.Now()
	resp, err := client.client.Do(req)
	log.Debugf("request to %s complete in %s", req.URL.String(), time.Now().Sub(now))
	if nil == err && (resp.StatusCode < 200 || resp.StatusCode > 299) {
		err = NewBadStatusError(req.Method, req.URL.String(), resp.StatusCode)
	}
	return resp, errors.Wrap(err)
}

// DoWithContext the request by sending the payload to the remote server and returning the response and any errors
// cancelling the request at the transport level when the context returns on it's 'Done' channel.
func (client *httpRedirectClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, errors.TracerError) {
	complete := make(chan bool, 1)
	var response *http.Response
	var err error
	go func() {
		response, err = client.Do(req)
		complete <- true
	}()
	select {
	case <-ctx.Done():
		client.transport.CancelRequest(req)
		err = errors.New("request to %s was cancelled by controlling context", req.URL.String())
	case <-complete:
		break
	}
	return response, errors.Wrap(err)
}

// AddCookieJar to http client to make them available to all future requests
func (client *httpRedirectClient) AddCookieJar(jar http.CookieJar) {
	client.client.Jar = jar
}

// Cookies lists cookies in the jar
func (client *httpRedirectClient) Cookies(url *url.URL) []*http.Cookie {
	return client.client.Jar.Cookies(url)
}

// SetCookies adds cookies to the jar
func (client *httpRedirectClient) SetCookies(url *url.URL, cookies []*http.Cookie) {
	client.client.Jar.SetCookies(url, cookies)
}
