package qb

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/stretchr/testify/assert"
)

func Test_NewIf(t *testing.T) {
	conditionField := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedConditionValue := generator.String(20)
	trueField := TableField{Name: generator.String(20), Table: generator.String(20)}
	falseField := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	condition := conditionField.Equal(expectedConditionValue)
	expression := If(condition, trueField, falseField, expectedAlias)

	assert.Equal(t, expectedAlias, expression.GetName())
	conditionSQL, conditionValues := condition.SQL()
	actualSQL, actualParams := expression.ParameterizedSQL()
	assert.Equal(t,
		fmt.Sprintf("IF(%s, %s, %s) AS `%s`", conditionSQL, trueField.SQL(), falseField.SQL(), expectedAlias),
		actualSQL,
	)
	assert.Equal(t, conditionValues, actualParams)
}

func Test_If_EmptyAliasOmitsAsClause(t *testing.T) {
	conditionField := TableField{Name: generator.String(20), Table: generator.String(20)}
	trueField := TableField{Name: generator.String(20), Table: generator.String(20)}
	falseField := TableField{Name: generator.String(20), Table: generator.String(20)}

	condition := conditionField.Equal(generator.String(20))
	expression := If(condition, trueField, falseField, "")

	assert.Empty(t, expression.GetName())
	conditionSQL, conditionValues := condition.SQL()
	actualSQL, actualParams := expression.ParameterizedSQL()
	assert.Equal(t,
		fmt.Sprintf("IF(%s, %s, %s)", conditionSQL, trueField.SQL(), falseField.SQL()),
		actualSQL,
	)
	assert.Equal(t, conditionValues, actualParams)
}

func Test_If_GetTablesUnionsAllBranches(t *testing.T) {
	conditionField := TableField{Name: generator.String(20), Table: generator.String(20)}
	trueField := TableField{Name: generator.String(20), Table: generator.String(20)}
	falseField := TableField{Name: generator.String(20), Table: generator.String(20)}

	condition := conditionField.Equal(generator.String(20))
	expression := If(condition, trueField, falseField, generator.String(20))

	tables := expression.GetTables()
	assert.Contains(t, tables, conditionField.Table)
	assert.Contains(t, tables, trueField.Table)
	assert.Contains(t, tables, falseField.Table)
}

func Test_If_ComposesInsideSum(t *testing.T) {
	conditionField := TableField{Name: generator.String(20), Table: generator.String(20)}
	trueField := TableField{Name: generator.String(20), Table: generator.String(20)}
	falseField := TableField{Name: generator.String(20), Table: generator.String(20)}
	expectedAlias := generator.String(20)

	condition := conditionField.Equal(generator.String(20))
	expression := Sum(If(condition, trueField, falseField, ""), expectedAlias)

	conditionSQL, conditionValues := condition.SQL()
	actualSQL, actualParams := expression.ParameterizedSQL()
	assert.Equal(t,
		fmt.Sprintf("SUM(IF(%s, %s, %s)) AS `%s`", conditionSQL, trueField.SQL(), falseField.SQL(), expectedAlias),
		actualSQL,
	)
	assert.Equal(t, conditionValues, actualParams)
}
