package net

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

const (
	// HeaderAuthorization is the HTTP header for Authorization
	HeaderAuthorization = "Authorization"
	// HeaderContentType is the HTTP header Content Type
	HeaderContentType = "Content-Type"
	// HeaderContentLength is the HTTP header name for specifying length of the body in bytes
	HeaderContentLength = "Content-Length"
	// HeaderAccept is the HTTP header Accept
	HeaderAccept = "Accept"
	// HeaderAllow is the HTTP header for Allow which indicates which methods are allowed on the URL
	HeaderAllow = "Allow"
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
	// HeaderContentEncoding specifies how the payload of the http message is encoded
	HeaderContentEncoding = "Content-Encoding"

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
