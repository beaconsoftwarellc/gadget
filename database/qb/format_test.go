package qb

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	expectedField := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedDecimals := generator.Int()
	expectedLocale := generator.String(5)
	expectedAlias := generator.String(20)

	expression := Format(expectedField, expectedDecimals, expectedLocale, expectedAlias)
	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("FORMAT(%s, %d, %s) AS `%s`", expectedField.SQL(), expectedDecimals, expectedLocale, expectedAlias), actualSql)
	assert.Empty(t, actualParams)

	expression = Format(expectedField, expectedDecimals, "", expectedAlias)
	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField.Table)
	actualSql, actualParams = expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("FORMAT(%s, %d) AS `%s`", expectedField.SQL(), expectedDecimals, expectedAlias), actualSql)
	assert.Empty(t, actualParams)

	expression = Format(expectedField, expectedDecimals, "", "")
	assert.Equal(t, "", expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField.Table)
	actualSql, actualParams = expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("FORMAT(%s, %d)", expectedField.SQL(), expectedDecimals), actualSql)
	assert.Empty(t, actualParams)
}
