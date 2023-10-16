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
	loggedQueries  map[string]time.Duration
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

func (tx *slowQueryLoggerTx) Preparex(query string) (*sqlx.Stmt, error) {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.implementation.Preparex(query)
}

func (tx *slowQueryLoggerTx) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.implementation.PrepareNamed(query)
}

func (tx *slowQueryLoggerTx) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.implementation.QueryRowx(query, args...)
}

func (tx *slowQueryLoggerTx) Select(dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	defer func() { tx.logSlow(query, time.Since(start)) }()
	return tx.implementation.Select(dest, query, args...)
}

func (tx *slowQueryLoggerTx) Commit() error {
	return tx.implementation.Commit()
}

func (tx *slowQueryLoggerTx) Rollback() error {
	return tx.implementation.Rollback()
}

func (tx *slowQueryLoggerTx) logSlow(query string, elapsed time.Duration) {
	if elapsed <= tx.slow {
		return
	}

	// do not log the slow query if it has already been logged with a slower time
	logged, ok := tx.loggedQueries[query]
	if ok && logged >= elapsed {
		return
	}

	err := errors.New(
		"[%s] query execution time: %s query: %s", tx.id, elapsed, query)
	_ = tx.log.Error(err)
	tx.loggedQueries[query] = elapsed
}
