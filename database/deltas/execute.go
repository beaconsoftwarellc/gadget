package deltas

import (
	"sync"

	"github.com/beaconsoftwarellc/gadget/database"
	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/log"
	"github.com/jmoiron/sqlx"
)

const (
	// DeltaTableName for use in queries
	DeltaTableName = "delta"

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

// Execute the passed deltas sequentially if they have not already been applied to the database.
// WARN: This function will create the table 'delta' which is needs to track changes if it does not
//		 already exist.
// NOTE: This function assumes that the database has fully transactional DDL
// 		 (not MySQL). Using this function with a non-transactional DDL database
// 		 will cause errors to leave the database in an indeterminate state.
func Execute(db *database.Database, schema string, deltas []*Delta) errors.TracerError {
	mutex.Lock()
	defer mutex.Unlock()
	log.Infof("executing %d deltas on %s", len(deltas), schema)
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
	return errors.Wrap(log.Error(tx.Commit()))
}

// ExecuteDelta checks if the passed delta has already been executed according to the Deltas table, and then executes
// if it has not been using the passed transaction for both queries.
func ExecuteDelta(tx *sqlx.Tx, db *database.Database, delta *Delta) errors.TracerError {
	log.Infof("processing delta %s %s", delta.ID, delta.Name)
	// check that the delta has not already been executed
	existing := new(DeltaRecord)
	var err error
	err = db.ReadOneWhereTx(existing, tx, DeltaMeta.ID.Equal(delta.ID))
	if nil == err {
		log.Infof("%s %s already executed at %s", delta.ID, delta.Name, existing.Created)
		return nil
	}
	if !database.IsNotFoundError(err) {
		return errors.Wrap(err)
	}

	// actually execute the script
	if _, err = tx.Exec(delta.Script); nil != err {
		return errors.Wrap(err)
	}
	log.Infof("successfully applied delta %s %s", delta.ID, delta.Name)
	return db.CreateTx(&DeltaRecord{ID: delta.ID, Name: delta.Name}, tx)
}
