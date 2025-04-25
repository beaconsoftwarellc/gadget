package record

import (
	"testing"

	_assert "github.com/stretchr/testify/assert"
)

func TestNewLimitOffset(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)
	lo := NewLimitOffset[int]()
	assert.Equal(defaultOffset, lo.Offset())
	assert.Equal(defaultLimit, lo.Limit())
}

func TestLimitOffset_SetLimit(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit uint
	}{
		{
			name:          "positive limit",
			inputLimit:    100,
			expectedLimit: 100,
		},
		{
			name:          "zero limit",
			inputLimit:    0,
			expectedLimit: 0,
		},
		{
			name:          "negative limit",
			inputLimit:    -50,
			expectedLimit: defaultLimit,
		},
		{
			name:          "large limit",
			inputLimit:    999999,
			expectedLimit: 999999,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lo := NewLimitOffset[int]()
			result := lo.SetLimit(tc.inputLimit)
			assert.Equal(tc.expectedLimit, result.Limit())
			assert.Equal(lo, result, "should return self for chaining")
		})
	}
}

func TestLimitOffset_SetOffset(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	tests := []struct {
		name           string
		inputOffset    int
		expectedOffset uint
	}{
		{
			name:           "positive offset",
			inputOffset:    50,
			expectedOffset: 50,
		},
		{
			name:           "zero offset",
			inputOffset:    0,
			expectedOffset: 0,
		},
		{
			name:           "negative offset",
			inputOffset:    -100,
			expectedOffset: 0,
		},
		{
			name:           "large offset",
			inputOffset:    1000000,
			expectedOffset: 1000000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lo := NewLimitOffset[int]()
			result := lo.SetOffset(tc.inputOffset)
			assert.Equal(tc.expectedOffset, result.Offset())
			assert.Equal(lo, result, "should return self for chaining")
		})
	}
}
