package qb

import (
	"fmt"
	"strings"
)

const (
	// SQLNow is the SQL NOW() function for use as a value in expressions.
	SQLNow = "NOW()"

	// SQLNull is the SQL representation of NULL
	SQLNull = "NULL"
)

type expressionUnion struct {
	value interface{}
	field *TableField
	multi []expressionUnion
	cond  *ConditionExpression
}

func newUnion(values ...interface{}) expressionUnion {
	if len(values) == 1 {

		switch v := values[0].(type) {
		case TableField:
			return expressionUnion{field: &v}
		case *ConditionExpression:
			return expressionUnion{cond: v}
		}

		return expressionUnion{value: values[0]}
	}
	multi := make([]expressionUnion, len(values))
	for i, obj := range values {
		multi[i] = newUnion(obj)
	}
	return expressionUnion{multi: multi}
}

func (union expressionUnion) isString() bool {
	_, ok := union.value.(string)
	return ok
}

func (union expressionUnion) isField() bool {
	return nil != union.field
}

func (union expressionUnion) isMulti() bool {
	return nil != union.multi
}

func (union expressionUnion) isConditionExpression() bool {
	return nil != union.cond
}

func (union expressionUnion) getTables() []string {
	if union.isField() {
		return union.field.GetTables()
	} else if union.isMulti() {
		tables := []string{}
		for _, exp := range union.multi {
			tables = append(tables, exp.getTables()...)
		}
		return tables
	} else {
		return []string{}
	}
}

func (union expressionUnion) sql() (string, []interface{}) {
	var sql string
	values := []interface{}{}

	switch {
	case union.isMulti():
		sa := make([]string, len(union.multi))
		var subvalues []interface{}
		for i, exp := range union.multi {
			sa[i], subvalues = exp.sql()
			values = append(values, subvalues...)
		}
		sql = "(" + strings.Join(sa, ", ") + ")"
	case union.isField():
		sql = union.field.SQL()
	case SQLNow == union.value || SQLNull == union.value:
		sql = fmt.Sprintf("%s", union.value)
	case union.isString() && strings.HasPrefix(union.value.(string), ":"):
		sql = fmt.Sprintf("%s", union.value)
	case union.isConditionExpression():
		subsql, subvalues := union.cond.SQL()
		values = append(values, subvalues...)
		sql = subsql
	default:
		sql = "?"
		values = append(values, union.value)
	}
	return sql, values
}

type comparisonExpression interface {
	SQL() (string, []interface{})
}

type parameterExpression struct {
	left       TableField
	comparison Comparison
}

func (be parameterExpression) SQL() (string, []interface{}) {
	left := be.left.SQL()
	return fmt.Sprintf("%s %s :%s", left, be.comparison, be.left.GetName()), make([]interface{}, 0)
}

type binaryExpression struct {
	left       TableField
	comparison Comparison
	right      expressionUnion
}

func (be binaryExpression) SQL() (string, []interface{}) {
	left := be.left.SQL()
	right, values := be.right.sql()
	return fmt.Sprintf("%s %s %s", left, be.comparison, right), values
}

// ConditionExpression represents an expression that can be used as a condition in a where or join on.
type ConditionExpression struct {
	binary   *binaryExpression
	left     *ConditionExpression
	operator string
	right    *ConditionExpression
}

// Tables that are used in this expression or it's sub expressions.
func (exp *ConditionExpression) Tables() []string {
	tables := []string{}
	if nil != exp.binary {
		tables = append(tables, exp.binary.left.GetTables()...)
		tables = append(tables, exp.binary.right.getTables()...)
	} else {
		tables = append(tables, exp.left.Tables()...)
		tables = append(tables, exp.right.Tables()...)
	}
	return tables
}

// FieldComparison to another field or a discrete value.
func FieldComparison(left TableField, comparison Comparison, right interface{}) *ConditionExpression {
	if nil == right {
		right = SQLNull
	}
	return &ConditionExpression{binary: &binaryExpression{left: left, comparison: comparison, right: newUnion(right)}}
}

// FieldIn a series of TableFields and/or values
func FieldIn(left TableField, in ...interface{}) *ConditionExpression {
	// swap any 'nils' for sql null
	rightValues := make([]interface{}, len(in))
	for i, value := range in {
		if nil == value {
			value = SQLNull
		}
		rightValues[i] = value
	}
	comparison := In
	if len(rightValues) == 1 {
		comparison = Equal
	}
	return &ConditionExpression{binary: &binaryExpression{left: left, comparison: comparison, right: newUnion(rightValues...)}}
}

func Bitwise(left TableField, operator BitwiseOperator, right interface{}) *ConditionExpression {
	return &ConditionExpression{
		binary: &binaryExpression{
			left:       left,
			comparison: Comparison(operator),
			right:      newUnion(right),
		},
	}
}

// And creates an expression with this and the passed expression with an AND conjunction.
func (exp *ConditionExpression) And(expression *ConditionExpression) *ConditionExpression {
	ptr := &ConditionExpression{}
	*ptr = *exp
	wrap := &ConditionExpression{left: ptr, right: expression, operator: And}
	*exp = *wrap
	return exp
}

// Or creates an expression with this and the passed expression with an OR conjunction.
func (exp *ConditionExpression) Or(expression *ConditionExpression) *ConditionExpression {
	ptr := &ConditionExpression{}
	*ptr = *exp
	wrap := &ConditionExpression{left: ptr, right: expression, operator: Or}
	*exp = *wrap
	return exp
}

// XOr creates an expression with this and the passed expression with an XOr conjunction.
func (exp *ConditionExpression) XOr(expression *ConditionExpression) *ConditionExpression {
	ptr := &ConditionExpression{}
	*ptr = *exp
	wrap := &ConditionExpression{left: ptr, right: expression, operator: XOr}
	*exp = *wrap
	return exp
}

// SQL returns this condition expression as a SQL expression.
func (exp *ConditionExpression) SQL() (string, []interface{}) {
	if nil != exp.binary {
		return exp.binary.SQL()
	}
	lsql, values := exp.left.SQL()
	rsql, rvalues := exp.right.SQL()
	values = append(values, rvalues...)
	return fmt.Sprintf("(%s %s %s)", lsql, exp.operator, rsql), values
}
