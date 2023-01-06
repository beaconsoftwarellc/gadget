package database

import (
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

// CommitOrRollback will rollback on an errors.TracerError otherwise commit
func CommitOrRollback(tx transaction.Transaction, err error, logger log.Logger) errors.TracerError {
	if err != nil {
		logger.Error(tx.Rollback())
		return errors.Wrap(err)
	}
	return errors.Wrap(tx.Commit())
}

func obfuscateConnection(connection string) string {
	obfuscateIndex := strings.LastIndex(connection, "@")
	obfuscatedConnection := connection
	if obfuscateIndex > 0 {
		// no '@' means the credentials are not part of the connection and we do not
		// need to obfuscate
		obfuscatedConnection = stringutil.ObfuscateLeft(obfuscatedConnection,
			obfuscateIndex, "*")
	}
	return obfuscatedConnection
}
