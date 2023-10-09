package database

import (
	"database/sql"

	dberrors "github.com/beaconsoftwarellc/gadget/v2/database/errors"
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// BulkUpdate allows for bulk updating of a single resource within a
// transaction.
type BulkUpdate[T record.Record] interface {
	CommitRollbackReset
	// Update buffers the update statements until commit is
	// called on this instance.
	Update(objs ...T)
}

type bulkUpdate[T record.Record] struct {
	*bulkOperation[T]
}

func (api *bulkUpdate[T]) Update(objs ...T) {
	for _, obj := range objs {
		obj.Initialize()
		api.pending = append(api.pending, obj)
	}
}

func (api *bulkUpdate[T]) Commit() (sql.Result, errors.TracerError) {
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
		result sql.Result
		log    = api.configuration.Logger()
		// grab a single instance to create the parameterized sql
		obj   = api.pending[0]
		query = qb.Update(obj.Meta())
	)
	for _, col := range obj.Meta().WriteColumns() {
		query.SetParam(col)
	}
	query.Where(obj.Meta().PrimaryKey().
		Equal(":" + obj.Meta().PrimaryKey().GetName()))
	stmt, err := query.ParameterizedSQL(qb.NoLimit)
	if nil != err {
		_ = log.Error(api.tx.Rollback())
		return nil, errors.Wrap(err)
	}
	_, err = api.tx.Implementation().NamedExec(stmt, api.pending)
	if nil != err {
		_ = log.Error(api.tx.Rollback())
		return nil, dberrors.TranslateError(err, dberrors.Update, stmt)
	}
	return result, api.tx.Commit()
}

func (api *bulkUpdate[T]) Rollback() errors.TracerError {
	if nil == api.tx {
		return errors.New("rollback called on nil transaction")
	}
	defer func() {
		api.pending = make([]T, 0)
		api.tx = nil
	}()
	return api.tx.Rollback()
}
