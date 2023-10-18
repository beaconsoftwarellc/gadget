//go:generate mockgen -source=$GOFILE -package $GOPACKAGE -destination connection.mock.gen.go
package database

import (
	"time"

	dberrors "github.com/beaconsoftwarellc/gadget/v2/database/errors"
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/database/utility"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/jmoiron/sqlx"
)

// Client interface for working with the database driver functions
type Client interface {
	// necessary for executing non-computed queries
	Select(dest interface{}, query string, args ...interface{}) error
	// Beginx starts a sqlx.Tx and returns it
	Beginx() (*sqlx.Tx, error)
	// Close this client
	Close() error
}

type transactable struct {
	db Client
}

func (t *transactable) Begin() (transaction.Implementation, error) {
	return t.db.Beginx()
}

// Connection represents a connection to a database
type Connection interface {
	// GetConfiguration used to create this instance
	GetConfiguration() Configuration
	// Client for working with this connection at the driver level
	Client() Client
	// Database API that this connection is connected to
	Database() API
	// Close this collection, further calls will panic
	Close() error
}

func connect(dialect, url string, logger log.Logger) (*sqlx.DB, errors.TracerError) {
	conn, err := sqlx.Connect(dialect, url)

	if nil != err {
		return nil, dberrors.NewDatabaseConnectionError(err)
	}

	if err = conn.Ping(); nil != err {
		logger.Warnf("Could not ping the database\n%v", err)
		return nil, dberrors.NewDatabaseConnectionError(err)
	}
	return conn, nil
}

func Connect(cfg Configuration) (Connection, error) {
	var (
		obfuscatedConnection = utility.ObfuscateConnection(cfg.DatabaseConnection())
		err                  errors.TracerError
		conn                 *sqlx.DB
	)
	log.Infof("initializing database connection: %s, %s", cfg.DatabaseDialect(),
		obfuscatedConnection)

	for retries := 0; retries < cfg.NumberOfRetries(); retries++ {
		conn, err = connect(cfg.DatabaseDialect(), cfg.DatabaseConnection(), cfg.Logger())
		if nil == err {
			break
		}
		log.Warnf("database connection failed retrying in %s: %s", cfg.WaitBetweenRetries(), err)
		time.Sleep(cfg.WaitBetweenRetries())
	}
	if nil != err {
		return nil, err
	}
	log.Infof("database connection success: %s, %s", cfg.DatabaseDialect(), obfuscatedConnection)
	return &connection{client: conn, configuration: cfg, connected: true}, nil
}

type connection struct {
	client        Client
	configuration Configuration
	connected     bool
}

func (c *connection) GetConfiguration() Configuration {
	return c.configuration
}

func (c *connection) Client() Client {
	if !c.connected {
		panic("Client() called on disconnected connection")
	}
	return c.client
}

func (c *connection) Database() API {
	if !c.connected {
		panic("Database() called on disconnected connection")
	}
	return &api{db: &transactable{c.client}, configuration: c.configuration}
}

// NewBulkCreate API creating multiple records at the same time.
func NewBulkCreate[T record.Record](cfg Configuration) (BulkCreate[T], error) {
	// get a new connection with multistatement enabled
	var (
		connectionString     = utility.SetMultiStatement(cfg.DatabaseConnection())
		obfuscatedConnection = utility.ObfuscateConnection(connectionString)
		err                  error
		connection           *sqlx.DB
	)
	connection, err = connect(cfg.DatabaseDialect(), cfg.DatabaseConnection(), cfg.Logger())
	if nil != err {
		cfg.Logger().Errorf("failed to connect to '%s': %s", obfuscatedConnection, err)
		return nil, err
	}
	bc := &bulkCreate[T]{
		bulkOperation: &bulkOperation[T]{
			db: &transactable{connection}, configuration: cfg,
		},
	}
	bc.tx, err = transaction.New(bc.db, cfg.Logger(),
		cfg.SlowQueryThreshold(), cfg.LoggedSlowQueries())
	return bc, err
}

// NewBulkUpdate of the columns on type T. Only the specified columns
// will be updated on commit.
func NewBulkUpdate[T record.Record](
	cfg Configuration,
	columns ...qb.TableField,
) (BulkUpdate[T], error) {
	if len(columns) == 0 {
		return nil, errors.New("at least one column is required")
	}
	// get a new connection with multistatement enabled
	var (
		connectionString     = utility.SetMultiStatement(cfg.DatabaseConnection())
		obfuscatedConnection = utility.ObfuscateConnection(connectionString)
		err                  error
		connection           *sqlx.DB
	)
	connection, err = connect(cfg.DatabaseDialect(),
		cfg.DatabaseConnection(), cfg.Logger())
	if nil != err {
		cfg.Logger().Errorf("failed to connect to '%s': %s", obfuscatedConnection, err)
		return nil, err
	}
	bu := &bulkUpdate[T]{
		bulkOperation: &bulkOperation[T]{
			db: &transactable{connection}, configuration: cfg,
		},
		columns: columns,
	}
	bu.tx, err = transaction.New(bu.db, cfg.Logger(),
		cfg.SlowQueryThreshold(), cfg.LoggedSlowQueries())
	return bu, err
}

func (c *connection) Close() error {
	if !c.connected {
		return errors.New("Close() called on disconnected connection")
	}
	c.connected = false
	return c.client.Close()
}
