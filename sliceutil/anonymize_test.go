package sliceutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Anonymize(t *testing.T) {

	type testCase[T any] struct {
		name   string
		input  []T
		expect []interface{}
	}

	t.Run("ints", func(t *testing.T) {
		tcs := []testCase[int]{
			{
				name:   "base",
				input:  []int{1, 2, 3},
				expect: []interface{}{1, 2, 3},
			},
			{
				name:   "empty",
				input:  []int{},
				expect: []interface{}{},
			},
			{
				name:   "nil",
				input:  nil,
				expect: []interface{}{},
			},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				actual := Anonymize(tc.input)
				assert.Equal(t, tc.expect, actual)
			})
		}
	})

	t.Run("strings", func(t *testing.T) {
		tcs := []testCase[string]{
			{
				name:   "base",
				input:  []string{"a", "b", "c"},
				expect: []interface{}{"a", "b", "c"},
			},
			{
				name:   "empty",
				input:  []string{},
				expect: []interface{}{},
			},
			{
				name:   "nil",
				input:  nil,
				expect: []interface{}{},
			},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				actual := Anonymize(tc.input)
				assert.Equal(t, tc.expect, actual)
			})
		}
	})
}

func Test_String(t *testing.T) {

	type testCase[T any] struct {
		name   string
		input  []T
		expect []string
	}

	type foo string

	t.Run("ints", func(t *testing.T) {
		tcs := []testCase[foo]{
			{
				name:   "base",
				input:  []foo{"a", "b", "c"},
				expect: []string{"a", "b", "c"},
			},
			{
				name:   "empty",
				input:  []foo{},
				expect: []string{},
			},
			{
				name:   "nil",
				input:  nil,
				expect: []string{},
			},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				actual := String(tc.input)
				assert.Equal(t, tc.expect, actual)
			})
		}
	})

	t.Run("strings", func(t *testing.T) {
		tcs := []testCase[string]{
			{
				name:   "base",
				input:  []string{"a", "b", "c"},
				expect: []string{"a", "b", "c"},
			},
			{
				name:   "empty",
				input:  []string{},
				expect: []string{},
			},
			{
				name:   "nil",
				input:  nil,
				expect: []string{},
			},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				actual := String(tc.input)
				assert.Equal(t, tc.expect, actual)
			})
		}
	})
}
