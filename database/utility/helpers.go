package utility

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
	"github.com/samber/lo"
)

const multiStatementTrueQS = "multiStatements=true"

// CountSelect exposes a Count and Select method and is a subset of the db.API
// interface
type CountSelect interface {
	// Count the number of rows in the passed query
	Count(qb.Table, *qb.SelectQuery) (int32, error)
	// Select executes a given select query and populates the target
	Select(target interface{}, query *qb.SelectQuery, options *record.ListOptions) errors.TracerError
}

// SetMultiStatement on the passed connection so that multiple ';' delimited
// statements can be sent to the database at a time.
func SetMultiStatement(connectionString string) string {
	if strings.Contains(connectionString, multiStatementTrueQS) {
		return connectionString
	}
	conjunction := "?"
	split := strings.Split(connectionString, conjunction)
	if len(split) > 1 && !strings.Contains(split[1], multiStatementTrueQS) {
		conjunction = "&"
	}
	return fmt.Sprintf("%s%s%s", connectionString, conjunction,
		multiStatementTrueQS)
}

// CommitRollback exposes the commit and rollback methods that are a
// subset of a the transaction methods
type CommitRollback interface {
	// Commit this transaction
	Commit() errors.TracerError
	// Rollback this transaction
	Rollback() errors.TracerError
}

// CommitOrRollback will rollback on an errors.TracerError otherwise commit
func CommitOrRollback(tx CommitRollback, err error,
	logger log.Logger) errors.TracerError {
	if err != nil {
		logger.Error(tx.Rollback())
		return errors.Wrap(err)
	}
	return errors.Wrap(tx.Commit())
}

// ObfuscateConnection string so that it can be used in log statements.
func ObfuscateConnection(connection string) string {
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

// SelectWithTotal executes the select query populating target and returning the
// total records possible
func SelectWithTotal[T any](db CountSelect, table qb.Table, target T,
	query *qb.SelectQuery, limit, offset int) (T, int, error) {
	var (
		total int32
		err   error
	)
	if total, err = db.Count(table, query); err != nil || limit == 0 || total == 0 {
		return target, int(total), err
	}

	err = db.Select(&target, query, record.NewListOptions(limit, offset))
	if nil != err {
		return target, 0, err
	}
	return target, int(total), err
}

// AppendIfMissing field to the list of fields
func AppendIfMissing(slice []qb.TableField, i qb.TableField) []qb.TableField {
	if lo.Contains(slice, i) {
		return slice
	}
	return append(slice, i)
}
