package qb

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// InsertQuery for inserting a row into the database.
type InsertQuery struct {
	columns           []TableField
	values            [][]any
	onDuplicate       []TableField
	onDuplicateValues []any
	err               error
}

// Values to be inserted. Call multiple times to insert multiple rows.
func (q *InsertQuery) Values(values ...any) *InsertQuery {
	if len(values) != len(q.columns) {
		q.err = errors.New("insert field/value count mismatch")
	} else {
		q.values = append(q.values, values)
	}
	return q
}

// OnDuplicate update these fields / values
func (q *InsertQuery) OnDuplicate(fields []TableField) *InsertQuery {
	q.onDuplicate = append(q.onDuplicate, fields...)
	return q
}

// GetAlias of the passed table name in this query.
func (q *InsertQuery) GetAlias(tableName string) string {
	return tableName
}

// SQL that represents this insert query.
func (q *InsertQuery) SQL() (string, []any, error) {
	sql, err := q.getSQL(false)
	if err != nil {
		return "", nil, err
	}
	values := []any{}
	for _, valGrp := range q.values {
		values = append(values, valGrp...)
	}
	return sql, values, q.err
}

// ParameterizedSQL that represents this insert query.
func (q *InsertQuery) ParameterizedSQL() (string, error) {
	return q.getSQL(true)
}

func (q *InsertQuery) getSQL(parameterized bool) (string, error) {
	if len(q.columns) == 0 {
		return "", errors.New("no columns specified for insert")
	}
	colExp := make([]string, len(q.columns))
	valuePlaces := make([]string, len(q.columns))
	for i, col := range q.columns {
		colExp[i] = col.SQL()
		if col.Table != q.columns[0].Table {
			return "", errors.New("insert columns must be from the same table")
		}
		if parameterized {
			valuePlaces[i] = ":" + col.GetName()
		} else {
			valuePlaces[i] = "?"
		}
	}
	valExp := fmt.Sprintf("(%s)", strings.Join(valuePlaces, ", "))
	if !parameterized {
		valExps := make([]string, len(q.values))
		for i := range q.values {
			valExps[i] = valExp
		}
		valExp = strings.Join(valExps, ", ")
	}
	onDuplicate := ""
	if len(q.onDuplicate) > 0 {
		updateFields := make([]string, len(q.onDuplicate))
		for _, col := range q.onDuplicate {
			if col.Table != q.columns[0].Table {
				return "", errors.New("duplicate columns must be from the same table")
			}
			for i, col := range q.onDuplicate {
				updateFields[i] = fmt.Sprintf("%s = VALUES(%s)", col.SQL(), col.SQL())
			}
		}
		onDuplicate = " ON DUPLICATE KEY UPDATE " + strings.Join(updateFields, ", ")
	}
	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s%s", q.columns[0].Table, strings.Join(colExp, ", "),
		valExp, onDuplicate), q.err
}
