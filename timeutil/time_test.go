package timeutil

import (
	"testing"
	"time"

	_require "github.com/stretchr/testify/require"
)

func Test_Future(t *testing.T) {
	assert := _require.New(t)
	assert.True(Future(10 * time.Second).After(time.Now().UTC()))
}

func Test_Past(t *testing.T) {
	assert := _require.New(t)
	assert.True(Past(10 * time.Second).Before(time.Now().UTC()))
}

func Test_Today_and_Date(t *testing.T) {
	assert := _require.New(t)
	assert.Equal(Today(), Date(time.Now()))
}
