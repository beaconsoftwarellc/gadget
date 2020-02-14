package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDockerSocket(t *testing.T) {
	assert := assert.New(t)
	assert.True(IsDockerSocket(dockerSocket + "foo"))
	assert.False(IsDockerSocket("foo"))
}

func TestLookupHostPort(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(8080, LookupHostPort(80, "foo", 8080))
}
