package database

import (
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

const (
	// TableExistenceQueryFormat returns a single row and column indicating that the table
	// exists and when it was created. Takes format vars 'table_schema' and 'table_name'
	// NOTE: 'as' clause MUST be there or MySQL will return TABLE_NAME for the column name and
	//  mapping will fail
	// NOTE: Do not use 'create_time' for existence check on tables as it can be NULL
	//  for partitioned tables and is always null in aurora
	TableExistenceQueryFormat = `SELECT TABLE_NAME as "table_name" ` +
		`FROM information_schema.tables` +
		`	WHERE table_schema = '%s'` +
		`	AND table_name = '%s' LIMIT 1;`
	acquireLockQueryFormat = "SELECT GET_LOCK('%s', %d) AS STATUS"
	releaseLockQueryFormat = "SELECT RELEASE_LOCK('%s') AS STATUS"
	// DefaultMaxTries documentation hur
	DefaultMaxTries = 10
)

// Config defines the interface for a config to establish a database connection
type Config interface {
	// DatabaseDialect of SQL
	DatabaseDialect() string
	// DatabaseConnection string for addressing the database
	DatabaseConnection() string
	// NumberOfRetries for the connection before failing
	NumberOfRetries() int
	// WaitBetweenRetries before trying again
	WaitBetweenRetries() time.Duration
}

// TableNameResult is for holding the result row from the existence query
type TableNameResult struct {
	// TableName of the table
	TableName string `db:"table_name"`
}

// StatusResult is for capturing output from a function call on the database, you must use
// an 'AS STATUS' clause in your query in order for mapping to work correctly.
type StatusResult struct {
	// Status as returned by a function call usually
	Status int `db:"STATUS"`
}

// InstanceConfig is a simple struct that satisfies the Config interface
type InstanceConfig struct {
	// Dialect of this instance
	Dialect string
	// Connection string for this instance
	Connection string
	// ConnectRetries is the number of times to retry connecting
	ConnectRetries int
	// ConnectRetryWait is the time to wait between connection retries
	ConnectRetryWait time.Duration
	// DeltaLockMaxTries is used as the maximum retries when attempting to get a Lock during Delta execution
	DeltaLockMaxTries int
	// DeltaLockMinimumCycle is used as the minimum cycle duration when attempting to get a Lock during Delta execution
	DeltaLockMinimumCycle time.Duration
	// DeltaLockMaxCycle is used as the maximum wait time between executions when attempting to get a Lock during Delta execution
	DeltaLockMaxCycle time.Duration
}

// DatabaseDialect indicates the type of SQL this database uses
func (config *InstanceConfig) DatabaseDialect() string {
	return config.Dialect
}

// DatabaseConnection string
func (config *InstanceConfig) DatabaseConnection() string {
	return config.Connection
}

// NumberOfRetries on a connection to the database before failing
func (config *InstanceConfig) NumberOfRetries() int {
	return config.ConnectRetries
}

// WaitBetweenRetries when trying to connect to the database
func (config *InstanceConfig) WaitBetweenRetries() time.Duration {
	if config.ConnectRetryWait == 0 {
		config.ConnectRetryWait = time.Second
	}
	return config.ConnectRetryWait
}

// NumberOfDeltaLockTries on a connection to the database before failing
func (config *InstanceConfig) NumberOfDeltaLockTries() int {
	if config.DeltaLockMaxTries == 0 {
		config.DeltaLockMaxTries = DefaultMaxTries
	}
	return config.DeltaLockMaxTries
}

// MinimumWaitBetweenDeltaLockRetries when trying to connect to the database
func (config *InstanceConfig) MinimumWaitBetweenDeltaLockRetries() time.Duration {
	if config.DeltaLockMinimumCycle == 0 {
		config.DeltaLockMinimumCycle = time.Second
	}
	return config.DeltaLockMinimumCycle
}

// WaitBetweenRetries when trying to connect to the database
func (config *InstanceConfig) MaxWaitBetweenDeltaLockRetries() time.Duration {
	if config.ConnectRetryWait == 0 {
		config.ConnectRetryWait = 10 * time.Second
	}
	return config.ConnectRetryWait
}

// Record defines a database enabled record
type Record interface {
	// Initialize sets any expected values on a Record during Create
	Initialize()
	// PrimaryKey returns the value of the primary key of the Record
	PrimaryKey() PrimaryKeyValue
	// Key returns the name of the primary key of the Record
	Key() string
	// Meta returns the meta object for this Record
	Meta() qb.Table
}

// DefaultRecord implements the Key() as "id"
type DefaultRecord struct{}

// Key returns "ID" as the default Primary Key
func (record *DefaultRecord) Key() string {
	return "id"
}

// PrimaryKeyValue limits keys to string or int
type PrimaryKeyValue struct {
	intPK int
	strPK string
	isInt bool
}

// NewPrimaryKey returns a populated PrimaryKeyValue
func NewPrimaryKey(value interface{}) (pk PrimaryKeyValue) {
	switch t := reflect.TypeOf(value).Kind(); t {
	case reflect.String:
		pk.strPK = value.(string)
	case reflect.Int:
		pk.intPK = value.(int)
		pk.isInt = true
	}
	return
}

// Value returns the string or integer value for a Record
func (pk PrimaryKeyValue) Value() interface{} {
	if pk.isInt {
		return pk.intPK
	}
	return pk.strPK
}

// ListOptions provide limit and filtering capabilities for the List function
type ListOptions struct {
	Limit  uint
	Offset uint
}

// MaxLimit for records returned in an unbounded List request
const MaxLimit = 100

// NewListOptions generates a ListOptions
func NewListOptions(limit uint, offset uint) *ListOptions {
	if 0 == limit {
		limit = 1
	} else if limit > MaxLimit {
		limit = MaxLimit
	}
	return &ListOptions{
		Limit:  limit,
		Offset: offset,
	}
}

// IsNotFoundError returns a boolean indicating that the passed error (can be nil) is of
// type *database.NotFoundError
func IsNotFoundError(err error) bool {
	var dst *NotFoundError
	return errors.As(err, &dst)
}

// TableExists for the passed schema and table name on the passed database
func TableExists(db *Database, schema, name string) (bool, error) {
	var exists bool
	var err error
	var target []*TableNameResult
	err = db.DB.Select(&target, fmt.Sprintf(TableExistenceQueryFormat, schema, name))
	if len(target) == 1 {
		exists = true
	}
	return exists, err
}

// AcquireDatabaseLock with the specified name and timeout. Returns boolean indicating whether the
// lock was acquired or error on failure to execute.
// See: https://dev.mysql.com/doc/refman/5.7/en/locking-functions.html
func AcquireDatabaseLock(db *Database, name string, timeout time.Duration) (bool, errors.TracerError) {
	var err error
	var target []*StatusResult
	err = db.DB.Select(&target, fmt.Sprintf(acquireLockQueryFormat, name, int(timeout.Seconds())))
	if nil != err {
		return false, errors.Wrap(err)
	}
	if len(target) < 1 {
		return false, errors.New("no rows returned from acquire lock query")
	}
	return target[0].Status == 1, nil
}

// ReleaseDatabaseLock with the specified name
// See: https://dev.mysql.com/doc/refman/5.7/en/locking-functions.html
func ReleaseDatabaseLock(db *Database, name string) errors.TracerError {
	var target []*StatusResult
	return errors.Wrap(db.DB.Select(&target, fmt.Sprintf(releaseLockQueryFormat, name)))
}

// Database defines a connection to a database
type Database struct {
	*sqlx.DB
	Logger log.Logger
}

// Initialize establishes the database connection
func Initialize(config Config) *Database {
	logger := log.New("Database", log.FunctionFromEnv())
	log.Infof("initializing database connection: %s, %s", config.DatabaseDialect(), config.DatabaseConnection())
	var err errors.TracerError
	var conn *sqlx.DB
	for retries := 0; retries < config.NumberOfRetries(); retries++ {
		conn, err = connect(config.DatabaseDialect(), config.DatabaseConnection(), logger)
		if nil == err {
			break
		}
		log.Warnf("database connection failed retrying in %s: %s", config.WaitBetweenRetries(), err)
		time.Sleep(config.WaitBetweenRetries())
	}
	if nil != err {
		panic(err)
	}
	log.Infof("database connection success: %s, %s", config.DatabaseDialect(), config.DatabaseConnection())
	return &Database{DB: conn, Logger: logger}
}

func connect(dialect, url string, logger log.Logger) (*sqlx.DB, errors.TracerError) {
	conn, err := sqlx.Connect(dialect, url)

	if nil != err {
		return nil, NewDatabaseConnectionError(err)
	}

	if err = conn.Ping(); nil != err {
		logger.Warnf("Could not ping the database\n%v", err)
		return nil, NewDatabaseConnectionError(err)
	}
	return conn, nil
}

// Model represents the basic table information from a DB Record
type Model struct {
	Name         string
	PrimaryKey   string
	ReadColumns  []string
	WriteColumns []string
}

// Create initializes a Record and inserts it into the Database
func (db *Database) Create(obj Record) errors.TracerError {
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	err = db.CreateTx(obj, tx)
	return CommitOrRollback(tx, err)
}

func appendIfMissing(slice []qb.TableField, i qb.TableField) []qb.TableField {
	if contains(slice, i) {
		return slice
	}
	return append(slice, i)
}

func contains(slice []qb.TableField, i qb.TableField) bool {
	for _, ele := range slice {
		if ele == i {
			return true
		}
	}
	return false
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
	tracerErr := db.ListWhereTx(tx, meta, target, condition, options)
	if nil != tracerErr {
		log.Error(tx.Rollback())
		return tracerErr
	}
	return errors.Wrap(tx.Commit())
}

// ListWhereTx populates target with a list of Records from the database using the transaction
func (db *Database) ListWhereTx(tx *sqlx.Tx, meta Record, target interface{}, condition *qb.ConditionExpression,
	options *ListOptions) errors.TracerError {
	if nil == options {
		options = &ListOptions{
			Limit:  qb.NoLimit,
			Offset: 0,
		}
	}
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
	stmt, values, err := query.SQL(qb.NoLimit, 0)
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
	if nil == options {
		options = &ListOptions{
			Limit:  qb.NoLimit,
			Offset: 0,
		}
	}
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
	return CommitOrRollback(tx, err)
}

// UpdateTx replaces an entry in the database for the Record using a transaction
func (db *Database) UpdateTx(obj Record, tx *sqlx.Tx) errors.TracerError {
	query := qb.Update(obj.Meta())
	for _, col := range obj.Meta().WriteColumns() {
		query.SetParam(col)
	}
	query.Where(obj.Meta().PrimaryKey().Equal(":" + obj.Meta().PrimaryKey().GetName()))
	stmt, err := query.ParameterizedSQL(qb.NoLimit)
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
	return CommitOrRollback(tx, err)
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
	return CommitOrRollback(tx, err)
}

// DeleteWhereTx removes row(s) from the database based on a supplied where clause in a transaction
func (db *Database) DeleteWhereTx(obj Record, tx *sqlx.Tx, condition *qb.ConditionExpression) errors.TracerError {
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

// CommitOrRollback will rollback on an errors.TracerError otherwise commit
func CommitOrRollback(tx *sqlx.Tx, err error) errors.TracerError {
	if err != nil {
		log.Error(tx.Rollback())
		return errors.Wrap(err)
	}
	return errors.Wrap(tx.Commit())
}
