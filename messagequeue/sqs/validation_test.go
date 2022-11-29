package sqs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/stretchr/testify/assert"
)

func Test_NameIsValid(t *testing.T) {

	tcs := []struct {
		name  string
		input string
		err   error
	}{
		{
			name:  "invalid character",
			input: "foo$",
			err:   errors.New("name has invalid characters"),
		},
		{
			name:  "invalid utf-8 character",
			input: "fóó",
			err:   errors.New("name has invalid characters"),
		},
		{
			name:  "start with period",
			input: ".foo",
			err:   errors.New("name cannot begin, end, or contain sequences of '.'"),
		},
		{
			name:  "ends with period",
			input: "foo.",
			err:   errors.New("name cannot begin, end, or contain sequences of '.'"),
		},
		{
			name:  "period sequence",
			input: "foo..bar",
			err:   errors.New("name cannot begin, end, or contain sequences of '.'"),
		},
		{
			name:  "invalid prefix aws",
			input: "aws.foo",
			err:   errors.New("name has invalid prefix (amazon|aws)"),
		},
		{
			name:  "invalid prefix aws case insensitive",
			input: "AwS.foo",
			err:   errors.New("name has invalid prefix (amazon|aws)"),
		},
		{
			name:  "invalid prefix amazon",
			input: "amazon.foo",
			err:   errors.New("name has invalid prefix (amazon|aws)"),
		},
		{
			name:  "invalid prefix amazon case insensitive",
			input: "amAzon.foo",
			err:   errors.New("name has invalid prefix (amazon|aws)"),
		},
		{
			name:  "valid",
			input: "foo.bar",
			err:   nil,
		},
		{
			name:  "valid",
			input: "foo.bar-baz",
			err:   nil,
		},
		{
			name:  "valid",
			input: "foo.bar--baz_v_9.11",
			err:   nil,
		},
		{
			name:  "too short",
			input: "",
			err:   errors.New("name character count out of bounds [1, 256] (0)"),
		},
		{
			name:  "too long",
			input: generator.String(257),
			err:   errors.New("name character count out of bounds [1, 256] (257)"),
		},
		{
			name:  "bad prefix 1",
			input: fmt.Sprintf(".%s", generator.String(20)),
			err:   dotError,
		},
		{
			name:  "bad prefix 2",
			input: fmt.Sprintf("%s%s", prohibitedAWS, generator.String(20)),
			err:   prohibitedPrefixError,
		},
		{
			name: "bad prefix 3",
			input: fmt.Sprintf("%s%s", strings.ToUpper(prohibitedAWS),
				generator.String(20)),
			err: prohibitedPrefixError,
		},
		{
			name: "bad prefix 4",
			input: fmt.Sprintf("%s%s", strings.ToUpper(prohibitedAmazon),
				generator.String(20)),
			err: prohibitedPrefixError,
		},
		{
			name:  "bad prefix 5",
			input: fmt.Sprintf("%s%s", prohibitedAmazon, generator.String(20)),
			err:   prohibitedPrefixError,
		},
		{
			name:  "bad suffix",
			input: fmt.Sprintf("%s%s", generator.String(20), "."),
			err:   dotError,
		},
		{
			name:  "bad sequence",
			input: fmt.Sprintf("%s..%s", generator.String(20), generator.String(20)),
			err:   dotError,
		},
		{
			name:  "min",
			input: generator.String(1),
			err:   nil,
		},
		{
			name:  "max",
			input: generator.String(256),
			err:   nil,
		},
		{
			name:  "typical",
			input: generator.String(32),
			err:   nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := NameIsValid(tc.input)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBodyIsValid(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		err   error
	}{
		{
			name:  "below minimum char count",
			input: "",
			err:   bodyMinimumError,
		},
		{
			name:  "above maximum byte count",
			input: generator.String((maxBodyKilobytes+1)*1024 + 1),
			err:   errors.New("body cannot exceed 255 kilobytes (was 256 kb)"),
		},
		{
			name:  "typical",
			input: generator.String(32),
			err:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			err := BodyIsValid(test.input)
			if test.err != nil {
				assert.EqualError(err, test.err.Error())
			} else {
				assert.NoError(err)
			}
		})
	}
}
