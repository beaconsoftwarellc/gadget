package database

import (
	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// Count the number of rows in the passed query
func Count(db API, table qb.Table, query *qb.SelectQuery) (int32, error) {
	var target []*qb.RowCount

	err := db.Select(&target, query.SelectFrom(qb.NewCountExpression(table.GetName())))
	if err != nil {
		return 0, err
	}
	if len(target) == 0 {
		return 0, errors.New("[COM.DB.1] row count query execution failed (no rows)")
	}
	return int32(target[0].Count), nil
}

// CountWhere the number of rows in the passed query
func CountWhere(db API, table qb.Table, condition *qb.ConditionExpression) (int32, error) {
	query := qb.Select(qb.NewCountExpression(table.GetName())).
		From(table).
		Where(condition)

	return Count(db, table, query)
}

// CountTableRows from the passed database and table name
func CountTableRows(db API, table qb.Table) (int32, error) {
	return CountWhere(db, table, nil)
}
