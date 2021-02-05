package deltas

import (
	"fmt"
	"strings"
	"sync"

	"github.com/beaconsoftwarellc/gadget/database"
	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/log"
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
	var err error
	config.Connection = setMultiStatement(config.Connection)
	db := database.Initialize(&config)

	log.Infof("executing %d deltas on %s", len(deltas), schema)
	// we want to just bail if someone else has the lock
	locked, err := database.AcquireDatabaseLock(db, LockName, 0)
	if nil != err {
		return errors.Wrap(err)
	}
	if !locked {
		log.Infof("cannot apply deltas: failed to acquire db lock")
		return nil
	}
	log.Debugf("db lock acquired")
	defer database.ReleaseDatabaseLock(db, LockName)

	// get the lock first
	tx, err := db.Beginx()
	if nil != err {
		return errors.Wrap(err)
	}
	exists, err := database.TableExists(db, schema, DeltaTableName)
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
		if err := ExecuteDelta(tx, db, delta); nil != err {
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
func ExecuteDelta(tx *sqlx.Tx, db *database.Database, delta *Delta) errors.TracerError {
	log.Infof("processing delta %d %s", delta.ID, delta.Name)
	// check that the delta has not already been executed
	existing := new(DeltaRecord)
	var err error
	err = db.ReadOneWhereTx(existing, tx, DeltaMeta.ID.Equal(delta.ID))
	if nil == err {
		log.Infof("%d %s already executed at %s", delta.ID, delta.Name, existing.Created)
		return nil
	}
	if !database.IsNotFoundError(err) {
		return errors.Wrap(err)
	}

	// actually execute the script
	if _, err = tx.Exec(delta.Script); nil != err {
		return errors.Wrap(err)
	}
	log.Infof("successfully applied delta %d %s", delta.ID, delta.Name)
	return db.CreateTx(&DeltaRecord{ID: delta.ID, Name: delta.Name}, tx)
}
