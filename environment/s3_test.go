package environment

import (
	"context"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
)

func TestBucketName(t *testing.T) {
	assert := assert1.New(t)

	bucket := NewBucket(
		context.Background(),
		"region",
		"bucket",
		"env",
		"proj",
		log.Global(),
	)
	project := "foo"
	key := "bar"

	value, ok := bucket.Get(project, key)
	assert.Nil(value)
	assert.False(ok)

	items := make(map[string]interface{})
	bucket.Add(project, items)
	value, ok = bucket.Get(project, key)
	assert.Nil(value)
	assert.False(ok)

	expected := "good"
	items[key] = expected
	bucket.Add(project, items)
	value, ok = bucket.Get(project, key)
	assert.Equal(expected, value)
	assert.True(ok)

	value, ok = bucket.Get(project, key)
	assert.True(ok)
	assert.Equal(expected, value)
}
