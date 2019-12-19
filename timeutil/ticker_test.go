package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTicker(t *testing.T) {
	assert := assert.New(t)
	expected := 1 * time.Second
	obj := NewTicker(expected)
	assert.NotNil(obj)
	tkr, ok := obj.(*ticker)
	assert.True(ok)
	assert.Equal(expected, tkr.period)
	assert.Nil(tkr.ticker)
}

func TestTicker_StartStop(t *testing.T) {
	assert := assert.New(t)
	obj := NewTicker(1 * time.Millisecond)
	assert.NotNil(obj)
	tkr, ok := obj.(*ticker)
	assert.True(ok)
	assert.Nil(tkr.ticker)
	tkr.Start()
	assert.NotNil(tkr.ticker)
	time.Sleep(2 * time.Millisecond)
	select {
	case <-tkr.Channel():
	default:
		assert.Fail("should have time in channel")
	}
	tkr.Stop()
	assert.Nil(tkr.ticker)
}

func TestTicker_Reset(t *testing.T) {
	assert := assert.New(t)
	obj := NewTicker(1 * time.Millisecond)
	assert.NotNil(obj)
	tkr, ok := obj.(*ticker)
	assert.True(ok)
	tkr.Reset()
	time.Sleep(2 * time.Millisecond)
	select {
	case <-tkr.Channel():
	default:
		assert.Fail("should have time in channel")
	}
	tkr.Reset()
	time.Sleep(2 * time.Millisecond)
	select {
	case <-tkr.Channel():
	default:
		assert.Fail("should have time in channel")
	}
}
