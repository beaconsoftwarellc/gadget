package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

// ErrTesting is for use in error test cases only. Should
// primarily be used for returning generic errors from mocks and
// ensuring the error is expected
type ErrTesting string

func (e ErrTesting) GetCode() codes.Code {
	return codes.Unknown
}

func (e ErrTesting) Error() string {
	return string(e)
}

// ErrParameterRequired is returned when a required parameter is empty or nil
// but requires a value.
type ErrParameterRequired string

func (e ErrParameterRequired) GetCode() codes.Code {
	return codes.InvalidArgument
}

func (e ErrParameterRequired) Error() string {
	return fmt.Sprintf("parameter '%s' cannot be empty", string(e))
}

// ErrRequiresSuperUser is returned when a user attempts to perform an action
// that requires superuser privileges.
type ErrRequiresSuperUser string

func (e ErrRequiresSuperUser) GetCode() codes.Code {
	return codes.PermissionDenied
}

func (e ErrRequiresSuperUser) Error() string {
	return fmt.Sprintf("action %s requires a superuser", string(e))
}

// ErrNotFound is returned when a resource does not exist for a
// given identifier.
type ErrNotFound string

func (e ErrNotFound) GetCode() codes.Code {
	return codes.NotFound
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("resource '%s' not found", string(e))
}

// ErrInvalidArgument value for the expected functionality.
type ErrInvalidArgument string

func (e ErrInvalidArgument) GetCode() codes.Code {
	return codes.InvalidArgument
}

func (e ErrInvalidArgument) Error() string {
	return fmt.Sprintf("argument '%s' value is invalid", string(e))
}
