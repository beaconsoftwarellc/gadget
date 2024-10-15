package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

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

func (e ErrRequiresSuperUser) Trace() string {
	return fmt.Sprintf("action %s requires a superuser", string(e))
}

func (e ErrRequiresSuperUser) GetCode() codes.Code {
	return codes.PermissionDenied
}

func (e ErrRequiresSuperUser) Error() string {
	return fmt.Sprintf("action %s requires a superuser", string(e))
}
