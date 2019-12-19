package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTag(t *testing.T) {
	assert := assert.New(t)

	tags := []struct {
		Raw     string
		Tag     string
		Options TagOptions
	}{
		{"foo", "foo", TagOptions{}},
		{"foo,", "foo", TagOptions{}},
		{",", "", TagOptions{}},
		{"foo,bar", "foo", TagOptions{"bar"}},
		{"foo,bar,happy", "foo", TagOptions{"bar", "happy"}},
	}

	for _, t := range tags {
		tag, options := ParseTag(t.Raw)

		assert.Equal(t.Tag, tag)
		assert.Equal(t.Options, options)
	}
}

func TestContains(t *testing.T) {
	assert := assert.New(t)

	_, options := ParseTag(",foo,bar")

	assert.True(options.Contains("foo"))
	assert.True(options.Contains("bar"))
	assert.False(options.Contains("happy"))
	assert.False(options.Contains(""))

	options = TagOptions{}
	assert.False(options.Contains(""))
	assert.False(options.Contains("foo"))
}
