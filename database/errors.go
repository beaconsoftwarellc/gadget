package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/generator"
	"github.com/beaconsoftwarellc/gadget/log"
)

// ConnectionError  is returned when unable to connect to database
type ConnectionError struct{ trace []string }

func (err *ConnectionError) Error() string {
	return "database connection failed"
}

// Trace returns the stack trace for the error
func (err *ConnectionError) Trace() []string {
	return err.trace
}

// NewDatabaseConnectionError instantiates a DatabaseConnectionError with a stack trace
func NewDatabaseConnectionError() errors.TracerError {
	return &ConnectionError{trace: errors.GetStackTrace()}
}

// SQLQueryType indicates the type of query being executed that caused and error
type SQLQueryType string

const (
	// Select indicates a SELECT statement triggered the error
	Select SQLQueryType = "SELECT"
	// Insert indicates an INSERT statement triggered the error
	Insert = "INSERT"
	// Delete indicates a DELETE statement triggered the error
	Delete = "DELETE"
	// Update indicates an UPDATE statement triggered the error
	Update = "UPDATE"
)

const (
	dbErrPrefix          = "dberr"
	invalidForeignKeyMsg = "invalid reference"
	dataTooLongMsg       = "data too long"
	duplicateRecordMsg   = "already exists"
)

const (
	mysqlDuplicateEntry    = 1062
	mysqlDataTooLong       = 1406
	mysqlInvalidForeignKey = 1452
)

const primaryKeyConstraintCheck = "for key 'PRIMARY'"

// TranslateError converts a mysql or other obtuse errors into discrete explicit errors
func TranslateError(err error, action SQLQueryType, stmt string, logger log.Logger) errors.TracerError {
	if nil == err {
		return nil
	}
	if sql.ErrNoRows == err {
		return NewNotFoundError()
	}
	driverErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return NewSystemError(action, stmt, err, logger)
	}
	switch driverErr.Number {
	// Duplicate primary key
	case mysqlDuplicateEntry:
		if strings.Contains(err.Error(), primaryKeyConstraintCheck) {
			return NewDuplicateRecordError(action, stmt, err, logger)
		}
		return NewUniqueConstraintError(action, stmt, err, logger)
	// Data too long for column
	case mysqlDataTooLong:
		return NewDataTooLongError(action, stmt, err, logger)
	// Invalid foreign key
	case mysqlInvalidForeignKey:
		return NewInvalidForeignKeyError(action, stmt, err, logger)
	default:
		return NewExecutionError(action, stmt, err, logger)
	}
}

// SQLExecutionError is returned when a query against the database fails
type SQLExecutionError struct {
	Action      SQLQueryType
	ReferenceID string
	message     string
	Stmt        string
	ErrMsg      string
	trace       []string
}

// NewExecutionError logs the error and returns an ExecutionError
func NewExecutionError(action SQLQueryType, stmt string, err error, logger log.Logger) errors.TracerError {
	e := &SQLExecutionError{
		ReferenceID: generator.ID(dbErrPrefix),
		Action:      action,
		Stmt:        stmt,
		ErrMsg:      err.Error(),
		trace:       errors.GetStackTrace(),
	}
	e.message = fmt.Sprintf("%s: caused a database error", e.Action)
	logger.Errorf("%#v", e)
	return e
}

// Error prints a ExecutionError
func (e *SQLExecutionError) Error() string {
	return fmt.Sprintf("%s (%s)", e.message, e.ReferenceID)
}

// Trace returns the stack trace for the error
func (e *SQLExecutionError) Trace() []string {
	return e.trace
}

// NewValidationError returns a ValidationError with a stack trace
func NewValidationError(msg string, subs ...interface{}) errors.TracerError {
	return &ValidationError{
		message: fmt.Sprintf(msg, subs...),
		trace:   errors.GetStackTrace(),
	}
}

// ValidationError is returned when a query against the database fails
type ValidationError struct {
	message string
	trace   []string
}

// Error prints a ValidationError
func (e *ValidationError) Error() string {
	return e.message
}

// Trace returns the stack trace for the error
func (e *ValidationError) Trace() []string {
	return e.trace
}

// NewNotFoundError returns a NotFoundError with a stack trace
func NewNotFoundError() errors.TracerError {
	return &NotFoundError{
		trace: errors.GetStackTrace(),
	}
}

// NotFoundError is returned when a query against the database fails
type NotFoundError struct {
	trace []string
}

// Error prints a NotFoundError
func (e *NotFoundError) Error() string {
	return "no-records"
}

// Trace returns the stack trace for the error
func (e *NotFoundError) Trace() []string {
	return e.trace
}

// SQLSystemError is returned when a database action fails
type SQLSystemError struct {
	SQLExecutionError
}

// NewSystemError logs the error and returns an ExecutionError
func NewSystemError(action SQLQueryType, stmt string, err error, logger log.Logger) errors.TracerError {
	e := &SQLSystemError{
		SQLExecutionError{
			ErrMsg:      err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     "could not execute query",
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
	logger.Errorf("%#v", e)
	return e
}

// NotAPointerError  indicates that a record object isn't a pointer
type NotAPointerError struct{ trace []string }

func (err *NotAPointerError) Error() string {
	return "must be a pointer"
}

// Trace returns the stack trace for the error
func (err *NotAPointerError) Trace() []string {
	return err.trace
}

// NewNotAPointerError instantiates a NotAPointerError with a stack trace
func NewNotAPointerError() errors.TracerError {
	return &NotAPointerError{trace: errors.GetStackTrace()}
}

// DuplicateRecordError is returned when a mysql error #1062 occurs for a PrimaryKey
type DuplicateRecordError struct {
	SQLExecutionError
}

// NewDuplicateRecordError is returned when a records is created/updated with a duplicate primary key
func NewDuplicateRecordError(action SQLQueryType, stmt string, err error, logger log.Logger) errors.TracerError {
	e := &DuplicateRecordError{
		SQLExecutionError{ErrMsg: err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     duplicateRecordMsg,
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
	logger.Error(e)
	return e
}

// UniqueConstraintError is returned when a mysql error #1062 occurs for a Unique constraint
type UniqueConstraintError struct {
	SQLExecutionError
}

// NewUniqueConstraintError is returned when a record is created/updated with a duplicate primary key
func NewUniqueConstraintError(action SQLQueryType, stmt string, err error, logger log.Logger) errors.TracerError {
	e := &UniqueConstraintError{
		SQLExecutionError{ErrMsg: err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     duplicateRecordMsg,
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
	logger.Error(e)
	return e
}

// DataTooLongError is returned when a mysql error #1406 occurs
type DataTooLongError struct {
	SQLExecutionError
}

// NewDataTooLongError logs the error and returns an instantiated DataTooLongError
func NewDataTooLongError(action SQLQueryType, stmt string, err error, logger log.Logger) errors.TracerError {
	e := &DataTooLongError{
		SQLExecutionError{ErrMsg: err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     dataTooLongMsg,
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
	logger.Error(e)
	return e
}

// InvalidForeignKeyError is returned when a mysql error #1452 occurs
type InvalidForeignKeyError struct {
	SQLExecutionError
}

// NewInvalidForeignKeyError logs the error and returns an instantiated InvalidForeignKeyError
func NewInvalidForeignKeyError(action SQLQueryType, stmt string, err error, logger log.Logger) errors.TracerError {
	e := &InvalidForeignKeyError{
		SQLExecutionError{ErrMsg: err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     invalidForeignKeyMsg,
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
	logger.Error(e)
	return e
}
