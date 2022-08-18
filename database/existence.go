package database

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

// StatusResult is for capturing output from a function call on the database, you must use
// an 'AS STATUS' clause in your query in order for mapping to work correctly.
type StatusResult struct {
	// Status as returned by a function call usually
	Status int `db:"STATUS"`
}
