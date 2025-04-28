//go:generate mockgen -source=$GOFILE -package $GOPACKAGE -destination api.mock.gen.go
package database

import (
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/database/utility"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

const defaultSlowQueryThreshold = 100 * time.Millisecond

// API is a database interface
type API interface {
	// Begin starts a transaction
	Begin() errors.TracerError
	// GetTransaction that is currently on this instance, Begin must be called first.
	GetTransaction() transaction.Transaction
	// Commit commits the transaction
	Commit() errors.TracerError
	// Rollback aborts the transaction
	Rollback() errors.TracerError
	// CommitOrRollback will rollback on an errors.TracerError otherwise commit
	CommitOrRollback(err error) errors.TracerError

	// Count the number of rows in the passed query
	Count(qb.Table, *qb.SelectQuery) (int32, error)
	// CountWhere rows match the passed condition in the specified table. Condition
	// may be nil in order to just count the table rows.
	CountWhere(qb.Table, *qb.ConditionExpression) (int32, error)
	// Sum calculates the total of the specified numeric field over all rows
	// matching the given select query.
	Sum(qb.TableField, *qb.SelectQuery) (int32, error)
	// Create initializes a Record and inserts it into the Database
	Create(obj record.Record) errors.TracerError
	// Read populates a Record from the database
	Read(obj record.Record, pk record.PrimaryKeyValue) errors.TracerError
	// ReadOneWhere populates a Record from a custom where clause
	ReadOneWhere(obj record.Record, condition *qb.ConditionExpression) errors.TracerError
	// Select executes a given select query and populates the target
	Select(target interface{}, query *qb.SelectQuery, options qb.LimitOffset) errors.TracerError
	// ListWhere populates target with a list of records from the database
	ListWhere(meta record.Record, target interface{},
		condition *qb.ConditionExpression, options qb.LimitOffset) errors.TracerError
	// Update replaces an entry in the database for the Record using a transaction
	Update(obj record.Record) errors.TracerError
	// UpdateWhere updates fields for the Record based on a supplied where clause
	UpdateWhere(obj record.Record, where *qb.ConditionExpression,
		fields ...qb.FieldValue) (int64, errors.TracerError)
	// Delete removes a row from the database
	Delete(obj record.Record) errors.TracerError
	// DeleteWhere removes row(s) from the database based on a supplied where
	// clause in a transaction
	DeleteWhere(obj record.Record, condition *qb.ConditionExpression) errors.TracerError
}

// SelectWithTotal executes the select query populating target and returning the
// total records possible
func SelectWithTotal[T any](db API, table qb.Table, target T,
	query *qb.SelectQuery, options qb.LimitOffset) (T, int, error) {
	var (
		total int32
		err   error
	)
	if total, err = db.Count(table, query); err != nil || options.Limit() == 0 || total == 0 {
		return target, int(total), err
	}

	err = db.Select(&target, query, options)
	if nil != err {
		return target, 0, err
	}
	return target, int(total), err
}

// ErrMissingTransaction is returned when a call requiring a transaction is made
// prior to Begin being called.
var ErrMissingTransaction = errors.New("missing transaction")

type api struct {
	tx            transaction.Transaction
	db            *transactable
	configuration Configuration
}

func (d *api) Begin() errors.TracerError {
	if d.tx != nil {
		return nil
	}
	var err error
	d.tx, err = transaction.New(
		d.db,
		d.configuration.Logger(),
		d.configuration.SlowQueryThreshold(),
		d.configuration.LoggedSlowQueries(),
	)
	return errors.Wrap(err)
}

func (d *api) GetTransaction() transaction.Transaction {
	return d.tx
}

func (d *api) Rollback() errors.TracerError {
	if d.tx != nil {
		err := d.tx.Rollback()
		d.tx = nil
		return err
	}

	return ErrMissingTransaction
}

func (d *api) Commit() errors.TracerError {
	if d.tx != nil {
		err := d.tx.Commit()
		d.tx = nil
		return err
	}

	return ErrMissingTransaction
}

func (d *api) CommitOrRollback(err error) errors.TracerError {
	if d.tx != nil {
		err = utility.CommitOrRollback(d.tx, err, d.configuration.Logger())
		d.tx = nil
		return errors.Wrap(err)
	}

	return ErrMissingTransaction
}

func (db *api) Count(table qb.Table, query *qb.SelectQuery) (int32, error) {
	var (
		target []*qb.RowCount
		err    error
	)
	err = db.Select(&target,
		query.SelectFrom(qb.NewCountExpression(table.GetName())),
		qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
	)
	if err != nil {
		return 0, err
	}
	if len(target) == 0 {
		// this is valid if the query has a group by
		return 0, nil
	}
	return int32(target[0].Count), nil
}

func (d *api) CountWhere(table qb.Table, where *qb.ConditionExpression) (int32, error) {
	return d.Count(table,
		qb.Select(qb.NewCountExpression(table.GetName())).
			From(table).
			Where(where))
}

func (db *api) Sum(field qb.TableField, query *qb.SelectQuery) (int32, error) {
	var (
		target []*qb.SumResult
		err    error
	)
	err = db.Select(&target,
		query.SelectFrom(qb.NewSumExpression(field.Table, field)),
		qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
	)
	if err != nil {
		return 0, err
	}
	if len(target) == 0 || target[0].Sum == nil {
		return 0, nil
	}
	return int32(*target[0].Sum), nil
}

func (d *api) Create(obj record.Record) errors.TracerError {
	return d.runInTransaction(func(tx transaction.Transaction) errors.TracerError {
		return tx.Create(obj)
	})
}

func (d *api) Read(obj record.Record, pk record.PrimaryKeyValue) errors.TracerError {
	return d.runInTransaction(func(tx transaction.Transaction) errors.TracerError {
		return tx.Read(obj, pk)
	})
}

func (d *api) ReadOneWhere(obj record.Record, condition *qb.ConditionExpression) errors.TracerError {
	return d.runInTransaction(func(tx transaction.Transaction) errors.TracerError {
		return tx.ReadOneWhere(obj, condition)
	})
}

func (d *api) Select(target any, query *qb.SelectQuery,
	options qb.LimitOffset) errors.TracerError {
	options = d.enforceLimits(options)
	return d.runInTransaction(func(tx transaction.Transaction) errors.TracerError {
		return tx.Select(target, query, options)
	})
}

func (d *api) ListWhere(meta record.Record, target interface{},
	condition *qb.ConditionExpression, options qb.LimitOffset) errors.TracerError {
	options = d.enforceLimits(options)
	return d.runInTransaction(func(tx transaction.Transaction) errors.TracerError {
		return tx.ListWhere(meta, target, condition, options)
	})
}

func (d *api) Update(obj record.Record) errors.TracerError {
	return d.runInTransaction(func(tx transaction.Transaction) errors.TracerError {
		return tx.Update(obj)
	})
}

func (d *api) UpdateWhere(obj record.Record, where *qb.ConditionExpression,
	fields ...qb.FieldValue) (int64, errors.TracerError) {
	var (
		total int64
		err   errors.TracerError
	)
	err = d.runInTransaction(func(tx transaction.Transaction) errors.TracerError {
		total, err = tx.UpdateWhere(obj, where, fields...)
		return err
	})

	return total, err
}

func (d *api) Delete(obj record.Record) errors.TracerError {
	return d.runInTransaction(func(tx transaction.Transaction) errors.TracerError {
		return tx.Delete(obj)
	})
}

func (d *api) DeleteWhere(obj record.Record, condition *qb.ConditionExpression) errors.TracerError {
	return d.runInTransaction(func(tx transaction.Transaction) errors.TracerError {
		return tx.DeleteWhere(obj, condition)
	})
}

func (d *api) enforceLimits(options qb.LimitOffset) qb.LimitOffset {
	if options == nil {
		options = qb.NewLimitOffset[uint]()
	}
	if d.configuration.MaxQueryLimit() != qb.NoLimit &&
		options.Limit() > d.configuration.MaxQueryLimit() {
		d.configuration.Logger().Warnf("limit %d exceeds max limit of %d", options.Limit,
			d.configuration.MaxQueryLimit())
		options = qb.NewLimitOffset[uint]().SetOffset(options.Offset()).
			SetLimit(d.configuration.MaxQueryLimit())
	}
	return options
}

func (d *api) runInTransaction(fn func(transaction.Transaction) errors.TracerError) errors.TracerError {
	var (
		err    errors.TracerError
		commit bool
	)
	if d.tx == nil {
		commit = true
		err = d.Begin()
	}
	if nil != err {
		return err
	}

	err = fn(d.tx)

	if commit {
		err = d.CommitOrRollback(err)
	}
	return err
}
