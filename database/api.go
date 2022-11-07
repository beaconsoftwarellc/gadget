//go:generate mockgen -source=$GOFILE -package $GOPACKAGE -destination api.mock.gen.go
package database

import (
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/jmoiron/sqlx"
)

// API is a database interface
type API interface {
	// Begin starts a transaction
	Begin() error
	// Commit commits the transaction
	Commit() error
	// Rollback aborts the transaction
	Rollback() error
	// CommitOrRollback will rollback on an errors.TracerError otherwise commit
	CommitOrRollback(err error) error

	// Create initializes a Record and inserts it into the Database
	Create(obj Record) error
	// Read populates a Record from the database
	Read(obj Record, pk PrimaryKeyValue) error
	// ReadOneWhere populates a Record from a custom where clause
	ReadOneWhere(obj Record, condition *qb.ConditionExpression) error
	// Select executes a given select query and populates the target
	Select(target interface{}, query *qb.SelectQuery) error
	// SelectList of Records into target based upon the passed query
	SelectList(target interface{}, query *qb.SelectQuery, options *ListOptions) error
	// ListWhere populates target with a list of records from the database
	ListWhere(meta Record, target interface{}, condition *qb.ConditionExpression, options *ListOptions) error
	// Update replaces an entry in the database for the Record using a transaction
	Update(obj Record) error
	// UpdateWhere updates fields for the Record based on a supplied where clause
	UpdateWhere(obj Record, where *qb.ConditionExpression, fields ...qb.FieldValue) (int64, error)
	// Delete removes a row from the database
	Delete(obj Record) error
	// DeleteWhereTx removes row(s) from the database based on a supplied where clause in a transaction
	DeleteWhere(obj Record, condition *qb.ConditionExpression) error
}

// NewAPI using the passed database and transaction. Transaction may be null
func NewAPI(db *Database, tx *sqlx.Tx) API {
	return &dbapi{tx: tx, db: db}
}

var _ API = &dbapi{}

var ErrMissingTransaction = errors.New("missing transaction")

type dbapi struct {
	tx *sqlx.Tx
	db *Database
}

func (d *dbapi) Begin() error {
	var err error

	if d.tx == nil {
		d.tx, err = d.db.Beginx()
	}

	return err
}

func (d *dbapi) Rollback() error {
	if d.tx != nil {
		err := d.tx.Rollback()
		d.tx = nil
		return err
	}

	return ErrMissingTransaction
}

func (d *dbapi) Commit() error {
	if d.tx != nil {
		err := d.tx.Commit()
		d.tx = nil
		return err
	}

	return ErrMissingTransaction
}

func (d *dbapi) CommitOrRollback(err error) error {
	if d.tx != nil {
		err = CommitOrRollback(d.tx, err)
		d.tx = nil
		return err
	}

	return ErrMissingTransaction
}

func (d *dbapi) Create(obj Record) error {
	return d.runInTransaction(func(tx *sqlx.Tx) error {
		return d.db.CreateTx(obj, d.tx)
	})
}

func (d *dbapi) Read(obj Record, pk PrimaryKeyValue) error {
	return d.runInTransaction(func(tx *sqlx.Tx) error {
		return d.db.ReadTx(obj, pk, d.tx)
	})
}

func (d *dbapi) ReadOneWhere(obj Record, condition *qb.ConditionExpression) error {
	return d.runInTransaction(func(tx *sqlx.Tx) error {
		return d.db.ReadOneWhereTx(obj, d.tx, condition)
	})
}

func (d *dbapi) Select(target interface{}, query *qb.SelectQuery) error {
	return d.runInTransaction(func(tx *sqlx.Tx) error {
		return d.db.SelectTx(d.tx, target, query)
	})
}

func (d *dbapi) SelectList(target interface{}, query *qb.SelectQuery, options *ListOptions) error {
	return d.runInTransaction(func(tx *sqlx.Tx) error {
		return d.db.SelectListTx(d.tx, target, query, options)
	})
}

func (d *dbapi) ListWhere(meta Record, target interface{},
	condition *qb.ConditionExpression, options *ListOptions) error {
	return d.runInTransaction(func(tx *sqlx.Tx) error {
		return d.db.ListWhereTx(d.tx, meta, target, condition, options)
	})
}

func (d *dbapi) Update(obj Record) error {
	return d.runInTransaction(func(tx *sqlx.Tx) error {
		return d.db.UpdateTx(obj, d.tx)
	})
}

func (d *dbapi) UpdateWhere(obj Record, where *qb.ConditionExpression, fields ...qb.FieldValue) (int64, error) {
	var (
		total int64
		err   error
	)
	err = d.runInTransaction(func(tx *sqlx.Tx) error {
		total, err = d.db.UpdateWhereTx(obj, d.tx, where, fields...)
		return err
	})

	return total, err
}

func (d *dbapi) Delete(obj Record) error {
	return d.runInTransaction(func(tx *sqlx.Tx) error {
		return d.db.DeleteTx(obj, d.tx)
	})
}

func (d *dbapi) DeleteWhere(obj Record, condition *qb.ConditionExpression) error {
	return d.runInTransaction(func(tx *sqlx.Tx) error {
		return d.db.DeleteWhereTx(obj, d.tx, condition)
	})
}

func (d *dbapi) runInTransaction(fn func(*sqlx.Tx) error) error {
	// if the transaction exists, execute a function in tx context
	if d.tx != nil {
		return fn(d.tx)
	}

	// otherwise create a temporary transaction and commit or rollback after execution
	err := d.Begin()
	if err != nil {
		return err
	}

	return d.CommitOrRollback(fn(d.tx))
}
