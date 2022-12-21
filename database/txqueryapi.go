package database

import (
	"database/sql"
)

type TransactionQuery interface {
	Transaction

	NamedExec(query string, arg interface{}) (sql.Result, error)
	Exec(query string, arg ...interface{}) (sql.Result, error)
}

func newTxQueryAPI(db *Database) *txQueryAPI {
	return &txQueryAPI{txAPI: NewTxAPI(db, nil)}
}

func newTxQueryAPIFromTxAPI(txAPI *txAPI) *txQueryAPI {
	return &txQueryAPI{txAPI: txAPI}
}

type txQueryAPI struct {
	*txAPI
}

func (t *txQueryAPI) NamedExec(query string, arg interface{}) (sql.Result, error) {
	return t.tx.NamedExec(query, arg)
}

func (t *txQueryAPI) Exec(query string, args ...any) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}
