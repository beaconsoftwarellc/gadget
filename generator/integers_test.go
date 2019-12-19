package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt16(t *testing.T) {
	assert := assert.New(t)
	rand1 := Int16()
	rand2 := Int16()
	assert.NotEqual(rand1, rand2)
}

func TestUInt16(t *testing.T) {
	assert := assert.New(t)
	rand1 := UInt16()
	rand2 := UInt16()
	assert.NotEqual(rand1, rand2)
}
