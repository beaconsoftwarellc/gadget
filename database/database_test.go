package database

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	assert1 "github.com/stretchr/testify/assert"
)

func Test_database_enforceLimits(t *testing.T) {
	var tests = []struct {
		name          string
		maxQueryLimit uint
		options       *record.ListOptions
		expected      *record.ListOptions
	}{
		{
			name:          "no limit",
			maxQueryLimit: 0,
			options:       &record.ListOptions{Limit: 100, Offset: 0},
			expected:      &record.ListOptions{Limit: 100, Offset: 0},
		},
		{
			name:          "limit enforced",
			maxQueryLimit: 10,
			options:       &record.ListOptions{Limit: 100, Offset: 0},
			expected:      &record.ListOptions{Limit: 10, Offset: 0},
		},
		{
			name:          "nil gets defaults",
			maxQueryLimit: 20,
			options:       nil,
			expected:      &record.ListOptions{Limit: 20, Offset: 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert1.New(t)
			conf := &InstanceConfig{}
			conf.MaxLimit = tc.maxQueryLimit
			database := &api{configuration: conf}
			database.enforceLimits(tc.options)
			assert.Equal(tc.expected, tc.options)
		})
	}
}
