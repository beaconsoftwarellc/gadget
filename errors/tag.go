package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HasCode exposes functionality for switching from an error to a
// status error
type HasCode interface {
	error
	// GetCode for this error
	GetCode() codes.Code
}

// Tag the passed error
func Tag(tag string, base HasCode) error {
	return status.Error(base.GetCode(),
		fmt.Sprintf("[%s]: %s", tag, base.Error()))
}

// TagAndCode the passed error
func TagAndCode(tag string, code codes.Code, base error) error {
	return status.Error(code, fmt.Sprintf("[%s]: %s", tag, base.Error()))
}

// GetOriginalError returns the original error from the passed status error
func GetOriginalError(statusError error) error {
	if statusError == nil {
		return nil
	}
	_status, ok := status.FromError(statusError)
	if !ok {
		return _status.Err()
	}
	return statusError
}
