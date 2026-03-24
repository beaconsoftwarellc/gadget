package qb

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/collection"
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

// CountDistinctExpression can be used as the expression in the count distinct query. Requires
// a [TableField] that exists in the query to bind to for validation.
type CountDistinctExpression struct {
	selectExpressions []SelectExpression
}

// NewCountDistinctExpression of the passed [TableField] for use in a SelectExpression
func NewCountDistinctExpression(tableField TableField) SelectExpression {
	return &CountDistinctExpression{
		selectExpressions: []SelectExpression{SelectExpression(tableField)},
	}
}

// NewCountDistinct of the passed expressions for use as a [SelectExpression] in a COUNT DISTINCT query.
func NewCountDistinct(selectExpressions []SelectExpression) SelectExpression {
	return &CountDistinctExpression{
		selectExpressions: selectExpressions,
	}
}

// GetName of this count distinct expression
func (cde CountDistinctExpression) GetName() string {
	names := make([]string, len(cde.selectExpressions))
	for i, expression := range cde.selectExpressions {
		names[i] = expression.GetName()
	}
	return fmt.Sprintf("COUNT(DISTINCT %s)", strings.Join(names, ", "))
}

// GetTables used by this count distinct expression
func (cde CountDistinctExpression) GetTables() []string {
	tableNames := collection.NewSet[string]()
	for _, expression := range cde.selectExpressions {
		tableNames.Add(expression.GetTables()...)
	}
	return tableNames.Elements()
}

// ParameterizedSQL that represents this count distinct expression
func (cde CountDistinctExpression) ParameterizedSQL() (string, []any) {
	sql := make([]string, len(cde.selectExpressions))
	for i, expression := range cde.selectExpressions {
		sql[i], _ = expression.ParameterizedSQL()
	}
	return fmt.Sprintf(countDistinctSQL, strings.Join(sql, ", ")), nil
}
