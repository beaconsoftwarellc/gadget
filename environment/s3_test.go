package environment

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
)

func TestBucketName(t *testing.T) {
	assert := assert1.New(t)

	bucket := NewBucket("env", "proj")
	project := "foo"
	key := "bar"

	value, ok := bucket.Has(project, key)
	assert.Nil(value)
	assert.False(ok)

	items := make(map[string]interface{})
	bucket.Add(project, items)
	value, ok = bucket.Has(project, key)
	assert.Nil(value)
	assert.False(ok)

	expected := "good"
	items[key] = expected
	bucket.Add(project, items)
	value, ok = bucket.Has(project, key)
	assert.Equal(expected, value)
	assert.True(ok)

	value = bucket.Get(project, key, log.NewStackLogger())
	assert.Equal(expected, value)
}
