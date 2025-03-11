package qb

import (
	"fmt"
	"strings"
)

// InsertSQLParameters is a utility function that inserts parameters into parameterized SQL.
func InsertSQLParameters(sql string, params []interface{}) string {
	if len(params) == 0 {
		return sql
	}

	sql = strings.Replace(sql, "?", "%#v", -1)
	sql = fmt.Sprintf(sql, params...)
	return sql
}
