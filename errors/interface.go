package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-stack/stack"
)

// Tracer provides a stack trace
type Tracer interface {
	// Trace returns the long format stack trace from where the error was instantiated
	Trace() []string
}

// TracerError is an amalgamation of the Tracer and error interfaces
type TracerError interface {
	Tracer
	error
}

// New instantiates a TracerError with a stack trace
func New(format string, args ...interface{}) TracerError {
	return &errorTracer{
		message: fmt.Sprintf(format, args...),
		trace:   GetStackTrace(),
	}
}

type errorTracer struct {
	message string
	trace   []string
}

func (err *errorTracer) Error() string {
	return err.message
}

func (err *errorTracer) Trace() []string {
	return err.trace
}

// GetStackTrace retrieves the full stack (minus any runtime related functions) from where the Error was instantiated
func GetStackTrace() []string {
	//may want to trim the log calls, may want to return this differently so that it can be formatted in other ways
	trace := stack.Trace().TrimRuntime()
	var longTrace []string
	for _, t := range trace {
		tf := fmt.Sprintf("%+v", t)
		// only remove frames that originate in this file
		if !strings.HasPrefix(tf, "github.com/beaconsoftwarellc/gadget/errors/interface.go") {
			longTrace = append(longTrace, tf)
		}
	}
	return longTrace
}

type wrapErrorTracer struct {
	err   error
	trace []string
}

// Wrap an existing error in a TracerError
func Wrap(err error) TracerError {
	if nil == err {
		return nil
	}

	tracerError, ok := err.(TracerError)
	if !ok {
		return &wrapErrorTracer{
			err:   err,
			trace: GetStackTrace(),
		}
	}
	return tracerError
}

func (err *wrapErrorTracer) Error() string {
	return err.err.Error()
}

func (err *wrapErrorTracer) Trace() []string {
	return err.trace
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}
