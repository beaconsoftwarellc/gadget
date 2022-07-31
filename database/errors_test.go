package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

func TestExecutionError(t *testing.T) {
	assert := assert1.New(t)
	err := NewExecutionError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*SQLExecutionError)
	err2 := NewExecutionError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*SQLExecutionError)
	assert.True(strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(err.ReferenceID, err2.ReferenceID)
	assert.Contains(err.Error(), err.ReferenceID)
	assert.Contains(err.Error(), err.message)
}

func TestNewNotFoundError(t *testing.T) {
	assert := assert1.New(t)
	err := NewNotFoundError()
	assert.EqualError(err, NewNotFoundError().Error())
}

func TestNewSystemError(t *testing.T) {
	assert := assert1.New(t)
	err := NewSystemError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*SQLSystemError)
	err2 := NewSystemError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*SQLSystemError)
	assert.True(strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(err.ReferenceID, err2.ReferenceID)
	assert.Contains(err.Error(), err.ReferenceID)
	assert.Contains(err.Error(), err.message)
}

func TestNewDuplicateRecordError(t *testing.T) {
	assert := assert1.New(t)
	err := NewDuplicateRecordError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*DuplicateRecordError)
	err2 := NewDuplicateRecordError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*DuplicateRecordError)
	assert.True(strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(err.ReferenceID, err2.ReferenceID)
	assert.Contains(err.Error(), err.ReferenceID)
	assert.Contains(err.Error(), err.message)
}

func TestNewDataTooLongError(t *testing.T) {
	assert := assert1.New(t)
	err := NewDataTooLongError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*DataTooLongError)
	err2 := NewDataTooLongError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*DataTooLongError)
	assert.True(strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(err.ReferenceID, err2.ReferenceID)
	assert.Contains(err.Error(), err.ReferenceID)
	assert.Contains(err.Error(), err.message)
}

func TestNewInvalidForeignKeyError(t *testing.T) {
	assert := assert1.New(t)
	err := NewInvalidForeignKeyError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*InvalidForeignKeyError)
	err2 := NewInvalidForeignKeyError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*InvalidForeignKeyError)
	assert.True(strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(err.ReferenceID, err2.ReferenceID)
	assert.Contains(err.Error(), err.ReferenceID)
	assert.Contains(err.Error(), err.message)
}

func TestTranslateError(t *testing.T) {
	assert := assert1.New(t)
	testData := []struct {
		err      error
		expected error
	}{
		{err: sql.ErrNoRows, expected: &NotFoundError{}},
		{err: &mysql.MySQLError{Number: mysqlDuplicateEntry, Message: "foo ... " + primaryKeyConstraintCheck}, expected: &DuplicateRecordError{}},
		{err: &mysql.MySQLError{Number: mysqlDuplicateEntry}, expected: &UniqueConstraintError{}},
		{err: &mysql.MySQLError{Number: mysqlDataTooLong}, expected: &DataTooLongError{}},
		{err: &mysql.MySQLError{Number: mysqlInvalidForeignKey}, expected: &InvalidForeignKeyError{}},
		{err: &mysql.MySQLError{}, expected: &SQLExecutionError{}},
		{err: errors.New("foo"), expected: &SQLSystemError{}},
	}
	for _, data := range testData {
		assert.IsType(data.expected, TranslateError(data.err, Select, generator.String(5), log.NewStackLogger()))
	}
}

func Test_getLogPrefix(t *testing.T) {
	assert := assert1.New(t)
	expected := "[GAD.DAT.96]"
	actual := getLogPrefix(1)
	assert.Equal(expected, actual)
}

func Test_getPrefixPart(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected string
	}{
		{
			name:     "empty",
			s:        "",
			expected: "___",
		},
		{
			name:     "single",
			s:        "a",
			expected: "A__",
		},
		{
			name:     "double",
			s:        "ab",
			expected: "AB_",
		},
		{
			name:     "triple",
			s:        "abc",
			expected: "ABC",
		},
		{
			name:     "spaces are removed",
			s:        "    abc    ",
			expected: "ABC",
		},
		{
			name:     "all whitespace",
			s:        "        ",
			expected: "___",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert1.New(t)
			assert.Equal(tt.expected, getPrefixPart(tt.s))
		})
	}
}

type action struct {
	alias      string
	ID         qb.TableField
	Name       qb.TableField
	allColumns qb.TableField
}

func (a *action) GetName() string {
	return "action"
}

func (a *action) GetAlias() string {
	return a.alias
}

func (a *action) PrimaryKey() qb.TableField {
	return a.ID
}

func (a *action) SortBy() (qb.TableField, qb.OrderDirection) {
	return a.ID, qb.Ascending
}

func (a *action) AllColumns() qb.TableField {
	return a.allColumns
}

func (a *action) ReadColumns() []qb.TableField {
	return []qb.TableField{
		a.ID,
		a.Name,
	}
}

func (a *action) WriteColumns() []qb.TableField {
	return a.ReadColumns()
}

func (a *action) Alias(alias string) *action {
	return &action{
		alias:      alias,
		ID:         qb.TableField{Name: "id", Table: alias},
		Name:       qb.TableField{Name: "name", Table: alias},
		allColumns: qb.TableField{Name: "*", Table: alias},
	}
}

var Action = (&action{}).Alias("action")

