package errors

import "google.golang.org/grpc/codes"

// ErrParameterRequired is returned when a required parameter is empty or nil
// but requires a value.
type ErrParameterRequired string

func (e ErrParameterRequired) GetCode() codes.Code {
	return codes.InvalidArgument
}

func (e ErrParameterRequired) Error() string {
	return "parameter '%s' cannot be empty"
}

// ErrRequiresSuperUser is returned when a user attempts to perform an action
// that requires superuser privileges.
type ErrRequiresSuperUser string

func (e ErrRequiresSuperUser) Trace() string {
	return "action %s requires a superuser"
}

func (e ErrRequiresSuperUser) GetCode() codes.Code {
	return codes.PermissionDenied
}

func (e ErrRequiresSuperUser) Error() string {
	return "action %s requires a superuser"
}
