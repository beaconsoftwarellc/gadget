package database

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/beaconsoftwarellc/gadget/database/qb"
	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/generator"
	"github.com/beaconsoftwarellc/gadget/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConnectionError  is returned when unable to connect to database
type ConnectionError struct {
	err   error
	trace []string
}

func (err *ConnectionError) Error() string {
	return err.err.Error()
}

// Trace returns the stack trace for the error
func (err *ConnectionError) Trace() []string {
	return err.trace
}

// NewDatabaseConnectionError instantiates a DatabaseConnectionError with a stack trace
func NewDatabaseConnectionError(err error) errors.TracerError {
	return &ConnectionError{err: err, trace: errors.GetStackTrace()}
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

// DatabaseToApiError handles conversion from a database error to a GRPC friendly
// error with code.
func DatabaseToApiError(primary qb.Table, dbError error) error {
	if nil == dbError {
		return nil
	}
	var err error
	prefix := getLogPrefix(2)
	switch dbError.(type) {
	case *NotFoundError:
		err = status.Error(codes.NotFound, fmt.Sprintf("%s %s not found", prefix, primary.GetName()))
	case *DataTooLongError:
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("%s %s field too long: %s",
			prefix, primary.GetName(), dbError))
	case *DuplicateRecordError:
		err = status.Error(codes.AlreadyExists, fmt.Sprintf("%s %s record already exists: %s",
			prefix, primary.GetName(), dbError))
	case *UniqueConstraintError:
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("%s %s unique constraint violation: %s",
			prefix, primary.GetName(), dbError))
	case *ValidationError:
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("%s operation on %s had a validation error: %s",
			prefix, primary.GetName(), dbError))
	case *ConnectionError, *NotAPointerError:
		_ = log.Errorf("[GAD.DAT.321] unexpected run time database error: %s", dbError)
		err = status.Error(codes.Internal, fmt.Sprintf("%s internal system error encountered", prefix))
	default:
		_ = log.Errorf("[GAD.DAT.324] unhandled error type %T: %s", dbError, dbError.Error())
		err = status.Error(codes.Aborted, fmt.Sprintf("%s (%s) database error encountered: %s",
			prefix, primary.GetName(), dbError))
	}
	return err
}

func getLogPrefix(frameSkip int) string {
	_, filePath, lineNumber, ok := runtime.Caller(frameSkip)
	if !ok {
		_ = log.Warnf("failed to lookup runtime.Caller(%d) lookup failed", frameSkip)
		return "[UNK]"
	}
	pathSplit := strings.Split(filePath, string(os.PathSeparator))
	var a, b string
	if len(pathSplit) > 2 {
		a = getPrefixPart(pathSplit[len(pathSplit)-2])
		b = getPrefixPart(pathSplit[len(pathSplit)-3])
	} else {
		a = "UNK"
		b = getPrefixPart(filePath)
	}
	return fmt.Sprintf("[%s.%s.%d]", b, a, lineNumber)
}

func getPrefixPart(s string) string {
	runes := []rune(strings.TrimSpace(s))
	part := []rune{'_', '_', '_'}
	for i := 0; i < len(part) && i < len(runes); i++ {
		part[i] = runes[i]
	}
	return strings.ToUpper(string(part))
}

type assertion interface {
	EqualError(theError error, errString string, msgAndArgs ...interface{}) bool
}

var logPrefixRegex = regexp.MustCompile(`^(.*\[\w{1,3}\.\w{1,3}\.)(\d+)(\].*)$`)

// EqualLogError asserts that a function returned an error (i.e. not `nil`)
// and that it is equal to the provided error, ignoring line number in the log prefix.
func EqualLogError(assert assertion, theError error, errString string, msgAndArgs ...interface{}) bool {
	//normalize log prefix line number
	if nil == theError {
		return false
	}
	line := logPrefixRegex.ReplaceAllString(theError.Error(), "${2}")
	newErrStr := logPrefixRegex.ReplaceAllString(errString, "${1}"+line+"${3}")
	return assert.EqualError(theError, newErrStr, msgAndArgs...)
}
