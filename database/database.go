package database

import (
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/jmoiron/sqlx"
)

// Database defines a connection to a database
type Database struct {
	*sqlx.DB
	Configuration Config
	Logger        log.Logger
}

func (db *Database) enforceLimits(options *ListOptions) *ListOptions {
	if nil == options {
		options = NewListOptions(db.Configuration.MaxQueryLimit(), 0)
	} else if db.Configuration.MaxQueryLimit() != qb.NoLimit &&
		options.Limit > db.Configuration.MaxQueryLimit() {
		db.Logger.Warnf("limit %d exceeds max limit of %d", options.Limit,
			db.Configuration.MaxQueryLimit())
		options.Limit = db.Configuration.MaxQueryLimit()
	}
	return options
}

// Create initializes a Record and inserts it into the Database
func (db *Database) Create(obj Record) errors.TracerError {
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	err = db.CreateTx(obj, tx)
	return CommitOrRollback(tx, err, db.Logger)
}

// CreateTx initializes a Record and inserts it into the Database
func (db *Database) CreateTx(obj Record, tx *sqlx.Tx) errors.TracerError {
	var tracerErr errors.TracerError
	var previousPK PrimaryKeyValue
	obj.Initialize()
	for i := 0; i < 5; i++ {
		writeCols := appendIfMissing(obj.Meta().WriteColumns(), obj.Meta().PrimaryKey())
		query := qb.Insert(writeCols...)
		stmt, err := query.ParameterizedSQL()
		if nil != err {
			return errors.Wrap(err)
		}

		_, err = tx.NamedExec(stmt, obj)
		if nil == err {
			return db.ReadTx(obj, obj.PrimaryKey(), tx)
		}
		tracerErr = TranslateError(err, Insert, stmt, db.Logger)
		switch tracerErr.(type) {
		case *DuplicateRecordError:
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

// UpsertTx a new entry into the database for the Record
func (db *Database) UpsertTx(obj Record, tx *sqlx.Tx) errors.TracerError {
	insertCols := appendIfMissing(obj.Meta().ReadColumns(), obj.Meta().PrimaryKey())
	updateCols := make([]qb.TableField, len(obj.Meta().WriteColumns()))
	copy(updateCols, obj.Meta().WriteColumns())
	createdOn := qb.TableField{Name: "created_on", Table: obj.Meta().GetName()}
	if contains(obj.Meta().ReadColumns(), createdOn) {
		updateCols = appendIfMissing(updateCols, createdOn)
	}
	updateOn := qb.TableField{Name: "updated_on", Table: obj.Meta().GetName()}
	if contains(obj.Meta().ReadColumns(), updateOn) {
		updateCols = appendIfMissing(updateCols, updateOn)
	}

	query := qb.Insert(insertCols...).OnDuplicate(updateCols)
	stmt, err := query.ParameterizedSQL()
	if nil != err {
		return errors.Wrap(err)
	}

	_, err = tx.NamedExec(stmt, obj)

	if nil != err {
		return TranslateError(err, Insert, stmt, db.Logger)
	}
	return db.ReadTx(obj, obj.PrimaryKey(), tx)
}

// Read populates a Record from the database
func (db *Database) Read(obj Record, pk PrimaryKeyValue) errors.TracerError {
	tx := db.MustBegin()
	defer tx.Commit() // Since this is a read only no need for a rollback
	return db.ReadTx(obj, pk, tx)
}

// ReadTx populates a Record from the database using a transaction
func (db *Database) ReadTx(obj Record, pk PrimaryKeyValue, tx *sqlx.Tx) errors.TracerError {
	return db.ReadOneWhereTx(obj, tx, obj.Meta().PrimaryKey().Equal(pk.Value()))
}

// ReadOneWhere populates a Record from a custom where clause
func (db *Database) ReadOneWhere(obj Record, condition *qb.ConditionExpression) errors.TracerError {
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	defer tx.Commit() // Since this is a read only no need for a rollback
	return db.ReadOneWhereTx(obj, tx, condition)
}

// ReadOneWhereTx populates a Record from a custom where clause using a transaction
func (db *Database) ReadOneWhereTx(obj Record, tx *sqlx.Tx, condition *qb.ConditionExpression) errors.TracerError {
	stmt, args, err := qb.Select(obj.Meta().AllColumns()).From(obj.Meta()).Where(condition).SQL(1, 0)
	if nil != err {
		return errors.Wrap(err)
	}

	if err = tx.QueryRowx(stmt, args...).StructScan(obj); nil != err {
		return TranslateError(err, Select, stmt, db.Logger)
	}
	return nil
}

// List populates obj with a list of Records from the database
func (db *Database) List(def Record, obj interface{}, options *ListOptions) errors.TracerError {
	options = db.enforceLimits(options)
	stmt, _, err := qb.Select(def.Meta().AllColumns()).
		From(def.Meta()).
		OrderBy(def.Meta().SortBy()).
		SQL(options.Limit, options.Offset)
	if err != nil {
		return errors.Wrap(err)
	}
	if err = db.DB.Select(obj, stmt); nil != err {
		return TranslateError(err, Select, stmt, db.Logger)
	}
	return nil
}

// ListWhere populates obj with a list of Records from the database
func (db *Database) ListWhere(meta Record, target interface{}, condition *qb.ConditionExpression, options *ListOptions) errors.TracerError {
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	options = db.enforceLimits(options)
	tracerErr := db.ListWhereTx(tx, meta, target, condition, options)
	if nil != tracerErr {
		db.Logger.Error(tx.Rollback())
		return tracerErr
	}
	return errors.Wrap(tx.Commit())
}

// ListWhereTx populates target with a list of Records from the database using the transaction
func (db *Database) ListWhereTx(tx *sqlx.Tx, meta Record, target interface{},
	condition *qb.ConditionExpression, options *ListOptions) errors.TracerError {
	options = db.enforceLimits(options)
	stmt, values, err := db.buildListWhere(meta, condition).SQL(options.Limit, options.Offset)
	if nil != err {
		return errors.Wrap(err)
	}
	if err = tx.Select(target, stmt, values...); nil != err {
		return TranslateError(err, Select, stmt, db.Logger)
	}
	return nil
}

func (db *Database) buildListWhere(def Record, condition *qb.ConditionExpression) *qb.SelectQuery {
	return qb.Select(def.Meta().AllColumns()).
		From(def.Meta()).
		Where(condition).
		OrderBy(def.Meta().SortBy())
}

// Select executes a given select query and populates the target
func (db *Database) Select(target interface{}, query *qb.SelectQuery) errors.TracerError {
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	defer tx.Commit() // Since this is a read only no need for a rollback
	return db.SelectTx(tx, target, query)
}

// SelectTx executes a given select query and populates the target
func (db *Database) SelectTx(tx *sqlx.Tx, target interface{}, query *qb.SelectQuery) errors.TracerError {
	stmt, values, err := query.SQL(db.Configuration.MaxQueryLimit(), 0)
	if err != nil {
		return errors.Wrap(err)
	}

	if err = tx.Select(target, stmt, values...); nil != err {
		return TranslateError(err, Select, stmt, db.Logger)
	}
	return nil
}

// SelectList of Records into target based upon the passed query
func (db *Database) SelectList(target interface{}, query *qb.SelectQuery,
	options *ListOptions) errors.TracerError {
	options = db.enforceLimits(options)
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	tracerErr := db.SelectListTx(tx, target, query, options)
	if nil != tracerErr {
		log.Error(tx.Rollback())
		return tracerErr
	}
	return errors.Wrap(tx.Commit())
}

// SelectListTx of Records into target in a transaction based upon the passed query
func (db *Database) SelectListTx(tx *sqlx.Tx, target interface{}, query *qb.SelectQuery,
	options *ListOptions) errors.TracerError {
	options = db.enforceLimits(options)
	stmt, values, err := query.SQL(options.Limit, options.Offset)
	if err != nil {
		return errors.Wrap(err)
	}
	if err = tx.Select(target, stmt, values...); nil != err {
		return TranslateError(err, Select, stmt, db.Logger)
	}
	return nil
}

// Update replaces an entry in the database for the Record
func (db *Database) Update(obj Record) errors.TracerError {
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	err = db.UpdateTx(obj, tx)
	return CommitOrRollback(tx, err, db.Logger)
}

// UpdateTx replaces an entry in the database for the Record using a transaction
func (db *Database) UpdateTx(obj Record, tx *sqlx.Tx) errors.TracerError {
	query := qb.Update(obj.Meta())
	for _, col := range obj.Meta().WriteColumns() {
		query.SetParam(col)
	}
	query.Where(obj.Meta().PrimaryKey().Equal(":" + obj.Meta().PrimaryKey().GetName()))
	stmt, err := query.ParameterizedSQL(1)
	if nil != err {
		return errors.Wrap(err)
	}

	_, err = tx.NamedExec(stmt, obj)
	if nil != err {
		return TranslateError(err, Update, stmt, db.Logger)
	}

	return db.ReadTx(obj, obj.PrimaryKey(), tx)
}

// Delete removes a row from the database
func (db *Database) Delete(obj Record) errors.TracerError {
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	err = db.DeleteTx(obj, tx)
	return CommitOrRollback(tx, err, db.Logger)
}

// DeleteTx removes a row from the database using a transaction
func (db *Database) DeleteTx(obj Record, tx *sqlx.Tx) errors.TracerError {
	where := obj.Meta().PrimaryKey().Equal(obj.PrimaryKey().Value())
	return db.DeleteWhereTx(obj, tx, where)
}

// DeleteWhere removes a row(s) from the database based on a supplied where clause
func (db *Database) DeleteWhere(obj Record, where *qb.ConditionExpression) errors.TracerError {
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	err = db.DeleteWhereTx(obj, tx, where)
	return CommitOrRollback(tx, err, db.Logger)
}

// DeleteWhereTx removes row(s) from the database based on a supplied where clause in a transaction
func (db *Database) DeleteWhereTx(obj Record, tx *sqlx.Tx,
	condition *qb.ConditionExpression) errors.TracerError {
	stmt, values, err := qb.Delete(obj.Meta()).Where(condition).SQL()
	if nil != err {
		return errors.Wrap(err)
	}

	_, err = tx.Exec(stmt, values...)

	if nil != err {
		return TranslateError(err, Delete, stmt, db.Logger)
	}
	return nil
}

// UpdateWhere updates fields for the Record based on a supplied where clause
func (db *Database) UpdateWhere(obj Record,
	where *qb.ConditionExpression, fields ...qb.FieldValue) (int64, error) {
	tx, err := db.Beginx()
	if nil != err {
		return 0, errors.Wrap(err)
	}

	rowsAffected, err := db.UpdateWhereTx(obj, tx, where, fields...)
	return rowsAffected, CommitOrRollback(tx, err, db.Logger)
}

// UpdateWhereTx updates fields for the Record based on a supplied where clause in a transaction
func (db *Database) UpdateWhereTx(obj Record, tx *sqlx.Tx,
	where *qb.ConditionExpression, fields ...qb.FieldValue) (int64, error) {
	query := qb.Update(obj.Meta())

	for _, f := range fields {
		query = query.Set(f.Field, f.Value)
	}

	stmt, values, err := query.Where(where).SQL(qb.NoLimit)
	if nil != err {
		return 0, errors.Wrap(err)
	}

	result, err := tx.Exec(stmt, values...)
	if nil != err {
		return 0, TranslateError(err, Update, stmt, db.Logger)
	}

	rowsAffected, err := result.RowsAffected()
	if nil != err {
		return 0, errors.Wrap(err)
	}

	return rowsAffected, nil
}
