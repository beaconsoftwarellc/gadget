//go:generate mockgen -source=$GOFILE -package mocks -destination mocks/configuration.mock.gen.go
package database

import (
	"fmt"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/log"
)

const (
	// DefaultMaxTries for connecting to the database
	DefaultMaxTries = 10
	// DefaultMaxLimit for row counts on select queries
	DefaultMaxLimit = 100
)

// Configuration defines the interface for a specification to establish a database connection
type Configuration interface {
	// Logger for use with databases configured by this instance
	Logger() log.Logger
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
	// SlowQueryThreshold for logging slow queries
	SlowQueryThreshold() time.Duration
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
	// SlowQuery duration establishes the defintion of a slow query for logging
	SlowQuery time.Duration
	// Log for this instance
	Log log.Logger
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

// MaxQueryLimit for row counts on select queries
func (config *InstanceConfig) MaxQueryLimit() uint {
	if config.MaxLimit == 0 {
		config.MaxLimit = DefaultMaxLimit
	}
	return config.MaxLimit
}

// SlowQueryThreshold for logging slow queries
func (config *InstanceConfig) SlowQueryThreshold() time.Duration {
	if config.SlowQuery == 0 {
		config.SlowQuery = defaultSlowQueryThreshold
	}
	return config.SlowQuery
}

// Logger for use with databases configured by this instance
func (config *InstanceConfig) Logger() log.Logger {
	if nil == config.Log {
		config.Log = log.New(
			fmt.Sprintf("%sInstance", config.DatabaseDialect()),
			log.FunctionFromEnv())
	}
	return config.Log
}
