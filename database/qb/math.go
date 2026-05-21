package qb

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

type mathOperator string

const (
	add      mathOperator = " + "
	subtract mathOperator = " - "
	multiply mathOperator = " * "
	divide   mathOperator = " / "
)

type math struct {
	expressions []SelectExpression
	operator    mathOperator
	alias       string
}

func (a math) GetName() string {
	return a.alias
}

func (a math) GetTables() []string {
	tables := make([]string, 0)
	for _, expression := range a.expressions {
		tables = append(tables, expression.GetTables()...)
	}
	return tables
}

func (a math) ParameterizedSQL() (string, []any) {
	parts := make([]string, len(a.expressions))
	values := []any{}
	for i, expression := range a.expressions {
		sql, expValues := expression.ParameterizedSQL()
		parts[i] = sql
		values = append(values, expValues...)
	}
	sql := strings.Join(parts, string(a.operator))
	if !stringutil.IsWhiteSpace(a.alias) {
		sql = fmt.Sprintf("%s AS `%s`", sql, a.alias)
	}
	return sql, values
}

// Add the passed expressions together. Pass an empty aliasName to embed the
// result inside another SelectExpression (e.g. Sum), or a non-empty aliasName
// when using as a top-level column.
func Add(expressions []SelectExpression, aliasName string) SelectExpression {
	return math{expressions: expressions, alias: aliasName, operator: add}
}

// Subtract the passed expressions. Pass an empty aliasName to embed the result
// inside another SelectExpression (e.g. Sum), or a non-empty aliasName when
// using as a top-level column.
func Subtract(expressions []SelectExpression, aliasName string) SelectExpression {
	return math{expressions: expressions, alias: aliasName, operator: subtract}
}

// Multiply the passed expressions together. Pass an empty aliasName to embed
// the result inside another SelectExpression (e.g. Sum), or a non-empty
// aliasName when using as a top-level column.
func Multiply(expressions []SelectExpression, aliasName string) SelectExpression {
	return math{expressions: expressions, alias: aliasName, operator: multiply}
}

// Divide the passed expressions. Pass an empty aliasName to embed the result
// inside another SelectExpression (e.g. Sum), or a non-empty aliasName when
// using as a top-level column.
func Divide(expressions []SelectExpression, aliasName string) SelectExpression {
	return math{expressions: expressions, alias: aliasName, operator: divide}
}
