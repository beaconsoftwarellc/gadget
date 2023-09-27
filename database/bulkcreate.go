package database

import (
	"database/sql"

	dberrors "github.com/beaconsoftwarellc/gadget/v2/database/errors"
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/database/utility"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// BulkCreate allows for bulk creation of a single resource within a
// transaction.
type BulkCreate[T record.Record] interface {
	// Reset the pending records and transaction on this instance
	Reset() errors.TracerError
	// Create initializes the passed  Records and inserts them into the Database
	// as a single request.
	Create(objs ...T)
	// Commit the bulk operation to the database
	Commit() (sql.Result, errors.TracerError)
	// Rollback the bulk operation and close the transaction
	Rollback() errors.TracerError
}

type bulkCreate[T record.Record] struct {
	tx            transaction.Transaction
	pending       []T
	db            *transactable
	configuration Configuration
}

func (api *bulkCreate[T]) Reset() errors.TracerError {
	if nil != api.tx {
		return errors.New("transaction should be committed or rolled " +
			"back prior to calling Reset")
	}
	var err error
	api.pending = make([]T, 0)
	api.tx, err = transaction.New(api.db,
		api.configuration.Logger(),
		api.configuration.SlowQueryThreshold(),
		api.configuration.LoggedSlowQueries(),
	)
	return errors.Wrap(err)
}

func (api *bulkCreate[T]) Create(objs ...T) {
	for _, obj := range objs {
		obj.Initialize()
		api.pending = append(api.pending, obj)
	}
}

func (api *bulkCreate[T]) Commit() (sql.Result, errors.TracerError) {
	if nil == api.tx {
		return nil, errors.New("commit called on nil transaction")
	}
	defer func() {
		api.pending = make([]T, 0)
		api.tx = nil
	}()
	if len(api.pending) == 0 {
		return nil, api.tx.Commit()
	}
	var (
		result    sql.Result
		log       = api.configuration.Logger()
		writeCols = utility.AppendIfMissing(
			api.pending[0].Meta().WriteColumns(),
			api.pending[0].Meta().PrimaryKey(),
		)
		query     = qb.Insert(writeCols...)
		stmt, err = query.ParameterizedSQL()
	)
	if nil != err {
		_ = log.Error(api.tx.Rollback())
		return nil, errors.Wrap(err)
	}
	result, err = api.tx.Implementation().NamedExec(stmt, api.pending)
	if nil != err {
		_ = log.Error(api.tx.Rollback())
		return nil, dberrors.TranslateError(err, dberrors.Insert, stmt)
	}
	return result, api.tx.Commit()
}

func (api *bulkCreate[T]) Rollback() errors.TracerError {
	if nil == api.tx {
		return errors.New("rollback called on nil transaction")
	}
	defer func() {
		api.pending = make([]T, 0)
		api.tx = nil
	}()
	return api.tx.Rollback()
}
