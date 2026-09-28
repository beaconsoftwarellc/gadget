package qb

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

type format struct {
	expression    SelectExpression
	decimalPlaces int
	languageTag   language.Tag
}

func (f format) GetName() string {
	return f.expression.GetName()
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
	if f.languageTag != language.Und {
		parts = append(parts, ", ", f.locale())
	}
	parts = append(parts, ")", " AS `", f.GetName(), "`")
	return strings.Join(parts, ""), expValues
}

func (f format) locale() string {
	var (
		base, _   = f.languageTag.Base()
		region, _ = f.languageTag.Region()
	)
	return fmt.Sprintf("\"%s_%s\"", base, region)
}

// Format the Select expression to the specified decimal places leverage the
// optional locale and alias
func Format(expression SelectExpression, decimalPlaces int, languageTag language.Tag) SelectExpression {
	return format{expression: expression,
		decimalPlaces: decimalPlaces,
		languageTag:   languageTag,
	}
}
