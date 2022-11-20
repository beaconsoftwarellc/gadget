package deltas

import (
	"database/sql"
	"testing"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/database"
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
)

type MockLockingDatabase struct {
	AcquireNamedLockReturn bool
	AcquireNamedLockErr    errors.TracerError
	ReleaseNamedLockErr    errors.TracerError
	BeginxReturn           lockingDatabaseTx
	BeginxErr              error
	CreateTxErr            errors.TracerError
	CloseErr               error
	ReadOneWhereTxErr      errors.TracerError
	TableExistsReturn      bool
	TableExistsErr         errors.TracerError
}

func (mld *MockLockingDatabase) AcquireNamedLock(name string, timeout time.Duration) (bool, errors.TracerError) {
	return mld.AcquireNamedLockReturn, mld.AcquireNamedLockErr
}

func (mld *MockLockingDatabase) ReleaseNamedLock(name string) errors.TracerError {
	return mld.ReleaseNamedLockErr
}

func (mld *MockLockingDatabase) Beginx() (lockingDatabaseTx, error) {
	return mld.BeginxReturn, mld.BeginxErr
}

func (mld *MockLockingDatabase) CreateTx(record database.Record) errors.TracerError {
	return mld.CreateTxErr
}

func (mld *MockLockingDatabase) Close() error {
	return mld.CloseErr
}

func (mld *MockLockingDatabase) ReadOneWhereTx(record database.Record,
	condition *qb.ConditionExpression) errors.TracerError {
	return mld.ReadOneWhereTxErr
}

func (mld *MockLockingDatabase) TableExists(schema, name string) (bool, error) {
	return mld.TableExistsReturn, mld.TableExistsErr
}

type MockTransaction struct {
	ExecReturn  sql.Result
	ExecErr     error
	RollbackErr error
	CommitErr   error
}

func (mt *MockTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return mt.ExecReturn, mt.ExecErr
}

func (mt *MockTransaction) Rollback() error {
	return mt.RollbackErr
}

func (mt *MockTransaction) Commit() error {
	return mt.CommitErr
}

func TestExecute(t *testing.T) {
	assert := assert1.New(t)
	mld := &MockLockingDatabase{
		AcquireNamedLockReturn: true,
		TableExistsReturn:      true,
		BeginxReturn:           &MockTransaction{},
	}

	deltas := []*Delta{
		{
			ID:   1,
			Name: "test",
		},
	}

	err := execute(database.InstanceConfig{}, "", deltas, mld, log.NewStackLogger())
	assert.NoError(err)
}

func TestExecute_BeginxError(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))
	mld := &MockLockingDatabase{
		AcquireNamedLockReturn: true,
		BeginxErr:              expected,
	}

	err := execute(database.InstanceConfig{}, "", nil, mld, log.NewStackLogger())
	assert.EqualError(err, expected.Error())
}

func TestExecute_TableExistsError(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))
	mld := &MockLockingDatabase{
		AcquireNamedLockReturn: true,
		BeginxReturn:           &MockTransaction{},
		TableExistsErr:         expected,
	}

	err := execute(database.InstanceConfig{}, "", nil, mld, log.NewStackLogger())
	assert.EqualError(err, expected.Error())
}

func TestExecute_ExecError(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))
	mld := &MockLockingDatabase{
		AcquireNamedLockReturn: true,
		BeginxReturn:           &MockTransaction{ExecErr: expected},
	}

	err := execute(database.InstanceConfig{}, "", nil, mld, log.NewStackLogger())
	assert.EqualError(err, expected.Error())
}

func TestExecute_ExecuteDeltaError(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))
	mld := &MockLockingDatabase{
		AcquireNamedLockReturn: true,
		BeginxReturn:           &MockTransaction{},
		ReadOneWhereTxErr:      expected,
	}

	deltas := []*Delta{
		{
			ID:   1,
			Name: "test",
		},
	}

	err := execute(database.InstanceConfig{}, "", deltas, mld, log.NewStackLogger())
	assert.EqualError(err, expected.Error())
}

func TestExecuteDelta(t *testing.T) {
	assert := assert1.New(t)
	mld := &MockLockingDatabase{
		ReadOneWhereTxErr: database.NewNotFoundError(),
	}

	mt := &MockTransaction{}

	delta := &Delta{
		ID:   1,
		Name: "test",
	}

	err := ExecuteDelta(mt, mld, delta, log.NewStackLogger())
	assert.NoError(err)
}

func TestExecuteDelta_AlreadyExecuted(t *testing.T) {
	assert := assert1.New(t)
	mld := &MockLockingDatabase{}
	mt := &MockTransaction{}

	delta := &Delta{
		ID:   1,
		Name: "test",
	}

	err := ExecuteDelta(mt, mld, delta, log.NewStackLogger())
	assert.NoError(err)
}

func TestExecuteDelta_ReadError(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))
	mld := &MockLockingDatabase{
		ReadOneWhereTxErr: expected,
	}

	mt := &MockTransaction{}

	delta := &Delta{
		ID:   1,
		Name: "test",
	}

	err := ExecuteDelta(mt, mld, delta, log.NewStackLogger())
	assert.EqualError(err, expected.Error())
}

func TestExecuteDelta_ExecError(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))
	mld := &MockLockingDatabase{
		ReadOneWhereTxErr: database.NewNotFoundError(),
	}

	mt := &MockTransaction{
		ExecErr: expected,
	}

	delta := &Delta{
		ID:   1,
		Name: "test",
	}

	err := ExecuteDelta(mt, mld, delta, log.NewStackLogger())
	assert.EqualError(err, expected.Error())
}

func TestExecuteDelta_CreateTxError(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))
	mld := &MockLockingDatabase{
		ReadOneWhereTxErr: database.NewNotFoundError(),
		CreateTxErr:       expected,
	}

	mt := &MockTransaction{}

	delta := &Delta{
		ID:   1,
		Name: "test",
	}

	err := ExecuteDelta(mt, mld, delta, log.NewStackLogger())
	assert.EqualError(err, expected.Error())
}
