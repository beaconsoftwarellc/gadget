package qb

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// ValidationFromNotSetError  is set on the query when From has not been called on this query.
type ValidationFromNotSetError struct{ trace []string }

func (err *ValidationFromNotSetError) Error() string {
	return "validation: from table must be set"
}

// Trace returns the stack trace for the error
func (err *ValidationFromNotSetError) Trace() []string {
	return err.trace
}

// NewValidationFromNotSetError instantiates a ValidationFromNotSetError with a stack trace
func NewValidationFromNotSetError() errors.TracerError {
	return &ValidationFromNotSetError{trace: errors.GetStackTrace()}
}

// MissingTablesError is returned when column's are being used from a table that is not part of the query.
type MissingTablesError struct {
	Tables []string
	trace  []string
}

func (err *MissingTablesError) Error() string {
	return fmt.Sprintf("validation: the following tables are required but were not included in a join or from: %s",
		err.Tables)
}

// Trace returns the stack trace for the error
func (err *MissingTablesError) Trace() []string {
	return err.trace
}

// NewMissingTablesError is returned when column's are being used from a table that is not part of the query.
func NewMissingTablesError(tables []string) errors.TracerError {
	return &MissingTablesError{
		Tables: tables,
		trace:  errors.GetStackTrace(),
	}
}

// Comparison of two fields
type Comparison string

// JoinType inner or outer
type JoinType string

// JoinDirection left or right
type JoinDirection string

// OrderDirection for use in an order by Ascending or Descending
type OrderDirection string

const (
	// Equal Comparison Operator
	Equal Comparison = "="
	// NotEqual Comparison Operator
	NotEqual Comparison = "!="
	// LessThan Comparison Operator
	LessThan Comparison = "<"
	// LessThanEqual Comparison Operator
	LessThanEqual Comparison = "<="
	// GreaterThan Comparison Operator
	GreaterThan Comparison = ">"
	// GreaterThanEqual Comparison Operator
	GreaterThanEqual Comparison = ">="
	// NullSafeEqual Comparison Operator
	NullSafeEqual Comparison = "<=>"
	// Is Comparison Operator
	Is Comparison = "IS"
	// IsNot Comparison Operator
	IsNot Comparison = "IS NOT"
	// In Comparison Operator
	In Comparison = "IN"
	// LIKE Comparison Operator
	Like Comparison = "LIKE"
	// Inner JoinType
	Inner JoinType = "INNER"
	// Outer JoinType
	Outer JoinType = "OUTER"
	// Cross JoinType
	Cross JoinType = "CROSS"
	// Left JoinDirection
	Left JoinDirection = "LEFT"
	// Right JoinDirection
	Right JoinDirection = "RIGHT"
	// Ascending OrderDirection
	Ascending OrderDirection = "ASC"
	// Descending OrderDirection
	Descending OrderDirection = "DESC"
	// And expression conjunction
	And = "AND"
	// Or expression conjunction
	Or = "OR"
	// XOr expression conjunction
	XOr = "XOR"
	// NoLimit is the value that represents not applying a limit on the query
	// we are not using ^uint(0) because that is easy to guess and would make
	// this value not comparable to int
	NoLimit = 1<<16 - 1337 // 4@><0R5 yea yea
)

type BitwiseOperator string

const (
	// Binary and expression conjunction
	BitwiseAnd BitwiseOperator = "&"
	// Bitwise or expression conjunction
	BitwiseOr BitwiseOperator = "|"
	// Bitwise xor expression conjunction
	BitwiseXor BitwiseOperator = "^"
	// Bitwise negation expression conjunction
	BitwiseNegation BitwiseOperator = "~"
	// Bitwise and negation expression conjunction
	BitwiseAndNegation BitwiseOperator = "&~"
	// Bitwise or negation expression conjunction
	BitwiseOrNegation BitwiseOperator = "|~"
	// Bitwise xor negation expression conjunction
	BitwiseXorNegation BitwiseOperator = "^~"
)

