package qb

import (
	"fmt"
)

type literal struct {
	value any
}

func (c literal) GetName() string {
	return fmt.Sprintf("%s", c.value)
}

func (c literal) GetTables() []string {
	return []string{}
}

func (c literal) ParameterizedSQL() (string, []any) {
	return "?", []any{c.value}
}

// Literal value to use as a SelectExpression
func Literal(value any) SelectExpression {
	return literal{value: value}
}
