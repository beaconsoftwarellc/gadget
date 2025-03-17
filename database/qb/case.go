package qb

import (
	"fmt"

	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

type whenExpression struct {
	condition *ConditionExpression
	value     any
}

// CaseExpression with one or more when clauses and an optional else clause
type CaseExpression struct {
	whens     []whenExpression
	elseValue any
	alias     string
}

// Case expression for use in or as a SelectExpression
func Case(condition *ConditionExpression, value any) *CaseExpression {
	return &CaseExpression{
		whens: []whenExpression{{condition: condition, value: value}},
	}
}

// Else sets the else value for this case expression
func (exp *CaseExpression) Else(value any) *CaseExpression {
	exp.elseValue = value
	return exp
}

// As sets the alias for this case expression
func (exp *CaseExpression) As(alias string) *CaseExpression {
	exp.alias = alias
	return exp
}

// When adds an additional when clause to this case expression
func (exp *CaseExpression) When(condition *ConditionExpression, value any) *CaseExpression {
	exp.whens = append(exp.whens, whenExpression{condition: condition, value: value})
	return exp
}

// GetTables used by this case expression
func (exp *CaseExpression) GetTables() []string {
	tables := []string{}
	for _, when := range exp.whens {
		tables = append(tables, when.condition.Tables()...)
	}
	return tables
}

// GetName of this case expression
func (exp *CaseExpression) GetName() string {
	return exp.alias
}

// ParameterizedSQL that represents this case expression
func (exp *CaseExpression) ParameterizedSQL() (string, []any) {
	sql := "CASE"
	values := []any{}
	for _, when := range exp.whens {
		conditionSQL, conditionValues := when.condition.SQL()
		values = append(values, conditionValues...)
		sql += fmt.Sprintf(" WHEN %s THEN ?", conditionSQL)
		values = append(values, when.value)
	}
	if nil != exp.elseValue {
		sql += " ELSE ?"
		values = append(values, exp.elseValue)
	}
	sql += " END"
	if !stringutil.IsWhiteSpace(exp.alias) {
		sql += fmt.Sprintf(" AS `%s`", exp.alias)
	}
	return sql, values
}
