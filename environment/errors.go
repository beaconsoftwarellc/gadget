package environment

import (
	"fmt"
	"reflect"

	"github.com/beaconsoftwarellc/gadget/errors"
)

// InvalidSpecificationError  indicates that a config is of the wrong type.
type InvalidSpecificationError struct{ trace []string }

func (err *InvalidSpecificationError) Error() string {
	return "specification must be a struct pointer"
}

// Trace returns the stack trace for the error
func (err *InvalidSpecificationError) Trace() []string {
	return err.trace
}

// NewInvalidSpecificationError instantiates a InvalidSpecificationError with a stack trace
func NewInvalidSpecificationError() errors.TracerError {
	return &InvalidSpecificationError{trace: errors.GetStackTrace()}
}

// MissingEnvironmentVariableError indicates that a non-optional variable was not found in the environment
type MissingEnvironmentVariableError struct {
	Field string
	Tag   string
	trace []string
}

// NewMissingEnvironmentVariableError instantiates a MissingEnvironmentVariableError with a stack trace
func NewMissingEnvironmentVariableError(field string, tag string) errors.TracerError {
	return &MissingEnvironmentVariableError{
		Field: field,
		Tag:   tag,
		trace: errors.GetStackTrace(),
	}
}

func (err MissingEnvironmentVariableError) Error() string {
	return fmt.Sprintf("required environment variable %s was not set for %s", err.Tag, err.Field)
}

// Trace returns the stack trace for the error
func (err *MissingEnvironmentVariableError) Trace() []string {
	return err.trace
}

// UnsupportedDataTypeError indicates that no conversion from string to the given type has been implemented
type UnsupportedDataTypeError struct {
	Type  reflect.Kind
	Field string
	trace []string
}

// NewUnsupportedDataTypeError instantiates a UnsupportedDataTypeError with a stack trace
func NewUnsupportedDataTypeError(dataType reflect.Kind, field string) errors.TracerError {
	return &UnsupportedDataTypeError{
		Type:  dataType,
		Field: field,
		trace: errors.GetStackTrace(),
	}
}

func (err UnsupportedDataTypeError) Error() string {
	return fmt.Sprintf("type %s for %s is not supported", err.Type, err.Field)
}

// Trace returns the stack trace for the error
func (err *UnsupportedDataTypeError) Trace() []string {
	return err.trace
}
