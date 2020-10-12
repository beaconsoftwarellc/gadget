package qb

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/errors"
)

/*
   DELETE [LOW_PRIORITY] [QUICK] [IGNORE]
   tbl_name[.*] [, tbl_name[.*]] ...
   FROM table_references
   [WHERE where_condition]
*/

// DeleteQuery for removing rows from a database
type DeleteQuery struct {
	tables []Table
	from   Table
	joins  []*Join
	where  *whereCondition
	// NICE TO HAVE: Add orderby and limit logic, order by and limit only apply to single table case
	err error
}

// GetAlias of the passed table name in this query
func (q *DeleteQuery) GetAlias(tableName string) string {
	return tableName
}

// From sets the primary table the query will find rows in.
func (q *DeleteQuery) From(table Table) *DeleteQuery {
	q.from = table
	return q
}

// InnerJoin with another table in the database.
func (q *DeleteQuery) InnerJoin(table Table) *Join {
	join := NewJoin(Inner, Right, table)
	q.joins = append(q.joins, join)
	return join
}

// OuterJoin with another table in the database.
func (q *DeleteQuery) OuterJoin(direction JoinDirection, table Table) *Join {
	join := NewJoin(Outer, direction, table)
	q.joins = append(q.joins, join)
	return join
}

// Where determines what rows to delete from.
func (q *DeleteQuery) Where(condition *ConditionExpression) *DeleteQuery {
	q.where.expression = condition
	return q
}

// Validate that this query is executable
func (q *DeleteQuery) Validate() bool {
	if nil == q.from && len(q.tables) == 0 {
		q.err = errors.New("at least one table must be specified to delete from")
		return false
	}

	if nil == q.where.expression {
		q.err = errors.New("delete requires a where clause")
		return false
	}

	for _, join := range q.joins {
		if nil != join.err {
			q.err = join.err
			return false
		}
	}

	return true
}

// SQL representation of this delete query.
func (q *DeleteQuery) SQL() (string, []interface{}, error) {
	if !q.Validate() {
		return "", nil, q.err
	}
	lines := []string{"DELETE"}
	values := []interface{}{}
	rowsInLines := make([]string, len(q.tables))

	if len(q.tables) == 1 && nil == q.from {
		q.from = q.tables[0]
	} else {
		for i, table := range q.tables {
			rowsInLines[i] = fmt.Sprintf("`%s`", table.GetName())
		}
		lines = append(lines, strings.Join(rowsInLines, ", "))
	}

	// FROM
	lines = append(lines, fmt.Sprintf("FROM `%s`", q.from.GetName()))

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

	return strings.Join(lines, " "), values, q.err
}
