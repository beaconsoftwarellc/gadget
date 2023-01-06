package database

import "fmt"

const (
	// TableExistenceQueryFormat returns a single row and column indicating that the table
	// exists and when it was created. Takes format vars 'table_schema' and 'table_name'
	// NOTE: 'as' clause MUST be there or MySQL will return TABLE_NAME for the column name and
	//  mapping will fail
	// NOTE: Do not use 'create_time' for existence check on tables as it can be NULL
	//  for partitioned tables and is always null in aurora
	TableExistenceQueryFormat = `SELECT TABLE_NAME as "table_name" ` +
		`FROM information_schema.tables` +
		`	WHERE table_schema = '%s'` +
		`	AND table_name = '%s' LIMIT 1;`
)

// TableNameResult is for holding the result row from the existence query
type TableNameResult struct {
	// TableName of the table
	TableName string `db:"table_name"`
}

// TableExists for the passed schema and table name on the passed database
func TableExists(db Client, schema, name string) (bool, error) {
	var exists bool
	var err error
	var target []*TableNameResult
	err = db.Select(&target, fmt.Sprintf(TableExistenceQueryFormat, schema, name))
	if len(target) == 1 {
		exists = true
	}
	return exists, err
}
