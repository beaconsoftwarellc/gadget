package environment

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
)

func TestBucketName(t *testing.T) {
	assert := assert1.New(t)

	bucket := NewBucket()
	bucketName := "foo"
	item := "bar"
	key := "bad"

	value, ok := bucket.Has(bucketName, item, key)
	assert.Nil(value)
	assert.False(ok)

	items := make(map[string]interface{})
	bucket.Add(bucketName, item, items)
	value, ok = bucket.Has(bucketName, item, key)
	assert.Nil(value)
	assert.False(ok)

	expected := "good"
	items[key] = expected
	bucket.Add(bucketName, item, items)
	value, ok = bucket.Has(bucketName, item, key)
	assert.Equal(expected, value)
	assert.True(ok)

	value = bucket.Get(bucketName, item, key, log.NewStackLogger())
	assert.Equal(expected, value)
}
