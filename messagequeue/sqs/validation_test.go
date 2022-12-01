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
			err:   errors.New(dotError),
		},
		{
			name:  "bad prefix 2",
			input: fmt.Sprintf("%s%s", prohibitedAWS, generator.String(20)),
			err:   errors.New(prohibitedPrefixError),
		},
		{
			name: "bad prefix 3",
			input: fmt.Sprintf("%s%s", strings.ToUpper(prohibitedAWS),
				generator.String(20)),
			err: errors.New(prohibitedPrefixError),
		},
		{
			name: "bad prefix 4",
			input: fmt.Sprintf("%s%s", strings.ToUpper(prohibitedAmazon),
				generator.String(20)),
			err: errors.New(prohibitedPrefixError),
		},
		{
			name:  "bad prefix 5",
			input: fmt.Sprintf("%s%s", prohibitedAmazon, generator.String(20)),
			err:   errors.New(prohibitedPrefixError),
		},
		{
			name:  "bad suffix",
			input: fmt.Sprintf("%s%s", generator.String(20), "."),
			err:   errors.New(dotError),
		},
		{
			name:  "bad sequence",
			input: fmt.Sprintf("%s..%s", generator.String(20), generator.String(20)),
			err:   errors.New(dotError),
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

func Test_BodyIsValid(t *testing.T) {

	tcs := []struct {
		name  string
		input string
		err   error
	}{
		{
			name:  "empty",
			input: "",
			err:   errors.New("minimum character count is 1"),
		},
		{
			name:  "too long",
			input: generator.String(256*1000 + 1),
			err:   errors.New("body cannot exceed 256 kilobytes (was 256001 bytes)"),
		},
		{
			name:  "null char",
			input: "foo\x00",
			err:   errors.New("char 0x0 is not allowed unicode character"),
		},
		{
			name:  "forbidden utf char",
			input: "foo\x0f",
			err:   errors.New("char 0xf is not allowed unicode character"),
		},
		{
			name:  "ok",
			input: "foo 😁",
			err:   nil,
		},
		{
			name:  "single allowed chars",
			input: "\x09\x0A\x0D",
			err:   nil,
		},
		{
			name:  "allowed ranges",
			input: "\x20 \ud7fe \ud7ff \ue000 \ue001 \ufffd \u10000 \U00010000 \U00010001 \U0010ffff",
			err:   nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := BodyIsValid(tc.input)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
