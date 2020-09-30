package qb

import (
	"fmt"
	"strings"
)

// SelectExpression for use in identifying the fields desired in a select query.
type SelectExpression interface {
	// GetName that can be used to reference this expression
	GetName() string
	// GetTables that are used in this expression
	GetTables() []string
	// SQL that represents this SelectExpression
	SQL() string
}

type alias struct {
	field TableField
	alias string
}

func (a alias) GetName() string {
	return a.alias
}

func (a alias) GetTables() []string {
	return a.field.GetTables()
}

func (a alias) SQL() string {
	return fmt.Sprintf("%s AS `%s`", a.field.SQL(), a.alias)
}

// Alias the passed table field for use in or as a SelectExpression
func Alias(tableField TableField, aliasName string) SelectExpression {
	return alias{field: tableField, alias: aliasName}
}

type notNull struct {
	field TableField
	alias string
}

func (nn notNull) GetName() string {
	return nn.alias
}

func (nn notNull) GetTables() []string {
	return nn.field.GetTables()
}

func (nn notNull) SQL() string {
	return fmt.Sprintf("(%s IS NOT NULL) AS `%s`", nn.field.SQL(), nn.alias)
}

// NotNull field for use as a select expression
func NotNull(tableField TableField, alias string) SelectExpression {
	return notNull{field: tableField, alias: alias}
}

type coalesce struct {
	field TableField
	def   string
	name  string
}

func (c coalesce) GetName() string {
	return c.name
}

func (c coalesce) GetTables() []string {
	return c.field.GetTables()
}

func (c coalesce) SQL() string {
	return fmt.Sprintf("COALESCE(%s, '%s') AS `%s`", c.field.SQL(), c.def, c.name)
}

// Coalesce creates a SQL coalesce that can be used as a SelectExpression
func Coalesce(column TableField, defaultValue string, alias string) SelectExpression {
	return coalesce{field: column, def: defaultValue, name: alias}
}

// SelectQuery for retrieving data from a database table.
type SelectQuery struct {
	distinct   bool
	from       Table
	selectExps []SelectExpression
	joins      []*Join
	orderBy    *orderBy
	groupBy    []SelectExpression
	where      *whereCondition
	Seperator  string
	err        error
}

// GetAlias of the passed table name in this query.
func (q *SelectQuery) GetAlias(tableName string) string {
	return tableName
}

// From sets the primary table the query will get values from.
func (q *SelectQuery) From(table Table) *SelectQuery {
	q.from = table
	return q
}

// InnerJoin with another table in the database.
func (q *SelectQuery) InnerJoin(table Table) *Join {
	join := NewJoin(Inner, Right, table)
	q.joins = append(q.joins, join)
	return join
}

// OuterJoin with another table in the database.
func (q *SelectQuery) OuterJoin(direction JoinDirection, table Table) *Join {
	join := NewJoin(Outer, direction, table)
	q.joins = append(q.joins, join)
	return join
}

// Where the comparison between the two tablefields evaluates to true.
func (q *SelectQuery) Where(condition *ConditionExpression) *SelectQuery {
	q.where.expression = condition
	return q
}

// OrderBy the passed field and direction.
func (q *SelectQuery) OrderBy(field TableField, direction OrderDirection) *SelectQuery {
	q.orderBy.addExpression(field, direction)
	return q
}

// GroupBy the passed table field.
func (q *SelectQuery) GroupBy(expressions ...SelectExpression) *SelectQuery {
	q.groupBy = append(q.groupBy, expressions...)
	return q
}

func (q *SelectQuery) selectExpressionsSQL() string {
	var prefix string
	if q.distinct {
		prefix = "SELECT DISTINCT"
	} else {
		prefix = "SELECT"
	}
	expressions := make([]string, len(q.selectExps))
	for i, exp := range q.selectExps {
		expressions[i] = exp.SQL()
	}
	return fmt.Sprintf("%s %s", prefix, strings.Join(expressions, ", "))
}

// Validate that this query can be executed.
func (q *SelectQuery) Validate() bool {
	q.err = nil
	// gather up all the tables that must be present in the from or in a join
	tablesRequired := make(map[string]bool)
	// check the select
	for _, exp := range q.selectExps {
		for _, table := range exp.GetTables() {
			tablesRequired[table] = true
		}
	}
	// check the where expressions
	for _, table := range q.where.tables() {
		tablesRequired[table] = true
	}
	// now get the tables from the order by
	for _, table := range q.orderBy.getTables() {
		tablesRequired[table] = true
	}
	// grab the tables from the group by
	for _, tf := range q.groupBy {
		for _, table := range tf.GetTables() {
			tablesRequired[table] = true
		}
	}
	// check that the from table is set
	if nil == q.from {
		q.err = NewValidationFromNotSetError()
		return false
	}
	delete(tablesRequired, q.from.GetAlias())

	for _, join := range q.joins {
		delete(tablesRequired, join.table.GetAlias())
		if nil != join.err {
			q.err = join.err
			return false
		}
	}

	if len(tablesRequired) > 0 {
		missingTables := make([]string, len(tablesRequired))
		i := 0
		for key := range tablesRequired {
			missingTables[i] = key
			i++
		}
		q.err = NewMissingTablesError(missingTables)
		return false
	}

	return true
}

// SQL statement corresponding to this query.
func (q *SelectQuery) SQL(limit, offset uint) (string, []interface{}, error) {
	if !q.Validate() {
		return "", []interface{}{}, q.err
	}
	// SELECT
	lines := []string{q.selectExpressionsSQL()}
	var values []interface{}

	// FROM
	from := fmt.Sprintf("FROM `%s` AS `%s`",
		q.from.GetName(),
		q.from.GetAlias())
	lines = append(lines, from)

	// JOIN
	for _, join := range q.joins {
		joinSQL, joinValues := join.SQL()
		lines = append(lines, joinSQL)
		values = append(values, joinValues...)
	}

	// WHERE
	if where, whereValues, ok := q.where.sql(); ok {
		lines = append(lines, "WHERE", where)
		values = append(values, whereValues...)
	}

	// GROUP BY
	if len(q.groupBy) > 0 {
		groupByLines := []string{}
		for _, tf := range q.groupBy {
			groupByLines = append(groupByLines, tf.SQL())
		}
		groupByStatement := "GROUP BY " + strings.Join(groupByLines, ", ")
		lines = append(lines, groupByStatement)
	}

	// ORDER BY
	if orderby, ok := q.orderBy.sql(); ok {
		lines = append(lines, orderby)
	}

	// LIMIT, OFFSET
	if NoLimit != limit {
		lines = append(lines, fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset))
	}
	return strings.Join(lines, q.Seperator), values, q.err
}
