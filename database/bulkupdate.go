package database

import (
	dberrors "github.com/beaconsoftwarellc/gadget/v2/database/errors"
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
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
	columns []qb.TableField
}

func (api *bulkUpdate[T]) Update(objs ...T) {
	for _, obj := range objs {
		obj.Initialize()
		api.pending = append(api.pending, obj)
	}
}

func (api *bulkUpdate[T]) Commit() errors.TracerError {
	if nil == api.tx {
		return errors.New("commit called on nil transaction")
	}
	defer func() {
		api.pending = make([]T, 0)
		api.tx = nil
	}()
	if len(api.pending) == 0 {
		return api.tx.Commit()
	}
	var (
		log = api.configuration.Logger()
		// grab a single instance to create the parameterized sql
		obj            = api.pending[0]
		query          = qb.Update(obj.Meta())
		namedStatement transaction.NamedStatement
		tracerErr      errors.TracerError
		err            error
	)
	// values are inconsequential because we are using named
	// and not order based
	for _, column := range api.columns {
		query.SetParam(column)
	}
	query.Where(obj.Meta().PrimaryKey().Equal(":" + obj.Meta().PrimaryKey().Name))
	sql, err := query.ParameterizedSQL(qb.NoLimit)
	if nil != err {
		_ = log.Error(api.tx.Rollback())
		return errors.Wrap(err)
	}
	namedStatement, err = api.tx.PrepareNamed(sql)
	if nil != err {
		_ = log.Error(api.tx.Rollback())
		return errors.Wrap(err)
	}
	for _, obj := range api.pending {
		_, err = namedStatement.Exec(obj)
		if nil != err {
			_ = log.Error(api.tx.Rollback())
			return dberrors.TranslateError(err, dberrors.Update, sql)
		}
	}
	tracerErr = api.tx.Commit()
	if nil != tracerErr {
		return tracerErr
	}
	err = namedStatement.Close()
	return errors.Wrap(err)
}
