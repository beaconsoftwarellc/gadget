package database

import (
	"database/sql"

	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// CommitRollbackReset is the base interface for Bulk operation
// clients
type CommitRollbackReset interface {
	// Reset the pending records and transaction on this instance
	Reset() errors.TracerError
	// Commit the bulk operation to the database
	Commit() (sql.Result, errors.TracerError)
	// Rollback the bulk operation and close the transaction
	Rollback() errors.TracerError
}

type bulkOperation[T record.Record] struct {
	tx            transaction.Transaction
	pending       []T
	db            *transactable
	configuration Configuration
}

func (bop *bulkOperation[T]) Reset() errors.TracerError {
	if nil != bop.tx {
		return errors.New("transaction should be committed or rolled " +
			"back prior to calling Reset")
	}
	var err error
	bop.pending = make([]T, 0)
	bop.tx, err = transaction.New(bop.db,
		bop.configuration.Logger(),
		bop.configuration.SlowQueryThreshold(),
		bop.configuration.LoggedSlowQueries(),
	)
	return errors.Wrap(err)
}

func (bop *bulkOperation[T]) Rollback() errors.TracerError {
	if nil == bop.tx {
		return errors.New("rollback called on nil transaction")
	}
	defer func() {
		bop.pending = make([]T, 0)
		bop.tx = nil
	}()
	return bop.tx.Rollback()
}
