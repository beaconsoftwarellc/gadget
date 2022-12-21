package database

import (
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/jmoiron/sqlx"
)

type Transaction interface {
	// Begin starts a transaction
	Begin() error
	// Commit commits the transaction
	Commit() error
	// Rollback aborts the transaction
	Rollback() error
	// Open reports if transation is open
	Open() bool
	// RunInTransaction
	RunInTransaction(func(tx *sqlx.Tx) error) error
	// CommitOrRollback will rollback on an errors.TracerError otherwise commit
	CommitOrRollback(err error) error
}

type txAPI struct {
	id string
	db *Database
	tx *sqlx.Tx
}

func NewTxAPI(db *Database, tx *sqlx.Tx) *txAPI {
	txAPI := &txAPI{tx: tx, db: db}
	txAPI.generateID()

	return txAPI
}

func (d *txAPI) generateID() {
	d.id = generator.ID("tx")
}

func (d *txAPI) Begin() error {
	var err error

	if !d.Open() {
		d.tx, err = d.db.Beginx()
		d.generateID()
	}

	return err
}

func (d *txAPI) Rollback() error {
	if d.Open() {
		err := d.tx.Rollback()
		d.tx = nil
		return err
	}

	return ErrMissingTransaction
}

func (d *txAPI) Commit() error {
	if d.Open() {
		err := d.tx.Commit()
		d.tx = nil
		return err
	}

	return ErrMissingTransaction
}

func (d *txAPI) CommitOrRollback(err error) error {
	if d.Open() {
		err = CommitOrRollback(d.tx, err, d.db.Logger)
		d.tx = nil
		return err
	}

	return ErrMissingTransaction
}

func (d *txAPI) Open() bool {
	return d.tx != nil
}

func (d *txAPI) RunInTransaction(fn func(*sqlx.Tx) error) error {
	var (
		err    error
		commit bool
	)

	if !d.Open() {
		commit = true
		if err = d.Begin(); err != nil {
			return err
		}
	}

	err = fn(d.tx)

	if commit {
		err = d.CommitOrRollback(err)
	}

	return err
}
