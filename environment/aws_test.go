package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAWSMetaServiceLookup(t *testing.T) {
	assert.True(t, isAWSMetaServiceLookup(awsMetaService+"foo"))
	assert.False(t, isAWSMetaServiceLookup("foo"))
	assert.False(t, isAWSMetaServiceLookup("foo"+awsMetaService))
}

func TestAWSLookup(t *testing.T) {
	assert := assert.New(t)
	expected := "foo"
	assert.Equal(expected, AWSLookup(expected))
	expected = awsMetaService + "foo"
	assert.Equal(expected, AWSLookup(expected))
	expected = awsMetaService + "amazon.com"
	assert.NotEqual(expected, AWSLookup(expected))
}
