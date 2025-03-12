package qb

import (
	"fmt"
)

const (
	sumSQL = "SUM(%s) as sum"
)

// SumResult serves as a target for queries returning only a sum of passed field.
type SumResult struct {
	Sum *int `db:"sum"`
}

// SumExpression can be used as the expression in sum query's. Requires a table
// name that exists in the query to bind to for validation and a field name to sum.
type SumExpression struct {
	table string
	field TableField
}

// NewSumExpression for the passed table and field.
func NewSumExpression(table string, field TableField) SelectExpression {
	return &SumExpression{
		table: table,
		field: field,
	}
}

// GetName of this sum expression
func (se *SumExpression) GetName() string {
	return fmt.Sprintf("SUM(%s)", se.field.GetName())
}

// GetTables used by this sum expression
func (se *SumExpression) GetTables() []string {
	return []string{se.table}
}

// ParameterizedSQL that represents this sum expression
func (se *SumExpression) ParameterizedSQL() (string, []any) {
	return fmt.Sprintf(sumSQL, se.field.GetName()), nil
}
