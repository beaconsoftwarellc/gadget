package errors

import (
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

// TaggedError is an error that has been tagged with a status.Status
type TaggedError interface {
	error
	// GRPCStatus for this error
	GRPCStatus() *status.Status
}

type taggedError struct {
	*status.Status
	err error
}

func (e taggedError) Error() string {
	return e.Status.Err().Error()
}

func (e taggedError) GRPCStatus() *status.Status {
	return e.Status
}

// TagStatus by recomposing its error and return as a [TaggedError]
func TagStatus(tag string, base *status.Status) TaggedError {
	var (
		err       = base.Err()
		baseError string
	)
	if err != nil {
		baseError = err.Error()
	} else {
		baseError = "<empty>"
	}
	return &taggedError{
		Status: status.Newf(base.Code(),
			"[%s]: %s", tag, baseError),
		err: err,
	}
}

// Tag the passed error
func Tag(tag string, base HasCode) TaggedError {
	if base == nil {
		return nil
	}
	return TagAndCode(tag, base.GetCode(), base)
}

// TagAndCode the passed error
func TagAndCode(tag string, code codes.Code, base error) TaggedError {
	if base == nil {
		return nil
	}
	return &taggedError{
		Status: status.Newf(code,
			"[%s]: %s", tag, base.Error()),
		err: base,
	}
}

// GetBase returns the original error from the passed tagged error
func GetBase(err error) error {
	if err == nil {
		return nil
	}
	if tagged, ok := err.(*taggedError); ok {
		return tagged.err
	}
	// status does not preserve the original error, but
	// better than nothing
	if _status, ok := status.FromError(err); !ok {
		return _status.Err()
	}
	return err
}
