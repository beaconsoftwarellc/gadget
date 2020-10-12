package qb

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/errors"
)

// InsertQuery for inserting a row into the database.
type InsertQuery struct {
	columns           []TableField
	values            [][]interface{}
	onDuplicate       []TableField
	onDuplicateValues []interface{}
	err               error
}

// Values to be inserted. Call multiple times to insert multiple rows.
func (q *InsertQuery) Values(values ...interface{}) *InsertQuery {
	if len(values) != len(q.columns) {
		q.err = errors.New("insert field/value count mismatch")
	} else {
		q.values = append(q.values, values)
	}
	return q
}

// OnDuplicate update these fields / values
func (q *InsertQuery) OnDuplicate(fields []TableField, values ...interface{}) *InsertQuery {
	q.onDuplicate = append(q.onDuplicate, fields...)
	q.onDuplicateValues = values
	return q
}

// GetAlias of the passed table name in this query.
func (q *InsertQuery) GetAlias(tableName string) string {
	return tableName
}

// SQL that represents this insert query.
func (q *InsertQuery) SQL() (string, []interface{}, error) {
	if len(q.columns) == 0 {
		return "", nil, errors.New("no columns specified for insert")
	}
	colExp := make([]string, len(q.columns))
	qms := make([]string, len(q.columns))
	for i, col := range q.columns {
		colExp[i] = col.SQL()
		if col.Table != q.columns[0].Table {
			return "", nil, errors.New("insert columns must be from the same table")
		}
		qms[i] = "?"
	}
	valExp := fmt.Sprintf("(%s)", strings.Join(qms, ", "))
	valExps := make([]string, len(q.values))
	values := []interface{}{}
	for i, valGrp := range q.values {
		valExps[i] = valExp
		values = append(values, valGrp...)
	}
	onDuplicate := ""
	if len(q.onDuplicate) > 0 {
		if len(q.values) > 1 {
			return "", nil, errors.New("cannot use on duplicate with multi-insert")
		}
		updateFields := make([]string, len(q.onDuplicate))
		for _, col := range q.onDuplicate {
			if col.Table != q.columns[0].Table {
				return "", nil, errors.New("insert columns must be from the same table")
			}
			for i, col := range q.onDuplicate {
				updateFields[i] = fmt.Sprintf("%s = ?", col.SQL())
			}
		}
		values = append(values, q.onDuplicateValues...)
		onDuplicate = " ON DUPLICATE KEY UPDATE " + strings.Join(updateFields, ", ")
	}
	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s%s", q.columns[0].Table, strings.Join(colExp, ", "),
		strings.Join(valExps, ", "), onDuplicate), values, q.err
}

// ParameterizedSQL that represents this insert query.
func (q *InsertQuery) ParameterizedSQL() (string, error) {
	if len(q.columns) == 0 {
		return "", errors.New("no columns specified for insert")
	}
	colExp := make([]string, len(q.columns))
	qms := make([]string, len(q.columns))
	for i, col := range q.columns {
		colExp[i] = col.SQL()
		if col.Table != q.columns[0].Table {
			return "", errors.New("insert columns must be from the same table")
		}
		qms[i] = ":" + col.GetName()
	}
	onDuplicate := ""
	if len(q.onDuplicate) > 0 {
		if len(q.values) > 1 {
			return "", errors.New("cannot use on duplicate with multi-insert")
		}
		updateFields := make([]string, len(q.onDuplicate))
		for i, field := range q.onDuplicate {
			updateFields[i] = fmt.Sprintf("%s = :%s", field.SQL(), field.GetName())
		}
		onDuplicate = " ON DUPLICATE KEY UPDATE " + strings.Join(updateFields, ", ")
	}
	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)%s", q.columns[0].Table, strings.Join(colExp, ", "),
		strings.Join(qms, ", "), onDuplicate), q.err
}
