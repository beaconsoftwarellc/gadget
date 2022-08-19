package database

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func Test_database_enforceLimits(t *testing.T) {
	var tests = []struct {
		name          string
		maxQueryLimit uint
		options       *ListOptions
		expected      *ListOptions
	}{
		{
			name:          "no limit",
			maxQueryLimit: 0,
			options:       &ListOptions{Limit: 100, Offset: 0},
			expected:      &ListOptions{Limit: 100, Offset: 0},
		},
		{
			name:          "limit enforced",
			maxQueryLimit: 10,
			options:       &ListOptions{Limit: 100, Offset: 0},
			expected:      &ListOptions{Limit: 10, Offset: 0},
		},
		{
			name:          "nil gets defaults",
			maxQueryLimit: 20,
			options:       nil,
			expected:      &ListOptions{Limit: 20, Offset: 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert1.New(t)
			spec := newSpecification()
			conf := spec.DB.Configuration.(*specification)
			conf.QueryLimit = tc.maxQueryLimit

			actual := conf.DB.enforceLimits(tc.options)
			assert.Equal(tc.expected, actual)
		})
	}
}
