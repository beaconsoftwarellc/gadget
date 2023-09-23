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

func (api *bulkCreate[T]) Create(objs ...T) {
	for _, obj := range objs {
		obj.Initialize()
		api.pending = append(api.pending, obj)
	}
}

func (api *bulkCreate[T]) Commit() (sql.Result, errors.TracerError) {
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
	return result, nil
}

func (api *bulkCreate[T]) Rollback() errors.TracerError {
	return api.tx.Rollback()
}
