package errors

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SQLQueryType indicates the type of query being executed that caused an error
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
	dbErrPrefix               = "dberr"
	invalidForeignKeyMsg      = "invalid reference"
	dataTooLongMsg            = "data too long"
	duplicateRecordMsg        = "already exists"
	mysqlDuplicateEntry       = 1062
	mysqlDataTooLong          = 1406
	mysqlInvalidForeignKey    = 1452
	primaryKeyConstraintCheck = "for key 'PRIMARY'"
)

// IsNotFoundError returns a boolean indicating that the passed error (can be nil) is of
// type *database.NotFoundError
func IsNotFoundError(err error) bool {
	var dst *NotFoundError
	return errors.As(err, &dst)
}

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

// TranslateError converts a mysql or other obtuse errors into discrete explicit errors
func TranslateError(err error, action SQLQueryType, stmt string) errors.TracerError {
	if nil == err {
		return nil
	}
	if sql.ErrNoRows == err {
		return NewNotFoundError()
	}
	driverErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return NewSystemError(action, stmt, err)
	}
	switch driverErr.Number {
	// Duplicate primary key
	case mysqlDuplicateEntry:
		if strings.Contains(err.Error(), primaryKeyConstraintCheck) {
			return NewDuplicateRecordError(action, stmt, err)
		}
		return NewUniqueConstraintError(action, stmt, err)
	// Data too long for column
	case mysqlDataTooLong:
		return NewDataTooLongError(action, stmt, err)
	// Invalid foreign key
	case mysqlInvalidForeignKey:
		return NewInvalidForeignKeyError(action, stmt, err)
	default:
		return NewExecutionError(action, stmt, err)
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

// NewExecutionError with wrapping the passed error with the passed action and sql statement
func NewExecutionError(action SQLQueryType, stmt string, err error) errors.TracerError {
	e := &SQLExecutionError{
		ReferenceID: generator.ID(dbErrPrefix),
		Action:      action,
		Stmt:        stmt,
		ErrMsg:      err.Error(),
		trace:       errors.GetStackTrace(),
	}
	e.message = fmt.Sprintf("%s: caused a database error", e.Action)
	return e
}

// Error logs the reference id and statement for reference and returns a string
// representation of this error containing the reference ID
func (e *SQLExecutionError) Error() string {
	log.Infof("[Ref:%s] STMT: %s", e.ReferenceID, e.Stmt)
	return fmt.Sprintf("%s: %s [Ref:%s]",
		e.message, e.ErrMsg, e.ReferenceID)
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

// NewSystemError returns an ExecutionError
func NewSystemError(action SQLQueryType, stmt string, err error) errors.TracerError {
	return &SQLSystemError{
		SQLExecutionError{
			ErrMsg:      err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     "could not execute query",
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
}

// NotAPointerError indicates that a record object isn't a pointer
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
func NewDuplicateRecordError(action SQLQueryType, stmt string, err error) errors.TracerError {
	return &DuplicateRecordError{
		SQLExecutionError{ErrMsg: err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     duplicateRecordMsg,
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
}

// UniqueConstraintError is returned when a mysql error #1062 occurs for a Unique constraint
type UniqueConstraintError struct {
	SQLExecutionError
}

// NewUniqueConstraintError is returned when a record is created/updated with a duplicate primary key
func NewUniqueConstraintError(action SQLQueryType, stmt string, err error) errors.TracerError {
	return &UniqueConstraintError{
		SQLExecutionError{ErrMsg: err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     duplicateRecordMsg,
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
}

// DataTooLongError is returned when a mysql error #1406 occurs
type DataTooLongError struct {
	SQLExecutionError
}

// NewDataTooLongError wrapping the passer error with references to the passed sql statement
// and action
func NewDataTooLongError(action SQLQueryType, stmt string, err error) errors.TracerError {
	return &DataTooLongError{
		SQLExecutionError{ErrMsg: err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     dataTooLongMsg,
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
}

// InvalidForeignKeyError is returned when a mysql error #1452 occurs
type InvalidForeignKeyError struct {
	SQLExecutionError
}

// NewInvalidForeignKeyError wrapping the passed error with references to the passed
// sql and action.
func NewInvalidForeignKeyError(action SQLQueryType, stmt string, err error) errors.TracerError {
	return &InvalidForeignKeyError{
		SQLExecutionError{ErrMsg: err.Error(),
			ReferenceID: generator.ID(dbErrPrefix),
			Action:      action,
			message:     invalidForeignKeyMsg,
			Stmt:        stmt,
			trace:       errors.GetStackTrace(),
		},
	}
}

// DatabaseToStatus translates the passed db error into a grpc Status with appropriate
// status code
func DatabaseToStatus(primary qb.Table, dbError error) *status.Status {
	// we have to call through so we have the same stack offset in both
	// public functions
	return databaseToStatus(primary, dbError)
}

func databaseToStatus(primary qb.Table, dbError error) *status.Status {
	if nil == dbError {
		return nil
	}
	var grpcStatus *status.Status
	prefix := getLogPrefix(3)
	switch dbError.(type) {
	case *NotFoundError:
		grpcStatus = status.Newf(codes.NotFound, "%s %s not found", prefix, primary.GetName())
	case *DataTooLongError:
		grpcStatus = status.Newf(codes.InvalidArgument, "%s %s field too long: %s",
			prefix, primary.GetName(), dbError)
	case *DuplicateRecordError:
		grpcStatus = status.Newf(codes.AlreadyExists, "%s %s record already exists: %s",
			prefix, primary.GetName(), dbError)
	case *UniqueConstraintError:
		grpcStatus = status.Newf(codes.AlreadyExists, "%s %s unique constraint violation: %s",
			prefix, primary.GetName(), dbError)
	case *InvalidForeignKeyError:
		grpcStatus = status.Newf(codes.InvalidArgument, "%s %s foreign key violation: %s",
			prefix, primary.GetName(), dbError)
	case *ValidationError:
		grpcStatus = status.Newf(codes.InvalidArgument, "%s operation on %s had a validation error: %s",
			prefix, primary.GetName(), dbError)
	case *ConnectionError, *NotAPointerError:
		_ = log.Errorf("[GAD.DAT.321] unexpected run time database error: %s", dbError)
		grpcStatus = status.Newf(codes.Internal, "%s internal system error encountered", prefix)
	default:
		_ = log.Errorf("[GAD.DAT.324] unhandled error type %T: %s", dbError, dbError.Error())
		grpcStatus = status.Newf(codes.Aborted, "%s (%s) database error encountered: %s",
			prefix, primary.GetName(), dbError)
	}
	return grpcStatus
}

// DatabaseToApiError handles conversion from a database error to a GRPC friendly
// error with code.
func DatabaseToApiError(primary qb.Table, dbError error) error {
	return databaseToStatus(primary, dbError).Err()
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

var logPrefixRegex = regexp.MustCompile(`([^\[]*\[\w{1,3}\.\w{1,3}\.)(\d+)(\][^\[]*)`)
var dbErrRegex = regexp.MustCompile(fmt.Sprintf("\\b%s_?[^\\W]*", dbErrPrefix))

// EqualLogError asserts that a function returned an error (i.e. not `nil`)
// and that it is equal to the provided error, ignoring line number in the log prefix
// and any database error ids.
func EqualLogError(assert assertion, theError error, errString string, msgAndArgs ...interface{}) bool {
	normError := theError
	normErrorStr := errString
	if nil != theError {
		// remove log prefix line numbers
		normErrorStr = logPrefixRegex.ReplaceAllString(errString, "${1}${3}")
		normError = errors.New(logPrefixRegex.ReplaceAllString(theError.Error(), "${1}${3}"))
		// remove db error string
		normErrorStr = dbErrRegex.ReplaceAllString(normErrorStr, "")
		normError = errors.New(dbErrRegex.ReplaceAllString(normError.Error(), ""))
	}
	return assert.EqualError(normError, normErrorStr, msgAndArgs...)
}
