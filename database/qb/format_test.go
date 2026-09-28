package qb

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestFormat(t *testing.T) {
	expectedField := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedDecimals := generator.Int()
	expectedLocale := language.Japanese

	expression := Format(expectedField, expectedDecimals, expectedLocale)
	assert.Equal(t, expectedField.GetName(), expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("FORMAT(%s, %d, \"ja_JP\")", expectedField.SQL(), expectedDecimals), actualSql)
	assert.Empty(t, actualParams)

	expression = Format(expectedField, expectedDecimals, language.Und)
	assert.Contains(t, expression.GetTables(), expectedField.Table)
	actualSql, actualParams = expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("FORMAT(%s, %d)", expectedField.SQL(), expectedDecimals), actualSql)
	assert.Empty(t, actualParams)
}
