package database

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/log"
)

func TestBootstrapError(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()
	logger := log.NewStackLogger()

	lines := []string{"this isn't really sql"}
	Bootstrap(spec.DB, lines, logger)
	message, err := logger.Pop()
	assert.NoError(err)
	assert.Contains(message, "Error bootstrapping")
}

func TestBootstrapSuccess(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()
	logger := log.NewStackLogger()

	lines := []string{"select * from `test_record`"}
	Bootstrap(spec.DB, lines, logger)
	assert.True(logger.IsEmpty())
}

func TestToSQLString(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	bs, ok := NewBootstrapper(spec.DB).(*bootstrapper)
	assert.True(ok)
	assert.Equal("0", bs.toSQLString(float64(0)))
	assert.Equal("'bob'", bs.toSQLString("bob"))
	assert.Equal("true", bs.toSQLString(true))
}
