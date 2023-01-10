//go:generate mockgen -source=$GOFILE -package $GOPACKAGE -destination implementation.mock.gen.go
package transaction

import (
	sql "database/sql"

	sqlx "github.com/jmoiron/sqlx"
)

type Implementation interface {
	// NamedQuery within a transaction.
	// Any named placeholder parameters are replaced with fields from arg.
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	// NamedExec a named query within a transaction.
	// Any named placeholder parameters are replaced with fields from arg.
	NamedExec(query string, arg interface{}) (sql.Result, error)
	// QueryRowx within a transaction.
	// Any placeholder parameters are replaced with supplied args.
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	// Select within a transaction.
	// Any placeholder parameters are replaced with supplied args.
	Select(dest interface{}, query string, args ...interface{}) error
	// Exec a query within a transaction.
	// Any named placeholder parameters are replaced with fields from arg.
	Exec(query string, args ...any) (sql.Result, error)
	// Commit this transaction
	Commit() error
	// Rollback this transaction
	Rollback() error
}
