package database

import (
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/jmoiron/sqlx"
)

// Initialize establishes the database connection
func Initialize(config Config, logger log.Logger) *Database {
	obfuscatedConnection := obfuscateConnection(config.DatabaseConnection())
	logger.Infof("initializing database connection: %s, %s", config.DatabaseDialect(), obfuscatedConnection)
	var err errors.TracerError
	var conn *sqlx.DB
	for retries := 0; retries < config.NumberOfRetries(); retries++ {
		conn, err = connect(config.DatabaseDialect(), config.DatabaseConnection(), logger)
		if nil == err {
			break
		}
		logger.Warnf("database connection failed retrying in %s: %s", config.WaitBetweenRetries(), err)
		time.Sleep(config.WaitBetweenRetries())
	}
	if nil != err {
		panic(err)
	}
	logger.Infof("database connection success: %s, %s", config.DatabaseDialect(), obfuscatedConnection)

	return &Database{DB: conn, Logger: logger, Configuration: config}
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
	// MaxQueryLimit for row counts on select queries
	MaxQueryLimit() uint
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
	// MaxLimit for row counts on select queries
	MaxLimit uint
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

func (config *InstanceConfig) MaxQueryLimit() uint {
	if config.MaxLimit == 0 {
		config.MaxLimit = DefaultMaxLimit
	}
	return config.MaxLimit
}
