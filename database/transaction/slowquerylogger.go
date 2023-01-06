package transaction

import (
	"database/sql"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"

	"time"

	"github.com/jmoiron/sqlx"
)

type slowQueryLoggerTx struct {
	implementation Implementation
	slow           time.Duration
	log            log.Logger
	id             string
}

func (tx *slowQueryLoggerTx) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.implementation.NamedQuery(query, arg)
}

func (tx *slowQueryLoggerTx) Exec(query string, args ...any) (sql.Result, error) {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.implementation.Exec(query, args...)
}

func (tx *slowQueryLoggerTx) NamedExec(query string, arg interface{}) (sql.Result, error) {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.implementation.NamedExec(query, arg)
}

func (tx *slowQueryLoggerTx) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.implementation.QueryRowx(query, args)
}

func (tx *slowQueryLoggerTx) Select(dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.implementation.Select(dest, query, args)
}

func (tx *slowQueryLoggerTx) Commit() error {
	return tx.implementation.Commit()
}

func (tx *slowQueryLoggerTx) Rollback() error {
	return tx.implementation.Rollback()
}

func (tx *slowQueryLoggerTx) logSlow(query string, elapsed time.Duration) {
	err := errors.New(
		"[%s] query execution time: %s query: %s", tx.id, elapsed, query)
	if elapsed > tx.slow {
		_ = tx.log.Error(err)
	} else {
		_ = tx.log.Debug(err)
	}
}
