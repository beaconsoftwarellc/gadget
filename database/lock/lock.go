package lock

import (
	"fmt"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/database"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

const (
	acquireLockQueryFormat = "SELECT GET_LOCK('%s', %d) AS STATUS"
	releaseLockQueryFormat = "SELECT RELEASE_LOCK('%s') AS STATUS"
)

// StatusResult is for capturing output from a function call on the database, you must use
// an 'AS STATUS' clause in your query in order for mapping to work correctly.
type StatusResult struct {
	// Status as returned by a function call usually
	Status int `db:"STATUS"`
}

// Acquire with the specified name and timeout. Returns boolean indicating whether the
// lock was acquired or error on failure to execute.
// See: https://dev.mysql.com/doc/refman/5.7/en/locking-functions.html
func Acquire(db database.Client, name string, timeout time.Duration) (bool, errors.TracerError) {
	var err error
	var target []*StatusResult
	err = db.Select(&target,
		fmt.Sprintf(acquireLockQueryFormat, name, int(timeout.Seconds())))
	if nil != err {
		return false, errors.Wrap(err)
	}
	if len(target) < 1 {
		return false, errors.New("no rows returned from acquire lock query")
	}
	return target[0].Status == 1, nil
}

// Release with the specified name
// See: https://dev.mysql.com/doc/refman/5.7/en/locking-functions.html
func Release(db database.Client, name string) errors.TracerError {
	var target []*StatusResult
	return errors.Wrap(db.Select(&target, fmt.Sprintf(releaseLockQueryFormat, name)))
}
