//go:generate mockgen -source=$GOFILE -package $GOPACKAGE -destination transaction.mock.gen.go
package transaction

import (
	"fmt"
	"time"

	dberrors "github.com/beaconsoftwarellc/gadget/v2/database/errors"
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/samber/lo"
)

// Transaction is an in-progress database transaction.
//
// A transaction must end with a call to Commit or Rollback.
//
// After a call to Commit or Rollback, all operations on the
// transaction fail with ErrTxDone.
type Transaction interface {
	// Create initializes a Record and inserts it into the Database
	Create(record.Record) errors.TracerError
	// Upsert a new entry into the database for the Record
	Upsert(record.Record) errors.TracerError
	// Read populates a Record from the database
	Read(record.Record, record.PrimaryKeyValue) errors.TracerError
	// ReadOneWhere populates a Record from a custom where clause
	ReadOneWhere(record.Record, *qb.ConditionExpression) errors.TracerError
	// List populates obj with a list of Records from the database
	// TODO: [COR-586] we can expand the Record interface to return an collection
	// 		 of its type so we don't have to pass this clumsily
	List(record.Record, any, record.ListOptions) errors.TracerError
	// ListWhere populates target with a list of Records from the database
	// TODO: [COR-586] we can expand the Record interface to return a collection
	// 		 of its type so we don't have to pass this clumsily
	ListWhere(record.Record, any, *qb.ConditionExpression, record.ListOptions) errors.TracerError
	// Select executes a given select query and populates the target
	// TODO: [COR-586] we can expand the Record interface to return a collection
	// 		 of its type so we don't have to pass this clumsily
	Select(any, *qb.SelectQuery, record.ListOptions) errors.TracerError
	// Update replaces an entry in the database for the Record
	Update(record.Record) errors.TracerError
	// UpdateWhere updates fields for the Record based on a supplied where clause in a transaction
	UpdateWhere(record.Record, *qb.ConditionExpression, ...qb.FieldValue) (int64, errors.TracerError)
	// Delete removes a row from the database
	Delete(record.Record) errors.TracerError
	// DeleteWhere removes row(s) from the database based on a supplied where clause
	DeleteWhere(record.Record, *qb.ConditionExpression) errors.TracerError
	// Commit this transaction
	Commit() errors.TracerError
	// Rollback this transaction
	Rollback() errors.TracerError

	// Implementation that is backing this transaction
	Implementation() Implementation
}

// New transaction that will log query executions that are slower than the passed
// duration
func New(db Begin, logger log.Logger, slow time.Duration) (Transaction, error) {
	tx, err := db.Begin()
	if nil != err {
		return nil, err
	}
	implementation := &slowQueryLoggerTx{
		implementation: tx,
		slow:           slow,
		id:             generator.ID("TX"),
	}
	return &transaction{implementation: implementation}, nil
}

type transaction struct {
	implementation Implementation
}

func (tx *transaction) Implementation() Implementation {
	return tx.implementation
}

func (tx *transaction) Create(obj record.Record) errors.TracerError {
	var tracerErr errors.TracerError
	var previousPK record.PrimaryKeyValue
	obj.Initialize()
	for i := 0; i < 5; i++ {
		writeCols := appendIfMissing(obj.Meta().WriteColumns(), obj.Meta().PrimaryKey())
		query := qb.Insert(writeCols...)
		stmt, err := query.ParameterizedSQL()
		if nil != err {
			return errors.Wrap(err)
		}

		_, err = tx.implementation.NamedExec(stmt, obj)
		if nil == err {
			return tx.Read(obj, obj.PrimaryKey())
		}
		tracerErr = dberrors.TranslateError(err, dberrors.Insert, stmt)
		switch tracerErr.(type) {
		case *dberrors.DuplicateRecordError:
			previousPK = obj.PrimaryKey()
			obj.Initialize()

			if previousPK == obj.PrimaryKey() {
				return tracerErr
			}
			continue
		default:
			return tracerErr
		}
	}
	return tracerErr
}

func (tx *transaction) Upsert(obj record.Record) errors.TracerError {
	insertCols := appendIfMissing(obj.Meta().ReadColumns(), obj.Meta().PrimaryKey())
	updateCols := make([]qb.TableField, len(obj.Meta().WriteColumns()))
	copy(updateCols, obj.Meta().WriteColumns())
	createdOn := qb.TableField{Name: "created_on", Table: obj.Meta().GetName()}

	if lo.Contains(obj.Meta().ReadColumns(), createdOn) {
		updateCols = appendIfMissing(updateCols, createdOn)
	}
	updateOn := qb.TableField{Name: "updated_on", Table: obj.Meta().GetName()}
	if lo.Contains(obj.Meta().ReadColumns(), updateOn) {
		updateCols = appendIfMissing(updateCols, updateOn)
	}

	query := qb.Insert(insertCols...).OnDuplicate(updateCols)
	stmt, err := query.ParameterizedSQL()
	if nil != err {
		return errors.Wrap(err)
	}

	_, err = tx.implementation.NamedExec(stmt, obj)

	if nil != err {
		return dberrors.TranslateError(err, dberrors.Insert, stmt)
	}
	return tx.Read(obj, obj.PrimaryKey())
}

