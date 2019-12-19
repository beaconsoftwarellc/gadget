package generator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDFunc(t *testing.T) {
	assert := assert.New(t)

	var testData = []IDPrefix{
		"",
		"f",
		"fo",
		"foo",
		"!adsf",
	}

	for _, prefix := range testData {
		id := ID(prefix)
		prefixLength := len(prefix) + 1
		assert.Len(id, IDSizeBytes)
		assert.Equal(fmt.Sprintf("%s_", prefix), id[:prefixLength])
	}

	var expected []string
	for i := 0; i < 1000; i++ {
		actual := ID(IDPrefix("test"))
		assert.NotContains(expected, actual)
		expected = append(expected, actual)
	}

	assert.Panics(func() { ID(IDPrefix(random(MaxPrefix+1, letters))) })
}
func TestBase32IDFunc(t *testing.T) {
	assert := assert.New(t)

	var testData = []IDPrefix{
		"",
		"f",
		"fo",
		"foo",
		"!adsf",
	}

	for _, prefix := range testData {
		id := ID(prefix)
		prefixLength := len(prefix) + 1
		assert.Len(id, IDSizeBytes)
		assert.Equal(fmt.Sprintf("%s_", prefix), id[:prefixLength])
	}

	var expected []string
	for i := 0; i < 1000; i++ {
		actual := Base32ID(IDPrefix("test"))
		assert.NotContains(expected, actual)
		expected = append(expected, actual)
	}

	assert.Panics(func() { ID(IDPrefix(random(MaxPrefix+1, letters))) })
}

func TestEmail(t *testing.T) {
	assert := assert.New(t)

	var expected []string
	for i := 0; i < 1000; i++ {
		actual := Email()
		assert.NotContains(expected, actual)
		assert.Contains(actual, "@")
		expected = append(expected, actual)
	}
}

func TestPassword(t *testing.T) {
	assert := assert.New(t)

	var expected []string
	for i := 0; i < 1000; i++ {
		actual := Password(16)
		assert.NotContains(expected, actual)
		assert.Len(actual, 16)
		expected = append(expected, actual)
	}
}

func TestSecret(t *testing.T) {
	assert := assert.New(t)

	var expected []string
	for i := 0; i < 1000; i++ {
		actual := Secret()
		assert.NotContains(expected, actual)
		assert.Len(actual, secretLength)
		expected = append(expected, actual)
	}
}

func TestRandom(t *testing.T) {
	assert := assert.New(t)

	var testData = []int{10, 15, 20}

	var expected []string
	for _, length := range testData {
		for i := 0; i < 100; i++ {
			actual := Password(length)
			assert.NotContains(expected, actual)
			assert.Len(actual, length)
			expected = append(expected, actual)
		}
	}
}

func TestName(t *testing.T) {
	assert := assert.New(t)

	var expected []string
	for i := 0; i < 100; i++ {
		actual := Name()
		assert.NotContains(expected, actual)
		assert.Len(actual, 19)
		expected = append(expected, actual)
	}
}

func TestString(t *testing.T) {
	assert := assert.New(t)

	var expected []string
	for i := 0; i < 100; i++ {
		length := (i % 13) + 10
		actual := String(length)
		assert.NotContains(expected, actual)
		assert.Len(actual, length)
		expected = append(expected, actual)
	}
}

func TestCode(t *testing.T) {
	assert := assert.New(t)

	var expected []string
	for i := 0; i < 100; i++ {
		length := (i % 13) + 10
		actual := Code(length)
		assert.NotContains(expected, actual)
		assert.Len(actual, length)
		expected = append(expected, actual)
	}
}

func TestHex(t *testing.T) {
	assert := assert.New(t)

	var expected []string
	for i := 0; i < 100; i++ {
		length := (i % 13) + 10
		actual := Hex(length)
		assert.NotContains(expected, actual)
		assert.Len(actual, length)
		expected = append(expected, actual)
	}
}
