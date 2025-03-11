package qb

import (
	"fmt"

	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

type whenExpression struct {
	condition *ConditionExpression
	value     interface{}
}

// CaseExpression with one or more when clauses and an optional else clause
type CaseExpression struct {
	whens     []whenExpression
	elseValue interface{}
	alias     string
}

// Case expression for use in or as a SelectExpression
func Case(condition *ConditionExpression, value interface{}) *CaseExpression {
	return &CaseExpression{
		whens: []whenExpression{{condition: condition, value: value}},
	}
}

// Else sets the else value for this case expression
func (exp *CaseExpression) Else(value interface{}) *CaseExpression {
	exp.elseValue = value
	return exp
}

// As sets the alias for this case expression
func (exp *CaseExpression) As(alias string) *CaseExpression {
	exp.alias = alias
	return exp
}

// When adds an additional when clause to this case expression
func (exp *CaseExpression) When(condition *ConditionExpression, value interface{}) *CaseExpression {
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

// SQL that represents this case expression
func (exp *CaseExpression) SQL() string {
	return InsertSQLParameters(exp.ParameterizedSQL())
}

// ParameterizedSQL that represents this case expression
func (exp *CaseExpression) ParameterizedSQL() (string, []interface{}) {
	sql := "CASE"
	values := []interface{}{}
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
