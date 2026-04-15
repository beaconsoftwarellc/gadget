package qb

import (
	"fmt"
	"strings"
)

type mathOperator string

const (
	add      mathOperator = " + "
	subtract mathOperator = " - "
	multiply mathOperator = " * "
	divide   mathOperator = " / "
)

type math struct {
	fields   []TableField
	operator mathOperator
	alias    string
}

func (a math) GetName() string {
	return a.alias
}

func (a math) GetTables() []string {
	tables := make([]string, 0)
	for _, field := range a.fields {
		tables = append(tables, field.GetTables()...)
	}
	return tables
}

func (a math) ParameterizedSQL() (string, []any) {
	fields := make([]string, 0)
	for _, field := range a.fields {
		fields = append(fields, field.SQL())
	}
	return fmt.Sprintf("%s AS `%s`", strings.Join(fields, string(a.operator)), a.alias), nil
}

// Add the 2 fields together
func Add(tableFields []TableField, aliasName string) SelectExpression {
	return math{fields: tableFields, alias: aliasName, operator: add}
}

// Subtract the 2 fields together
func Subtract(tableFields []TableField, aliasName string) SelectExpression {
	return math{fields: tableFields, alias: aliasName, operator: subtract}
}

// Multiply the 2 fields together
func Multiply(tableFields []TableField, aliasName string) SelectExpression {
	return math{fields: tableFields, alias: aliasName, operator: multiply}
}

// Divide the 2 fields together
func Divide(tableFields []TableField, aliasName string) SelectExpression {
	return math{fields: tableFields, alias: aliasName, operator: divide}
}
