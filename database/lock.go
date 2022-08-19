package database

import (
	"fmt"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

const (
	acquireLockQueryFormat = "SELECT GET_LOCK('%s', %d) AS STATUS"
	releaseLockQueryFormat = "SELECT RELEASE_LOCK('%s') AS STATUS"
)

// AcquireDatabaseLock with the specified name and timeout. Returns boolean indicating whether the
// lock was acquired or error on failure to execute.
// See: https://dev.mysql.com/doc/refman/5.7/en/locking-functions.html
func AcquireDatabaseLock(db *Database, name string, timeout time.Duration) (bool, errors.TracerError) {
	var err error
	var target []*StatusResult
	err = db.DB.Select(&target, fmt.Sprintf(acquireLockQueryFormat, name, int(timeout.Seconds())))
	if nil != err {
		return false, errors.Wrap(err)
	}
	if len(target) < 1 {
		return false, errors.New("no rows returned from acquire lock query")
	}
	return target[0].Status == 1, nil
}

// ReleaseDatabaseLock with the specified name
// See: https://dev.mysql.com/doc/refman/5.7/en/locking-functions.html
func ReleaseDatabaseLock(db *Database, name string) errors.TracerError {
	var target []*StatusResult
	return errors.Wrap(db.DB.Select(&target, fmt.Sprintf(releaseLockQueryFormat, name)))
}
