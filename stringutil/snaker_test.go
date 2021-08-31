package stringutil

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func TestUnderscore(t *testing.T) {
	assert := assert1.New(t)
	inputs := [][]string{
		{"ILoveGoAndJSONSoMuch", "i_love_go_and_json_so_much"},
		{"CamelCase", "camel_case"},
		{"Camel", "camel"},
		{"CAMEL", "camel"},
		{"camel", "camel"},
		{"BIGCase", "big_case"},
		{"SmallCASE", "small_case"},
		{"My fancy alpaca", "my_fancy_alpaca"},
	}
	for _, input := range inputs {
		output := Underscore(input[0])
		assert.Equal(input[1], output)
	}
}

func TestDash(t *testing.T) {
	assert := assert1.New(t)
	inputs := [][]string{
		{"ILoveGoAndJSONSoMuch", "i-love-go-and-json-so-much"},
		{"CamelCase", "camel-case"},
		{"Camel", "camel"},
		{"CAMEL", "camel"},
		{"camel", "camel"},
		{"BIGCase", "big-case"},
		{"My fancy alpaca", "my-fancy-alpaca"},
	}
	for _, input := range inputs {
		output := Dash(input[0])
		assert.Equal(input[1], output)
	}
}

func TestLowerCamelCase(t *testing.T) {
	assert := assert1.New(t)
	inputs := [][]string{
		{"ILoveGoAndJSONSoMuch", "iLoveGoAndJSONSoMuch"},
		{"CamelCase", "camelCase"},
		{"Camel_case", "camelCase"},
		{"camel-case", "camelCase"},
		{"Camel", "camel"},
		{"CAMEL", "camel"},
		{"camel", "camel"},
		{"BIGCase", "bigCase"},
		{"SmallCASE", "smallCASE"},
	}
	for _, input := range inputs {
		output := LowerCamelCase(input[0])
		assert.Equal(input[1], output)
	}
}

func TestUpperCamelCase(t *testing.T) {
	assert := assert1.New(t)
	inputs := [][]string{
		{"ILoveGoAndJSONSoMuch", "ILoveGoAndJSONSoMuch"},
		{"CamelCase", "CamelCase"},
		{"Camel_case", "CamelCase"},
		{"camel-case", "CamelCase"},
		{"Camel", "Camel"},
		{"CAMEL", "CAMEL"},
		{"camel", "Camel"},
		{"BIGCase", "BIGCase"},
		{"SmallCASE", "SmallCASE"},
	}
	for _, input := range inputs {
		output := UpperCamelCase(input[0])
		assert.Equal(input[1], output)
	}
}
