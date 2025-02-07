package qb

import (
	"fmt"
)

const (
	sumSQL = "SUM(%s) as sum"
)

// SumResult serves as a target for queries returning only a sum of passed field.
type SumResult struct {
	Sum int `db:"sum"`
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
func (ce *SumExpression) GetName() string {
	return fmt.Sprintf("SUM(%s)", ce.field.GetName())
}

// GetTables used by this sum expression
func (ce *SumExpression) GetTables() []string {
	return []string{ce.table}
}

// SQL fragment this sum expressions represents
func (ce *SumExpression) SQL() string {
	return fmt.Sprintf(sumSQL, ce.field.GetName())
}
