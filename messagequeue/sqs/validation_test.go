package sqs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
	"github.com/stretchr/testify/assert"
)

func TestNameIsValid(t *testing.T) {
	var tests = []struct {
		name     string
		s        string
		expected string
	}{
		{
			name:     "too short",
			s:        "",
			expected: "name character count out of bounds [1, 256] (0)",
		},
		{
			name:     "too long",
			s:        generator.String(257),
			expected: "name character count out of bounds [1, 256] (257)",
		},
		{
			name:     "bad prefix 1",
			s:        fmt.Sprintf(".%s", generator.String(20)),
			expected: dotError.Error(),
		},
		{
			name:     "bad prefix 2",
			s:        fmt.Sprintf("%s%s", prohibitedAWS, generator.String(20)),
			expected: prohibitedPrefixError.Error(),
		},
		{
			name: "bad prefix 3",
			s: fmt.Sprintf("%s%s", strings.ToUpper(prohibitedAWS),
				generator.String(20)),
			expected: prohibitedPrefixError.Error(),
		},
		{
			name: "bad prefix 4",
			s: fmt.Sprintf("%s%s", strings.ToUpper(prohibitedAmazon),
				generator.String(20)),
			expected: prohibitedPrefixError.Error(),
		},
		{
			name:     "bad prefix 5",
			s:        fmt.Sprintf("%s%s", prohibitedAmazon, generator.String(20)),
			expected: prohibitedPrefixError.Error(),
		},
		{
			name:     "bad suffix",
			s:        fmt.Sprintf("%s%s", generator.String(20), "."),
			expected: dotError.Error(),
		},
		{
			name:     "bad sequence",
			s:        fmt.Sprintf("%s..%s", generator.String(20), generator.String(20)),
			expected: dotError.Error(),
		},
		{
			name:     "min",
			s:        generator.String(1),
			expected: "",
		},
		{
			name:     "max",
			s:        generator.String(256),
			expected: "",
		},
		{
			name:     "typical",
			s:        generator.String(32),
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			actual := NameIsValid(test.s)
			if stringutil.IsWhiteSpace(test.expected) {
				assert.NoError(actual)
			} else {
				assert.EqualError(actual, test.expected)
			}
		})
	}
}

func TestBodyIsValid(t *testing.T) {
	var tests = []struct {
		name     string
		s        string
		expected string
	}{
		{
			name:     "below minimum char count",
			s:        "",
			expected: bodyMinimumError,
		},
		{
			name:     "above maximum byte count",
			s:        generator.String((maxBodyKilobytes+1)*1024 + 1),
			expected: "body cannot exceed 255 kilobytes (was 256 kb)",
		},
		{
			name:     "typical",
			s:        generator.String(32),
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			actual := BodyIsValid(test.s)
			if stringutil.IsWhiteSpace(test.expected) {
				assert.NoError(actual)
			} else {
				assert.EqualError(actual, test.expected)
			}
		})
	}
}
