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

	expression := Add([]SelectExpression{expectedField1, expectedField2}, expectedAlias)

	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField1.Table)
	assert.Contains(t, expression.GetTables(), expectedField2.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("%s + %s AS `%s`", expectedField1.SQL(), expectedField2.SQL(), expectedAlias), actualSql)
	assert.Empty(t, actualParams)
}

func Test_NewSubtract(t *testing.T) {
	expectedField1 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedField2 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	expression := Subtract([]SelectExpression{expectedField1, expectedField2}, expectedAlias)

	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField1.Table)
	assert.Contains(t, expression.GetTables(), expectedField2.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("%s - %s AS `%s`", expectedField1.SQL(), expectedField2.SQL(), expectedAlias), actualSql)
	assert.Empty(t, actualParams)
}

func Test_NewDivide(t *testing.T) {
	expectedField1 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedField2 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	expression := Divide([]SelectExpression{expectedField1, expectedField2}, expectedAlias)

	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField1.Table)
	assert.Contains(t, expression.GetTables(), expectedField2.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("%s / %s AS `%s`", expectedField1.SQL(), expectedField2.SQL(), expectedAlias), actualSql)
	assert.Empty(t, actualParams)
}

func Test_NewMultiply(t *testing.T) {
	expectedField1 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedField2 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	expression := Multiply([]SelectExpression{expectedField1, expectedField2}, expectedAlias)

	assert.Equal(t, expectedAlias, expression.GetName())
	assert.Contains(t, expression.GetTables(), expectedField1.Table)
	assert.Contains(t, expression.GetTables(), expectedField2.Table)
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("%s * %s AS `%s`", expectedField1.SQL(), expectedField2.SQL(), expectedAlias), actualSql)
	assert.Empty(t, actualParams)
}

func Test_Multiply_EmptyAliasOmitsAsClause(t *testing.T) {
	expectedField1 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedField2 := TableField{Name: generator.String(20), Table: generator.String(20)}

	expression := Multiply([]SelectExpression{expectedField1, expectedField2}, "")

	assert.Empty(t, expression.GetName())
	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("%s * %s", expectedField1.SQL(), expectedField2.SQL()), actualSql)
	assert.Empty(t, actualParams)
}

func Test_Multiply_ComposesInsideSum(t *testing.T) {
	expectedField1 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedField2 := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	expression := Sum(Multiply([]SelectExpression{expectedField1, expectedField2}, ""), expectedAlias)

	actualSql, actualParams := expression.ParameterizedSQL()
	assert.Equal(t, fmt.Sprintf("SUM(%s * %s) AS `%s`", expectedField1.SQL(), expectedField2.SQL(), expectedAlias), actualSql)
	assert.Empty(t, actualParams)
}
