package environment

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
)

func Test_ssm_Cache(t *testing.T) {
	assert := assert1.New(t)

	ssm := NewSSM("env")
	key := "%s-foo"

	value, ok := ssm.Has(key)
	assert.Empty(value)
	assert.False(ok)

	expected := "bar"
	ssm.Add("env-foo", expected)
	value, ok = ssm.Has(key)
	assert.Equal(expected, value)
	assert.True(ok)

	value = ssm.Get(key, log.NewStackLogger())
	assert.Equal(expected, value)
}
