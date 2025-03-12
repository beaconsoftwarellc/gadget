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
