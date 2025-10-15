package qb

import (
	"fmt"
)

const (
	countSQL         = "COUNT(*) as count"
	countDistinctSQL = "COUNT(DISTINCT %s) as count"
)

// RowCount serves as a target for queries returning only a count of rows
type RowCount struct {
	Count int `db:"count"`
}

// CountExpression can be used as the expression in count query's. Requires a table
// name that exists in the query to bind to for validation.
type CountExpression struct {
	table string
}

// NewCountExpression for the passed table. Table must be a part of the query.
func NewCountExpression(table string) SelectExpression {
	return &CountExpression{
		table: table,
	}
}

// GetName of this count expression
func (ce *CountExpression) GetName() string {
	return fmt.Sprintf("COUNT(%s)", ce.table)
}

// GetTables used by this count expression
func (ce *CountExpression) GetTables() []string {
	return []string{ce.table}
}

// ParameterizedSQL that represents this count expression
func (ce *CountExpression) ParameterizedSQL() (string, []any) {
	return countSQL, nil
}

// CountDistinctExpression can be used as the expression in count distinct query's. Requires
// a [TableField] that exists in the query to bind to for validation.
type CountDistinctExpression struct {
	tableField TableField
}

// NewCountDistinctExpression of the passed [TableField] for use in a SelectExpression
func NewCountDistinctExpression(tableField TableField) SelectExpression {
	return &CountDistinctExpression{
		tableField: tableField,
	}
}

// GetName of this count distinct expression
func (cde CountDistinctExpression) GetName() string {
	return fmt.Sprintf("COUNT(DISTINCT %s)", cde.tableField.GetName())
}

// GetTables used by this count distinct expression
func (cde CountDistinctExpression) GetTables() []string {
	return cde.tableField.GetTables()
}

// ParameterizedSQL that represents this count distinct expression
func (cde CountDistinctExpression) ParameterizedSQL() (string, []any) {
	return fmt.Sprintf(countDistinctSQL, cde.tableField.SQL()), nil
}
