package deltas

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/beaconsoftwarellc/gadget/database"
	"github.com/beaconsoftwarellc/gadget/database/qb"
	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/log"
	"github.com/beaconsoftwarellc/gadget/net"
	"github.com/jmoiron/sqlx"
)

const (
	// DeltaTableName for use in queries
	DeltaTableName = "delta"
	// LockName for preventing multiple executions of the sql delta's at the same time
	LockName             = "delta_exec"
	multiStatementTrueQS = "multiStatements=true"
	// CreateDeltaTableSQL creates the delta's table in the database for use by
	// this package.
	CreateDeltaTableSQL = `CREATE TABLE ` + "`" + DeltaTableName + "`" + ` (
		` + "`id`" + ` int NOT NULL,
		` + "`name`" + ` varchar(120) NOT NULL,
		` + "`created`" + ` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + "`modified`" + ` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (` + "`id`" + `)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
)

var mutex sync.Mutex

func setMultiStatement(connString string) string {
	split := strings.Split(connString, "?")
	if len(split) == 1 {
		connString = fmt.Sprintf("%s?%s", connString, multiStatementTrueQS)
	} else if !strings.Contains(split[1], multiStatementTrueQS) {
		connString = fmt.Sprintf("%s&%s", connString, multiStatementTrueQS)
	}
	return connString
}

// Execute the passed deltas sequentially if they have not already been applied to the database.
// WARN: This function will create the table 'delta' which is needs to track changes if it does not
//		 already exist.
// NOTE: This function assumes that the database has fully transactional DDL
// 		 (not MySQL). Using this function with a non-transactional DDL database
// 		 will cause errors to leave the database in an indeterminate state.
func Execute(config database.InstanceConfig, schema string, deltas []*Delta) errors.TracerError {
	mutex.Lock()
	defer mutex.Unlock()
	config.Connection = setMultiStatement(config.Connection)
	db := database.Initialize(&config)

	return execute(config, schema, deltas, db, getLock, database.AcquireDatabaseLock,
		database.ReleaseDatabaseLock, database.TableExists, database.IsNotFoundError)
}

type dbGetLock func(*fakeDatabase) func() error
type dbAcquireDatabaseLock func(*fakeDatabase, string, time.Duration) (bool, errors.TracerError)
type dbReleaseDatabaseLock func(*fakeDatabase, string) errors.TracerError
type dbTableExists func(*fakeDatabase, string, string) (bool, error)
type dbIsNotFoundError func(err error) bool

type fakeDatabase interface {
	Beginx() (*sqlx.Tx, error)
	Close() error
	ReadOneWhereTx(database.Record, *sqlx.Tx, *qb.ConditionExpression) errors.TracerError
	CreateTx(database.Record, *sqlx.Tx) errors.TracerError
}

func execute(config database.InstanceConfig, schema string, deltas []*Delta, db fakeDatabase,
	getLock dbGetLock, acquireDatabaseLock dbAcquireDatabaseLock, releaseDatabaseLock dbReleaseDatabaseLock, tableExists dbTableExists,
	isNotFoundError dbIsNotFoundError) errors.TracerError {
	var err error

	log.Infof("executing %d deltas on %s", len(deltas), schema)
	err = net.BackoffExtended(
		getLock(db),
		config.NumberOfDeltaLockTries(),
		config.MinimumWaitBetweenDeltaLockRetries(),
		config.MaxWaitBetweenDeltaLockRetries(),
	)
	if nil != err {
		return errors.Wrap(err)
	}

	log.Debugf("db lock acquired")
	defer releaseDatabaseLock(db, LockName)

	// get the lock first
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	exists, err := tableExists(db, schema, DeltaTableName)
	if nil != err {
		return errors.Wrap(err)
	}
	if !exists {
		log.Infof("deltas table does not exist, it will be created")
		_, err = tx.Exec(CreateDeltaTableSQL)
		if nil != err {
			return errors.Wrap(err)
		}
	}
	for i, delta := range deltas {
		if err := ExecuteDelta(tx, db, delta, isNotFoundError); nil != err {
			log.Errorf("rolling back deltas: error encountered executing delta %d %s: %s",
				i, delta.Name, err)
			log.Error(tx.Rollback())
			return err
		}
	}
	log.Infof("all deltas processed")
	terr := errors.Wrap(log.Error(tx.Commit()))
	log.Error(db.Close())
	return terr
}

// ExecuteDelta checks if the passed delta has already been executed according to the Deltas table, and then executes
// if it has not been using the passed transaction for both queries.
func ExecuteDelta(tx *sqlx.Tx, db fakeDatabase, delta *Delta, isNotFoundError dbIsNotFoundError) errors.TracerError {
	log.Infof("processing delta %d %s", delta.ID, delta.Name)
	// check that the delta has not already been executed
	existing := new(DeltaRecord)
	var err error
	err = db.ReadOneWhereTx(existing, tx, DeltaMeta.ID.Equal(delta.ID))
	if nil == err {
		log.Infof("%d %s already executed at %s", delta.ID, delta.Name, existing.Created)
		return nil
	}
	if !isNotFoundError(err) {
		return errors.Wrap(err)
	}

	// actually execute the script
	if _, err = tx.Exec(delta.Script); nil != err {
		return errors.Wrap(err)
	}
	log.Infof("successfully applied delta %d %s", delta.ID, delta.Name)
	return db.CreateTx(&DeltaRecord{ID: delta.ID, Name: delta.Name}, tx)
}

func getLock(db *fakeDatabase) func() error {
	return func() error {
		locked, err := database.AcquireDatabaseLock(db, LockName, 0)
		if nil != err {
			return err
		}
		if !locked {
			return errors.New("unable to acquire lock")
		}
		return nil
	}
}