func (tx *transaction) Read(obj record.Record, pk record.PrimaryKeyValue) errors.TracerError {
	return tx.ReadOneWhere(obj, obj.Meta().PrimaryKey().Equal(pk.Value()))
}

func (tx *transaction) ReadOneWhere(obj record.Record, condition *qb.ConditionExpression) errors.TracerError {
	stmt, args, err := qb.Select(obj.Meta().AllColumns()).From(obj.Meta()).Where(condition).SQL(1, 0)
	if nil != err {
		return errors.Wrap(err)
	}
	if err = tx.implementation.QueryRowx(stmt, args...).StructScan(obj); nil != err {
		return dberrors.TranslateError(err, dberrors.Select, fmt.Sprintf("%s %% %v", stmt, args))
	}
	return nil
}

func (tx *transaction) List(def record.Record, obj any,
	options record.ListOptions) errors.TracerError {
	stmt, _, err := qb.Select(def.Meta().AllColumns()).
		From(def.Meta()).
		OrderBy(def.Meta().SortBy()).
		SQL(options.Limit, options.Offset)
	if err != nil {
		return errors.Wrap(err)
	}
	if err = tx.implementation.Select(obj, stmt); nil != err {
		return dberrors.TranslateError(err, dberrors.Select, stmt)
	}
	return nil
}

func (tx *transaction) ListWhere(meta record.Record, target interface{},
	condition *qb.ConditionExpression, options record.ListOptions) errors.TracerError {
	stmt, values, err := tx.buildListWhere(meta, condition).SQL(options.Limit, options.Offset)
	if nil != err {
		return errors.Wrap(err)
	}
	if err = tx.implementation.Select(target, stmt, values...); nil != err {
		return dberrors.TranslateError(err, dberrors.Select, stmt)
	}
	return nil
}

func (tx *transaction) buildListWhere(def record.Record, condition *qb.ConditionExpression) *qb.SelectQuery {
	return qb.Select(def.Meta().AllColumns()).
		From(def.Meta()).
		Where(condition).
		OrderBy(def.Meta().SortBy())
}

func (tx *transaction) Select(target interface{}, query *qb.SelectQuery,
	options record.ListOptions) errors.TracerError {
	stmt, values, err := query.SQL(options.Limit, options.Offset)
	if err != nil {
		return errors.Wrap(err)
	}

	if err = tx.implementation.Select(target, stmt, values...); nil != err {
		return dberrors.TranslateError(err, dberrors.Select, stmt)
	}
	return nil
}

func (tx *transaction) Update(obj record.Record) errors.TracerError {
	query := qb.Update(obj.Meta())
	for _, col := range obj.Meta().WriteColumns() {
		query.SetParam(col)
	}
	query.Where(obj.Meta().PrimaryKey().Equal(":" + obj.Meta().PrimaryKey().GetName()))
	stmt, err := query.ParameterizedSQL(1)
	if nil != err {
		return errors.Wrap(err)
	}

	_, err = tx.implementation.NamedExec(stmt, obj)
	if nil != err {
		return dberrors.TranslateError(err, dberrors.Update, stmt)
	}

	return tx.Read(obj, obj.PrimaryKey())
}

func (tx *transaction) Delete(obj record.Record) errors.TracerError {
	where := obj.Meta().PrimaryKey().Equal(obj.PrimaryKey().Value())
	return tx.DeleteWhere(obj, where)
}

func (tx *transaction) DeleteWhere(obj record.Record,
	condition *qb.ConditionExpression) errors.TracerError {
	stmt, values, err := qb.Delete(obj.Meta()).Where(condition).SQL()
	if nil != err {
		return errors.Wrap(err)
	}

	_, err = tx.implementation.Exec(stmt, values...)

	if nil != err {
		return dberrors.TranslateError(err, dberrors.Delete, stmt)
	}
	return nil
}

func (tx *transaction) UpdateWhere(obj record.Record,
	where *qb.ConditionExpression, fields ...qb.FieldValue) (int64, errors.TracerError) {
	query := qb.Update(obj.Meta())

	for _, f := range fields {
		query = query.Set(f.Field, f.Value)
	}

	stmt, values, err := query.Where(where).SQL(qb.NoLimit)
	if nil != err {
		return 0, errors.Wrap(err)
	}

	result, err := tx.implementation.Exec(stmt, values...)
	if nil != err {
		return 0, dberrors.TranslateError(err, dberrors.Update, stmt)
	}

	rowsAffected, err := result.RowsAffected()
	if nil != err {
		return 0, errors.Wrap(err)
	}

	return rowsAffected, nil
}

func (tx *transaction) Commit() errors.TracerError {
	return errors.Wrap(tx.implementation.Commit())
}

func (tx *transaction) Rollback() errors.TracerError {
	return errors.Wrap(tx.implementation.Rollback())
}
