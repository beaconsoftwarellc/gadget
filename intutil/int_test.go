package intutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt64Max(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    int64
		max      int64
		expected int64
	}{
		{42, 43, 42},
		{42, 42, 42},
		{43, 42, 42},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Int64Max(data.input, data.max))
	}
}

func TestInt64Min(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    int64
		min      int64
		expected int64
	}{
		{42, 43, 43},
		{42, 42, 42},
		{43, 42, 43},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Int64Min(data.input, data.min))
	}
}

func TestInt32Max(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    int32
		max      int32
		expected int32
	}{
		{42, 43, 42},
		{42, 42, 42},
		{43, 42, 42},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Int32Max(data.input, data.max))
	}
}

func TestInt32Min(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    int32
		min      int32
		expected int32
	}{
		{42, 43, 43},
		{42, 42, 42},
		{43, 42, 43},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Int32Min(data.input, data.min))
	}
}

func TestInt16Max(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    int16
		max      int16
		expected int16
	}{
		{42, 43, 42},
		{42, 42, 42},
		{43, 42, 42},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Int16Max(data.input, data.max))
	}
}

func TestInt16Min(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    int16
		min      int16
		expected int16
	}{
		{42, 43, 43},
		{42, 42, 42},
		{43, 42, 43},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Int16Min(data.input, data.min))
	}
}

func TestInt8Max(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    int8
		max      int8
		expected int8
	}{
		{42, 43, 42},
		{42, 42, 42},
		{43, 42, 42},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Int8Max(data.input, data.max))
	}
}

func TestInt8Min(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    int8
		min      int8
		expected int8
	}{
		{42, 43, 43},
		{42, 42, 42},
		{43, 42, 43},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Int8Min(data.input, data.min))
	}
}

func TestUintMax(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint
		max      uint
		expected uint
	}{
		{42, 43, 42},
		{42, 42, 42},
		{43, 42, 42},
	}

	for _, data := range testData {
		assert.Equal(data.expected, UintMax(data.input, data.max))
	}
}

func TestUintMin(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint
		min      uint
		expected uint
	}{
		{42, 43, 43},
		{42, 42, 42},
		{43, 42, 43},
	}

	for _, data := range testData {
		assert.Equal(data.expected, UintMin(data.input, data.min))
	}
}

func TestUint64Max(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint64
		max      uint64
		expected uint64
	}{
		{42, 43, 42},
		{42, 42, 42},
		{43, 42, 42},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Uint64Max(data.input, data.max))
	}
}

func TestUint64Min(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint64
		min      uint64
		expected uint64
	}{
		{42, 43, 43},
		{42, 42, 42},
		{43, 42, 43},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Uint64Min(data.input, data.min))
	}
}

func TestUint32Max(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint32
		max      uint32
		expected uint32
	}{
		{42, 43, 42},
		{42, 42, 42},
		{43, 42, 42},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Uint32Max(data.input, data.max))
	}
}

func TestUint32Min(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint32
		min      uint32
		expected uint32
	}{
		{42, 43, 43},
		{42, 42, 42},
		{43, 42, 43},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Uint32Min(data.input, data.min))
	}
}

func TestUint16Max(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint16
		max      uint16
		expected uint16
	}{
		{42, 43, 42},
		{42, 42, 42},
		{43, 42, 42},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Uint16Max(data.input, data.max))
	}
}

func TestUint16Min(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint16
		min      uint16
		expected uint16
	}{
		{42, 43, 43},
		{42, 42, 42},
		{43, 42, 43},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Uint16Min(data.input, data.min))
	}
}

func TestUint8Max(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint8
		max      uint8
		expected uint8
	}{
		{42, 43, 42},
		{42, 42, 42},
		{43, 42, 42},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Uint8Max(data.input, data.max))
	}
}

func TestUint8Min(t *testing.T) {
	assert := assert.New(t)

	testData := []struct {
		input    uint8
		min      uint8
		expected uint8
	}{
		{42, 43, 43},
		{42, 42, 42},
		{43, 42, 43},
	}

	for _, data := range testData {
		assert.Equal(data.expected, Uint8Min(data.input, data.min))
	}
}
