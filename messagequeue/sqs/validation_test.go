package sqs

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/stretchr/testify/assert"
)

func Test_Validation(t *testing.T) {

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
			name:  "too short",
			input: "",
			err:   errors.New("name character count out of bounds [1, 256] (0)"),
		},
		{
			name:  "too long",
			input: generator.String(300),
			err:   errors.New("name character count out of bounds [1, 256] (300)"),
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
