package qb

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/stretchr/testify/assert"
)

func Test_NewAdd(t *testing.T) {
	expectedField1 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedField2 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	expression := Add([]TableField{expectedField1, expectedField2}, expectedAlias)

	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField1.Table)
	assert.Contains(t, expression.GetTables(), expectedField2.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("%s + %s AS `%s`", expectedField1.SQL(), expectedField2.SQL(), expectedAlias), actualSql)
	assert.Nil(t, actualParams)
}

func Test_NewSubtract(t *testing.T) {
	expectedField1 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedField2 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	expression := Subtract([]TableField{expectedField1, expectedField2}, expectedAlias)

	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField1.Table)
	assert.Contains(t, expression.GetTables(), expectedField2.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("%s - %s AS `%s`", expectedField1.SQL(), expectedField2.SQL(), expectedAlias), actualSql)
	assert.Nil(t, actualParams)
}

func Test_NewDivide(t *testing.T) {
	expectedField1 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedField2 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	expression := Divide([]TableField{expectedField1, expectedField2}, expectedAlias)

	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField1.Table)
	assert.Contains(t, expression.GetTables(), expectedField2.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("%s / %s AS `%s`", expectedField1.SQL(), expectedField2.SQL(), expectedAlias), actualSql)
	assert.Nil(t, actualParams)
}

func Test_NewMultiply(t *testing.T) {
	expectedField1 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedField2 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	expression := Multiply([]TableField{expectedField1, expectedField2}, expectedAlias)

	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField1.Table)
	assert.Contains(t, expression.GetTables(), expectedField2.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("%s * %s AS `%s`", expectedField1.SQL(), expectedField2.SQL(), expectedAlias), actualSql)
	assert.Nil(t, actualParams)
}
