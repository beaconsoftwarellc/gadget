package utility

import (
	"fmt"
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func TestSetMultiStatement(t *testing.T) {
	assert := assert1.New(t)
	var tests = []struct {
		name     string
		argument string
		expected string
	}{
		{
			name:     "empty",
			argument: "",
			expected: fmt.Sprintf("?%s", multiStatementTrueQS),
		},
		{
			name:     "no qs",
			argument: "sql:dababaycopmuter:root:neverguessit/mysql",
			expected: fmt.Sprintf("sql:dababaycopmuter:root:neverguessit/mysql?%s",
				multiStatementTrueQS),
		},
		{
			name:     "has qs",
			argument: "sql:rowland?columns=true",
			expected: fmt.Sprintf("sql:rowland?columns=true&%s", multiStatementTrueQS),
		},
		{
			name:     "already set",
			argument: fmt.Sprintf("sql:rowland?columns=true&%s", multiStatementTrueQS),
			expected: fmt.Sprintf("sql:rowland?columns=true&%s", multiStatementTrueQS),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := SetMultiStatement(tc.argument)
			assert.Equal(tc.expected, actual)
		})
	}
}
