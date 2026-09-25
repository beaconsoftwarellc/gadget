package qb

import (
	"strconv"
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

type format struct {
	expression    SelectExpression
	decimalPlaces int
	locale        string
	alias         string
}

func (f format) GetName() string {
	return f.alias
}

func (f format) GetTables() []string {
	return f.expression.GetTables()
}

func (f format) ParameterizedSQL() (string, []any) {
	sql, expValues := f.expression.ParameterizedSQL()
	parts := []string{
		"FORMAT(",
		sql,
		", ",
		strconv.Itoa(f.decimalPlaces),
	}
	if !stringutil.IsEmpty(f.locale) {
		parts = append(parts, ", ", f.locale)
	}
	parts = append(parts, ")")
	if !stringutil.IsEmpty(f.alias) {
		parts = append(parts, " AS `", f.alias, "`")
	}
	return strings.Join(parts, ""), expValues
}

// Format the Select expression to the specified decimal places leverage the
// optional locale and alias
func Format(expression SelectExpression, decimalPlaces int, locale, alias string) SelectExpression {
	return format{expression: expression, decimalPlaces: decimalPlaces, locale: locale, alias: alias}
}