// Table represents a db table
type Table interface {
	// GetName returns the name of the database table
	GetName() string
	// GetAlias returns the alias of the database table to be used in the query
	GetAlias() string
	// PrimaryKey returns the primary key TableField
	PrimaryKey() TableField
	// AllColumns returns the AllColumns TableField for this Table
	AllColumns() TableField
	// ReadColumns returns the default set of columns for a read operation
	ReadColumns() []TableField
	// WriteColumns returns the default set of columns for a write operation
	WriteColumns() []TableField
	// SortBy returns the name of the default sort by field
	SortBy() (TableField, OrderDirection)
}

// TableField represents a single column on a table.
type TableField struct {
	// Name of the column in the database table
	Name string
	// Table that the column is on
	Table string
}

// FieldValue represents a single field value on a table.
type FieldValue struct {
	Field TableField
	Value any
}

// GetName that can be used to reference this expression
func (tf TableField) GetName() string {
	return tf.Name
}

// GetTables that are used in this expression
func (tf TableField) GetTables() []string {
	return []string{tf.Table}
}

// SQL that represents this table field
func (tf TableField) SQL() string {
	if tf.Name == "*" {
		return fmt.Sprintf("`%s`.%s", tf.Table, tf.Name)
	}
	return fmt.Sprintf("`%s`.`%s`", tf.Table, tf.Name)
}

// ParameterizedSQL that represents this table field
func (tf TableField) ParameterizedSQL() (string, []any) {
	return tf.SQL(), nil
}

// Equal returns a condition expression for this table field Equal to the passed obj.
func (tf TableField) Equal(obj any) *ConditionExpression {
	return FieldComparison(tf, Equal, obj)
}

// NotEqual returns a condition expression for this table field NotEqual to the passed obj.
func (tf TableField) NotEqual(obj any) *ConditionExpression {
	return FieldComparison(tf, NotEqual, obj)
}

// LessThan returns a condition expression for this table field LessThan to the passed obj.
func (tf TableField) LessThan(obj any) *ConditionExpression {
	return FieldComparison(tf, LessThan, obj)
}

// LessThanEqual returns a condition expression for this table field LessThanEqual to the passed obj.
func (tf TableField) LessThanEqual(obj any) *ConditionExpression {
	return FieldComparison(tf, LessThanEqual, obj)
}

// GreaterThan returns a condition expression for this table field GreaterThan to the passed obj.
func (tf TableField) GreaterThan(obj any) *ConditionExpression {
	return FieldComparison(tf, GreaterThan, obj)
}

// GreaterThanEqual returns a condition expression for this table field GreaterThanEqual to the passed obj.
func (tf TableField) GreaterThanEqual(obj any) *ConditionExpression {
	return FieldComparison(tf, GreaterThanEqual, obj)
}

// NullSafeEqual returns a condition expression for this table field NullSafeEqual to the passed obj.
func (tf TableField) NullSafeEqual(obj any) *ConditionExpression {
	return FieldComparison(tf, NullSafeEqual, obj)
}

// In returns a condition expression for this table field in to the passed objs.
func (tf TableField) In(objs ...any) *ConditionExpression {
	return FieldIn(tf, objs...)
}

// Like returns a condition expression for this table field Like to the passed obj.
func (tf TableField) Like(obj any) *ConditionExpression {
	return FieldComparison(tf, Like, obj)
}

// IsNull returns a condition expression for this table field when it is NULL
func (tf TableField) IsNull() *ConditionExpression {
	return FieldComparison(tf, Is, SQLNull)
}

// IsNotNull returns a condition expression for this table field where it is not NULL
func (tf TableField) IsNotNull() *ConditionExpression {
	return FieldComparison(tf, IsNot, SQLNull)
}

type orderByExpression struct {
	field     TableField
	direction OrderDirection
}

type orderBy struct {
	expressions []orderByExpression
}

func (ob *orderBy) addExpression(field TableField, direction OrderDirection) *orderBy {
	exp := orderByExpression{field: field, direction: direction}
	if nil == ob.expressions {
		ob.expressions = []orderByExpression{exp}
	} else {
		ob.expressions = append(ob.expressions, exp)
	}
	return ob
}

func (ob *orderBy) getTables() []string {
	tables := make([]string, len(ob.expressions))
	for i, exp := range ob.expressions {
		tables[i] = exp.field.Table
	}
	return tables
}

