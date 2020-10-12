package qb

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/errors"
)

/*
	UPDATE [LOW_PRIORITY] [IGNORE] table_reference
		SET assignment_list
		[WHERE where_condition]
		[ORDER BY ...]
		[LIMIT row_count]

	value:
		{expr | DEFAULT}

	assignment:
		col_name = value

	assignment_list:
		assignment [, assignment] ...
*/

// UpdateQuery represents a query to update rows in a database
// Currently only supports single table, to change this the tableReference would have to be built out more.
type UpdateQuery struct {
	tableReference Table
	assignments    []comparisonExpression
	where          *whereCondition
	orderBy        *orderBy
	err            error
}

// GetAlias returns the alias for the passed tablename used in this query.
func (q *UpdateQuery) GetAlias(tableName string) string {
	// no aliasing in update
	return tableName
}

// Set adds a assignment to this update query.
func (q *UpdateQuery) Set(field TableField, value interface{}) *UpdateQuery {
	if field.Table != q.tableReference.GetName() {
		q.err = errors.New("field table does not match table reference on update query")
	} else {
		q.assignments = append(q.assignments, binaryExpression{left: field, comparison: Equal, right: newUnion(value)})
	}
	return q
}

// SetParam adds a parameterized assignment to this update query.
func (q *UpdateQuery) SetParam(field TableField) *UpdateQuery {
	if field.Table != q.tableReference.GetName() {
		q.err = errors.New("field table does not match table reference on update query")
	} else {
		q.assignments = append(q.assignments, parameterExpression{left: field, comparison: Equal})
	}
	return q
}

// Where determines the conditions by which the assignments in this query apply
func (q *UpdateQuery) Where(condition *ConditionExpression) *UpdateQuery {
	q.where.expression = condition
	return q
}

// OrderBy the passed field and direction.
func (q *UpdateQuery) OrderBy(field TableField, direction OrderDirection) *UpdateQuery {
	q.orderBy.addExpression(field, direction)
	return q
}

// SQL representation of this query.
func (q *UpdateQuery) SQL(limit int) (string, []interface{}, error) {
	if nil != q.err {
		return "", nil, q.err
	}
	if len(q.assignments) == 0 {
		return "", nil, errors.New("no assignments in update query")
	}
	sql := []string{fmt.Sprintf("UPDATE `%s` SET ", q.tableReference.GetName())}
	alines := []string{}
	values := []interface{}{}
	for _, assignment := range q.assignments {
		s, v := assignment.SQL()
		alines = append(alines, s)
		values = append(values, v...)
	}
	sql = append(sql, strings.Join(alines, ", "))
	// WHERE
	if where, whereValues, ok := q.where.sql(); ok {
		sql = append(sql, "WHERE", where)
		values = append(values, whereValues...)
	}
	// ORDER BY
	if s, ok := q.orderBy.sql(); ok {
		sql = append(sql, s)
	}
	// LIMIT
	if NoLimit != limit {
		sql = append(sql, fmt.Sprintf("LIMIT %d", limit))
	}
	return strings.Join(sql, " "), values, q.err
}

// ParameterizedSQL representation of this query.
func (q *UpdateQuery) ParameterizedSQL(limit int) (string, error) {
	if nil != q.err {
		return "", q.err
	}
	if len(q.assignments) == 0 {
		return "", errors.New("no assignments in update query")
	}
	sql := []string{fmt.Sprintf("UPDATE `%s` SET ", q.tableReference.GetName())}
	alines := []string{}
	values := []interface{}{}
	for _, assignment := range q.assignments {
		s, v := assignment.SQL()
		alines = append(alines, s)
		values = append(values, v...)
	}
	sql = append(sql, strings.Join(alines, ", "))
	// WHERE
	if where, _, ok := q.where.sql(); ok {
		sql = append(sql, "WHERE", where)
	}
	// ORDER BY
	if s, ok := q.orderBy.sql(); ok {
		sql = append(sql, s)
	}
	// LIMIT
	if NoLimit != limit {
		sql = append(sql, fmt.Sprintf("LIMIT %d", limit))
	}
	return strings.Join(sql, " "), q.err
}
