package qb

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/stretchr/testify/assert"
)

func Test_NewSumExpression(t *testing.T) {
	expectedTable := generator.String(20)
	exptectedField := TableField{Name: generator.String(20)}

	expression := NewSumExpression(expectedTable, exptectedField)
	assert.Equal(t, 1, len(expression.GetTables()))
	assert.Equal(t, expectedTable, expression.GetTables()[0])

	expectedName := fmt.Sprintf("SUM(%s)", exptectedField.GetName())
	assert.Equal(t, expectedName, expression.GetName())

	expectedSQL := fmt.Sprintf("SUM(%s) as sum", exptectedField.GetName())
	assert.Equal(t, expectedSQL, expression.SQL())
}
