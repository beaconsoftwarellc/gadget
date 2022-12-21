//go:generate mockgen -source=$GOFILE -package $GOPACKAGE -destination api.mock.gen.go
package database

import (
	"database/sql"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/jmoiron/sqlx"
)

type Transaction interface {
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...any) (sql.Result, error)
	Commit() error
	Rollback() error
}

type SlowQueryLoggerTx struct {
	*sqlx.Tx
	slow time.Duration
	log  log.Logger
	id   string
}

func (tx *SlowQueryLoggerTx) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.Tx.NamedQuery(query, arg)
}

func (tx *SlowQueryLoggerTx) NamedExec(query string, arg interface{}) (sql.Result, error) {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.Tx.NamedExec(query, arg)
}

func (tx *SlowQueryLoggerTx) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.Tx.QueryRowx(query, args...)
}

func (tx *SlowQueryLoggerTx) Select(dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.Tx.Select(dest, query, args...)
}

func (tx *SlowQueryLoggerTx) Exec(query string, args ...any) (sql.Result, error) {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.Tx.Exec(query, args...)
}

func (tx *SlowQueryLoggerTx) logSlow(query string, elapsed time.Duration) {
	err := errors.New(
		"[%s] query execution time: %s query: %s", tx.id, elapsed, query)
	if elapsed > tx.slow {
		_ = tx.log.Error(err)
	} else {
		_ = tx.log.Debug(err)
	}
}

// API is a database interface
type API interface {
	// SetSlowQueryDuration threshold, logs errors for queries above a threshold
	SetSlowQueryDuration(threshold time.Duration)

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

const defaultSlowQueryThreshold = 100 * time.Millisecond

// NewAPI using the passed database and transaction. Transaction may be null
func NewAPI(db *Database, tx Transaction, log log.Logger) API {
	return &dbapi{tx: tx, db: db, log: log, slowQueryThreshold: defaultSlowQueryThreshold}
}

var _ API = &dbapi{}

var ErrMissingTransaction = errors.New("missing transaction")

type dbapi struct {
	tx                 Transaction
	txID               string
	db                 *Database
	log                log.Logger
	slowQueryThreshold time.Duration
}

func (d *dbapi) SetSlowQueryDuration(threshold time.Duration) {
	d.slowQueryThreshold = threshold
}

func (d *dbapi) Begin() error {
	if d.tx != nil {
		return nil
	}
	tx, err := d.db.Beginx()
	if nil == err {
		d.tx = &SlowQueryLoggerTx{
			id:   generator.ID("tx"),
			log:  d.log,
			slow: d.slowQueryThreshold,
			Tx:   tx,
		}
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
		err = CommitOrRollback(d.tx, err, d.log)
		d.tx = nil
		return err
	}

	return ErrMissingTransaction
}

func (d *dbapi) Create(obj Record) error {
	return d.runInTransaction(func(tx Transaction) error {
		return d.db.CreateTx(obj, d.tx)
	})
}

func (d *dbapi) Read(obj Record, pk PrimaryKeyValue) error {
	return d.runInTransaction(func(tx Transaction) error {
		return d.db.ReadTx(obj, pk, d.tx)
	})
}

func (d *dbapi) ReadOneWhere(obj Record, condition *qb.ConditionExpression) error {
	return d.runInTransaction(func(tx Transaction) error {
		return d.db.ReadOneWhereTx(obj, d.tx, condition)
	})
}

func (d *dbapi) Select(target interface{}, query *qb.SelectQuery) error {
	return d.runInTransaction(func(tx Transaction) error {
		return d.db.SelectTx(d.tx, target, query)
	})
}

func (d *dbapi) SelectList(target interface{}, query *qb.SelectQuery, options *ListOptions) error {
	return d.runInTransaction(func(tx Transaction) error {
		return d.db.SelectListTx(d.tx, target, query, options)
	})
}

func (d *dbapi) ListWhere(meta Record, target interface{},
	condition *qb.ConditionExpression, options *ListOptions) error {
	return d.runInTransaction(func(tx Transaction) error {
		return d.db.ListWhereTx(d.tx, meta, target, condition, options)
	})
}

func (d *dbapi) Update(obj Record) error {
	return d.runInTransaction(func(tx Transaction) error {
		return d.db.UpdateTx(obj, d.tx)
	})
}

func (d *dbapi) UpdateWhere(obj Record, where *qb.ConditionExpression, fields ...qb.FieldValue) (int64, error) {
	var (
		total int64
		err   error
	)
	err = d.runInTransaction(func(tx Transaction) error {
		total, err = d.db.UpdateWhereTx(obj, d.tx, where, fields...)
		return err
	})

	return total, err
}

func (d *dbapi) Delete(obj Record) error {
	return d.runInTransaction(func(tx Transaction) error {
		return d.db.DeleteTx(obj, d.tx)
	})
}

func (d *dbapi) DeleteWhere(obj Record, condition *qb.ConditionExpression) error {
	return d.runInTransaction(func(tx Transaction) error {
		return d.db.DeleteWhereTx(obj, d.tx, condition)
	})
}

func (d *dbapi) runInTransaction(fn func(Transaction) error) error {
	var (
		err    error
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
