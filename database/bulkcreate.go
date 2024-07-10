package database

//go:generate mockgen -source=$GOFILE -package mocks -destination mocks/bulkcreate.mock.gen.go
import (
	"database/sql"

	dberrors "github.com/beaconsoftwarellc/gadget/v2/database/errors"
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/utility"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// BulkCreate allows for bulk creation of a single resource within a
// transaction.
type BulkCreate[T record.Record] interface {
	CommitRollbackReset
	// Create initializes the passed Records and buffers them pending
	// the call to commit
	Create(objs ...T)
}

type bulkCreate[T record.Record] struct {
	*bulkOperation[T]
	upsert bool
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
		stmt      string
		err       error
		log       = api.configuration.Logger()
		writeCols = utility.AppendIfMissing(
			api.pending[0].Meta().WriteColumns(),
			api.pending[0].Meta().PrimaryKey(),
		)
		query = qb.Insert(writeCols...)
	)
	if api.upsert {
		query.OnDuplicate(api.pending[0].Meta().WriteColumns())
	}
	stmt, err = query.ParameterizedSQL()
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
