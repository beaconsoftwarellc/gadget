package database

import (
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/beaconsoftwarellc/gadget/database/qb"
	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/log"
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
}

func (config *InstanceConfig) DatabaseDialect() string {
	return config.Dialect
}

func (config *InstanceConfig) DatabaseConnection() string {
	return config.Connection
}

func (config *InstanceConfig) NumberOfRetries() int {
	return config.ConnectRetries
}

func (config *InstanceConfig) WaitBetweenRetries() time.Duration {
	if config.ConnectRetryWait == 0 {
		config.ConnectRetryWait = time.Second
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
func (db *Database) ListWhere(def Record, target interface{}, condition *qb.ConditionExpression) errors.TracerError {
	return db.Select(target, db.buildListWhere(def, condition))
}

// ListWhereTx populates obj with a list of Records from the database using the transaction
func (db *Database) ListWhereTx(tx *sqlx.Tx, def Record, obj interface{}, where *qb.ConditionExpression) errors.TracerError {
	query, values, err := db.buildListWhere(def, where).SQL(qb.NoLimit, 0)
	if nil != err {
		return errors.Wrap(err)
	}

	if err = tx.Select(obj, query, values...); nil != err {
		return TranslateError(err, Select, query, db.Logger)
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
	if err = db.DB.Select(target, stmt, values...); nil != err {
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
