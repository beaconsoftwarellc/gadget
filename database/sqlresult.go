package database

import (
	"database/sql"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// implements sql.Result
type result struct {
	rowsAffected int64
}

// Consume the passed result and add it to this result
func (r *result) Consume(sqlResult sql.Result) error {
	rows, err := sqlResult.RowsAffected()
	if nil != err {
		return err
	}
	r.rowsAffected += rows
	return nil
}

func (r *result) LastInsertId() (int64, error) {
	return 0, errors.New("not supported")
}

func (r *result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}