func TestDatabaseToApiError(t *testing.T) {
	tests := []struct {
		name     string
		primary  qb.Table
		err      error
		expected string
	}{
		{
			name:     "nil error",
			primary:  Action,
			err:      nil,
			expected: "",
		},
		{
			name:     "not db error",
			primary:  Action,
			err:      errors.New("foo"),
			expected: "rpc error: code = Aborted desc = [GAD.DAT.262] (action) database error encountered: foo",
		},
		{
			name:     "data too long error",
			primary:  Action,
			err:      &DataTooLongError{},
			expected: "rpc error: code = InvalidArgument desc = [GAD.DAT.262] action field too long:  ()",
		},
		{
			name:     "not found",
			primary:  Action,
			err:      NewNotFoundError(),
			expected: "rpc error: code = NotFound desc = [GAD.DAT.262] action not found",
		},
		{
			name:     "duplicate record",
			primary:  Action,
			err:      &DuplicateRecordError{},
			expected: "rpc error: code = AlreadyExists desc = [GAD.DAT.262] action record already exists:  ()",
		},
		{
			name:     "unique constraint",
			primary:  Action,
			err:      &UniqueConstraintError{},
			expected: "rpc error: code = InvalidArgument desc = [GAD.DAT.262] action unique constraint violation:  ()",
		},
		{
			name:     "validation",
			primary:  Action,
			err:      &ValidationError{},
			expected: "rpc error: code = InvalidArgument desc = [GAD.DAT.262] operation on action had a validation error: ",
		},
		{
			name:     "not a pointer",
			primary:  Action,
			err:      &NotAPointerError{},
			expected: "rpc error: code = Internal desc = [GAD.DAT.262] internal system error encountered",
		},
		{
			name:     "connection",
			primary:  Action,
			err:      &ConnectionError{},
			expected: "rpc error: code = Internal desc = [GAD.DAT.262] internal system error encountered",
		},
		{
			name:     "foregin key",
			primary:  Action,
			err:      &InvalidForeignKeyError{},
			expected: "rpc error: code = InvalidArgument desc = [GAD.DAT.262] action foregin key violation:  ()",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert1.New(t)
			if stringutil.IsEmpty(tt.expected) {
				assert.NoError(DatabaseToApiError(tt.primary, tt.err))
			} else {
				EqualLogError(assert, DatabaseToApiError(tt.primary, tt.err), tt.expected)
			}
		})
	}
}

type MockAssertion struct{}

func (ma *MockAssertion) EqualError(theError error, errString string, msgAndArgs ...interface{}) bool {
	if nil == theError {
		return false
	}
	return theError.Error() == errString
}

func TestEqualLogError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
		equal    bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
			equal:    false,
		},
		{
			name:     "empty error",
			err:      errors.New(""),
			expected: "",
			equal:    true,
		},
		{
			name:     "line number match",
			err:      errors.New("[GAD.DAT.123]"),
			expected: "[GAD.DAT.123]",
			equal:    true,
		},
		{
			name:     "line number mismatch",
			err:      errors.New("[GAD.DAT.123]"),
			expected: "[GAD.DAT.456]",
			equal:    true,
		},
		{
			name:     "file prefix mismatch",
			err:      errors.New("[GAD.DAT.123]"),
			expected: "[GAD.QB.123]",
			equal:    false,
		},
		{
			name:     "short prefix",
			err:      errors.New("foo[A.B.1]bar"),
			expected: "foo[A.B.2]bar",
			equal:    true,
		},
		{
			name:     "error message",
			err:      errors.New("foo[GAD.DAT.123]bar"),
			expected: "foo[GAD.DAT.567]bar",
			equal:    true,
		},
		{
			name:     "error message mismatch",
			err:      errors.New("foo[GAD.DAT.123]bar"),
			expected: "foo[GAD.DAT.567]baz",
			equal:    false,
		},
		{
			name:     "no log prefix",
			err:      errors.New("foo bar"),
			expected: "foo bar",
			equal:    true,
		},
		{
			name:     "no log prefix mismatch",
			err:      errors.New("foo bar"),
			expected: "foo baz",
			equal:    false,
		},
		{
			name:     "multiple brackets",
			err:      errors.New("foo bar [ABC.DEF.123] [QWE.ZXC.456]"),
			expected: "foo bar [ABC.DEF.345] [QWE.ZXC.678]",
			equal:    true,
		},
		{
			name:     "multiple brackets line number mismatch",
			err:      errors.New("foo bar [ABC.DEF.123] [QWE.ZXC.256]"),
			expected: "foo bar [ABC.DEF.123] boo [QWE.ZXC.256]",
			equal:    false,
		},
		{
			name:     "random brackets",
			err:      errors.New("foo ]bar issue[ with '[blah]'"),
			expected: "foo ]bar issue[ with '[blah]'",
			equal:    true,
		},
		{
			name:     "random brackets mismatch",
			err:      errors.New("foo ]bar issue[ with '[blah]'"),
			expected: "foo bar issue[ with '[blah]'",
			equal:    false,
		},
		{
			name:     "db error prefix match",
			err:      errors.New("dberr123"),
			expected: "dberr123",
			equal:    true,
		},
		{
			name:     "db error prefix mismatch",
			err:      errors.New("dberr123"),
			expected: "dberr456",
			equal:    false,
		},
		{
			name:     "match with multiple db errors",
			err:      errors.New("[GAD.DAT.123] failed with dberr_123 message (dberr_456)"),
			expected: "[GAD.DAT.123] failed with dberr_123456 message (dberr_foobar)",
			equal:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert1.New(t)
			if tt.equal {
				EqualLogError(assert, tt.err, tt.expected)
			} else {
				mockAssert := &MockAssertion{}
				// if log errors are incorrectly equal, run again using the real
				// assert obj so test output shows useful information
				if EqualLogError(mockAssert, tt.err, tt.expected) {
					EqualLogError(assert, tt.err, tt.expected)
				}
			}
		})
	}
}
