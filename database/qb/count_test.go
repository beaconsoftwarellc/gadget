package qb

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	assert1 "github.com/stretchr/testify/assert"
)

func Test_NewCountExpression(t *testing.T) {
	assert := assert1.New(t)
	expected := generator.String(20)
	expression := NewCountExpression(expected)
	assert.Equal(1, len(expression.GetTables()))
	assert.Equal(expected, expression.GetTables()[0])

	expectedName := fmt.Sprintf("COUNT(%s)", expected)
	assert.Equal(expectedName, expression.GetName())

	expectedSQL := "COUNT(*) as count"
	actualSQL, values := expression.ParameterizedSQL()
	assert.Equal(expectedSQL, actualSQL)
	assert.Nil(values)
}

func Test_NewCountDistinctExpression(t *testing.T) {
	assert := assert1.New(t)
	expectedTable := generator.String(20)
	expectedField := generator.String(20)
	expression := NewCountDistinctExpression(TableField{
		Name:  expectedField,
		Table: expectedTable,
	})
	assert.Equal(1, len(expression.GetTables()))
	assert.Equal(expectedTable, expression.GetTables()[0])

	expectedName := fmt.Sprintf("COUNT(DISTINCT %s)", expectedField)
	assert.Equal(expectedName, expression.GetName())

	expectedSQL := fmt.Sprintf("COUNT(DISTINCT `%s`.`%s`) as count", expectedTable, expectedField)
	actualSQL, values := expression.ParameterizedSQL()
	assert.Equal(expectedSQL, actualSQL)
	assert.Nil(values)
}

func Test_NewCountDistinct(t *testing.T) {
	assert := assert1.New(t)
	expression := TableField{
		Name:  "foo",
		Table: "bar",
	}
	expression2 := TableField{
		Name:  "baz",
		Table: "quux",
	}

	countDistinct := NewCountDistinct([]SelectExpression{expression, expression2})

	expectedName := "COUNT(DISTINCT foo, baz)"
	assert.Equal(expectedName, countDistinct.GetName())

	expectedSQL := "COUNT(DISTINCT `bar`.`foo`, `quux`.`baz`) as count"
	actualSQL, values := countDistinct.ParameterizedSQL()
	assert.Equal(expectedSQL, actualSQL)
	assert.Nil(values)
}
