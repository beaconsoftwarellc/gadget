//go:generate mockgen -source=$GOFILE -package $GOPACKAGE -destination namedstatement.mock.gen.go
package transaction

import sql "database/sql"

// NamedStatement is a prepared statement that executes named queries. Prepare it
// how you would execute a NamedQuery, but pass in a struct or map when executing.
// Not all method represented, see: sqlx@1.3.5/named.go
type NamedStatement interface {
	// Close closes the named statement.
	Close() error
	// Exec executes a named statement using the struct passed.
	// Any named placeholder parameters are replaced with fields from arg.
	Exec(arg interface{}) (sql.Result, error)
	// Get using this NamedStmt
	// Any named placeholder parameters are replaced with fields from arg.
	Get(dest interface{}, arg interface{}) error
}