func (ob *orderBy) sql() (string, bool) {
	if nil == ob.expressions || len(ob.expressions) == 0 {
		return "", false
	}
	orderByLines := []string{}
	for _, orderBy := range ob.expressions {
		orderByLines = append(orderByLines, fmt.Sprintf("`%s`.`%s` %s", orderBy.field.Table, orderBy.field.Name, orderBy.direction))
	}
	return "ORDER BY " + strings.Join(orderByLines, ", "), true
}

type whereCondition struct {
	expression *ConditionExpression
}

func (wc *whereCondition) tables() []string {
	tables := []string{}
	if nil != wc.expression {
		tables = append(tables, wc.expression.Tables()...)
	}
	return tables
}

func (wc *whereCondition) sql() (string, []any, bool) {
	var sql string
	var values []any
	ok := false
	if nil != wc.expression {
		sql, values = wc.expression.SQL()
		ok = true
	}
	return sql, values, ok
}

// Join on the tables inside the query.
type Join struct {
	direction JoinDirection
	joinType  JoinType
	table     Table
	condition *ConditionExpression
	err       error
}

// JoinError signifying a problem with the created join.
type JoinError struct {
	conditionTables []string
	joinTable       string
}

// NewJoin of the specified type and direction.
func NewJoin(joinType JoinType, joinDirection JoinDirection, table Table) *Join {
	return &Join{joinType: joinType, direction: joinDirection, table: table,
		err: errors.New("no condition specified for join")}
}

func (err *JoinError) Error() string {
	return fmt.Sprintf("join field to field condition (tables: %s) does not include table being joined '%s'",
		err.conditionTables, err.joinTable)
}

// On specifies the the conditions of a join based upon two fields or a field and a discrete value
func (join *Join) On(left TableField, comparison Comparison, right any) *ConditionExpression {
	join.err = nil
	rt, ok := right.(TableField)
	if ok && left.Table != join.table.GetName() && rt.Table != join.table.GetName() {
		join.err = &JoinError{conditionTables: []string{left.Table, rt.Table}, joinTable: join.table.GetName()}
	} else if !ok && left.Table != join.table.GetName() {
		join.err = &JoinError{conditionTables: []string{left.Table}, joinTable: join.table.GetName()}
	}
	join.condition = FieldComparison(left, comparison, right)
	return join.condition
}

// SQL that represents this join.
func (join *Join) SQL() (string, []any) {
	if nil != join.err {
		return "", []any{}
	}
	var lines []string
	if join.joinType == Inner || join.joinType == Cross {
		lines = []string{fmt.Sprintf("%s JOIN `%s` AS `%s` ON", join.joinType,
			join.table.GetName(), join.table.GetAlias())}
	} else {
		lines = []string{fmt.Sprintf("%s %s JOIN `%s` AS `%s` ON", join.direction, join.joinType, join.table.GetName(),
			join.table.GetAlias())}
	}
	expressionSQL, values := join.condition.SQL()

	lines = append(lines, expressionSQL)
	return strings.Join(lines, " "), values
}

// Select creates a new select query based on the passed expressions for the select clause.
func Select(selectExpressions ...SelectExpression) *SelectQuery {
	query := &SelectQuery{
		selectExps: selectExpressions,
		orderBy:    &orderBy{},
		groupBy:    []TableField{},
		where:      &whereCondition{},
		Seperator:  " ",
	}
	return query
}

// SelectDistinct creates a new select query based on the passed expressions for the select clause with a distinct
// modifier.
func SelectDistinct(selectExpressions ...SelectExpression) *SelectQuery {
	query := Select(selectExpressions...)
	query.distinct = true
	return query
}

// Insert columns into a table
func Insert(columns ...TableField) *InsertQuery {
	return &InsertQuery{
		columns:           columns,
		values:            [][]any{},
		onDuplicate:       []TableField{},
		onDuplicateValues: []any{},
	}
}

// Update returns a query that can be used for updating rows in the passed table.
func Update(table Table) *UpdateQuery {
	return &UpdateQuery{
		tableReference: table,
		assignments:    []comparisonExpression{},
		orderBy:        &orderBy{},
		where:          &whereCondition{},
	}
}

// Delete from from the specified tables that match the criteria specified in where.
func Delete(rowsIn ...Table) *DeleteQuery {
	return &DeleteQuery{
		tables: rowsIn,
		joins:  []*Join{},
		where:  &whereCondition{},
	}
}
