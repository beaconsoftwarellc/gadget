package database

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
	"github.com/jmoiron/sqlx"
)

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

// CommitOrRollback will rollback on an errors.TracerError otherwise commit
func CommitOrRollback(tx *sqlx.Tx, err error, logger log.Logger) errors.TracerError {
	if err != nil {
		logger.Error(tx.Rollback())
		return errors.Wrap(err)
	}
	return errors.Wrap(tx.Commit())
}

func obfuscateConnection(connection string) string {
	obfuscateIndex := strings.LastIndex(connection, "@")
	obfuscatedConnection := connection
	if obfuscateIndex > 0 {
		// no '@' means the credentials are not part of the connection and we do not
		// need to obfuscate
		obfuscatedConnection = stringutil.ObfuscateLeft(obfuscatedConnection,
			obfuscateIndex, "*")
	}
	return obfuscatedConnection
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

// Count the number of rows in the passed query
func Count(db API, table qb.Table, query *qb.SelectQuery) (int32, error) {
	var target []*qb.RowCount

	err := db.Select(&target, query.SelectFrom(qb.NewCountExpression(table.GetName())))
	if err != nil {
		return 0, err
	}
	if len(target) == 0 {
		return 0, errors.New("[COM.DB.1] row count query execution failed (no rows)")
	}
	return int32(target[0].Count), nil
}

// CountWhere the number of rows in the passed query
func CountWhere(db API, table qb.Table, condition *qb.ConditionExpression) (int32, error) {
	query := qb.Select(qb.NewCountExpression(table.GetName())).
		From(table).
		Where(condition)

	return Count(db, table, query)
}

// CountTableRows from the passed database and table name
func CountTableRows(db API, table qb.Table) (int32, error) {
	return CountWhere(db, table, nil)
}
