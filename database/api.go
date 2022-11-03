//go:generate mockgen -source=$GOFILE -package $GOPACKAGE -destination api.mock.gen.go
package database

import (
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
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

// NewAPI returns API interface implementation
func NewAPI(db *Database, tx *sqlx.Tx) API {
	return &dbapi{Tx: tx, Database: db}
}

var _ API = &dbapi{}

type dbapi struct {
	*sqlx.Tx
	*Database
}

// Begin starts a transaction
func (d *dbapi) Begin() error {
	var err error

	if d.Tx == nil {
		d.Tx, err = d.Beginx()
	}

	return err
}

// Rollback aborts the transaction
func (d *dbapi) Rollback() error {
	if d.Tx != nil {
		defer d.cleanTx()
		return d.Tx.Rollback()
	}

	return nil
}

// Commit commits the transaction
func (d *dbapi) Commit() error {
	if d.Tx != nil {
		defer d.cleanTx()
		return d.Tx.Commit()
	}

	return nil
}

// Create initializes a Record and inserts it into the Database
func (d *dbapi) Create(obj Record) error {
	if d.Tx != nil {
		return d.Database.CreateTx(obj, d.Tx)
	}

	return d.Database.Create(obj)
}

// Read populates a Record from the database
func (d *dbapi) Read(obj Record, pk PrimaryKeyValue) error {
	if d.Tx != nil {
		return d.Database.ReadTx(obj, pk, d.Tx)
	}

	return d.Database.Read(obj, pk)
}

// ReadOneWhere populates a Record from a custom where clause
func (d *dbapi) ReadOneWhere(obj Record, condition *qb.ConditionExpression) error {
	if d.Tx != nil {
		return d.Database.ReadOneWhereTx(obj, d.Tx, condition)
	}

	return d.Database.ReadOneWhere(obj, condition)
}

func (d *dbapi) Select(target interface{}, query *qb.SelectQuery) error {
	if d.Tx != nil {
		return d.Database.SelectTx(d.Tx, target, query)
	}

	return d.Database.Select(target, query)
}

// SelectList of Records into target based upon the passed query
func (d *dbapi) SelectList(target interface{}, query *qb.SelectQuery, options *ListOptions) error {
	if d.Tx != nil {
		return d.Database.SelectListTx(d.Tx, target, query, options)
	}

	return d.Database.SelectList(target, query, options)
}

// ListWhere populates target with a list of records from the database
func (d *dbapi) ListWhere(meta Record, target interface{},
	condition *qb.ConditionExpression, options *ListOptions) error {
	if d.Tx != nil {
		return d.Database.ListWhereTx(d.Tx, meta, target, condition, options)
	}

	return d.Database.ListWhere(meta, target, condition, options)
}

func (d *dbapi) Update(obj Record) error {
	if d.Tx != nil {
		return d.Database.UpdateTx(obj, d.Tx)
	}

	return d.Database.Update(obj)
}

// UpdateWhere updates fields for the Record based on a supplied where clause
func (d *dbapi) UpdateWhere(obj Record, where *qb.ConditionExpression, fields ...qb.FieldValue) (int64, error) {
	if d.Tx != nil {
		return d.Database.UpdateWhereTx(obj, d.Tx, where, fields...)
	}

	return d.Database.UpdateWhere(obj, where, fields...)
}

// Delete removes a row from the database
func (d *dbapi) Delete(obj Record) error {
	if d.Tx != nil {
		return d.Database.DeleteTx(obj, d.Tx)
	}

	return d.Database.Delete(obj)
}

// DeleteWhereTx removes row(s) from the database based on a supplied where clause in a transaction
func (d *dbapi) DeleteWhere(obj Record, condition *qb.ConditionExpression) error {
	if d.Tx != nil {
		return d.Database.DeleteWhereTx(obj, d.Tx, condition)
	}

	return d.Database.DeleteWhere(obj, condition)
}

func (d *dbapi) cleanTx() {
	d.Tx = nil
}
