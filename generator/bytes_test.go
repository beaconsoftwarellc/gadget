package generator

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func TestBytes(t *testing.T) {
	assert := assert1.New(t)
	b1 := Bytes(10)
	b2 := Bytes(10)
	assert.NotEqual(b1, b2)
}
