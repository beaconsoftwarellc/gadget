package net

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/generator"
)

func TestValidateIPv4Address(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "IPv4 returns true",
			args: args{"127.0.0.1"},
			want: true,
		},
		{
			name: "IPv4 returns true",
			args: args{"255.255.255.255"},
			want: true,
		},
		{
			name: "Hostname returns false",
			args: args{"localhost"},
			want: false,
		},
		{
			name: "out of bounds returns false",
			args: args{"256.1.1.1"},
			want: false,
		},
		{
			name: "letters are right out",
			args: args{"256.1.1.b"},
			want: false,
		},
		{
			name: "with port should succeed",
			args: args{"255.1.1.2:80"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateIPv4Address(tt.args.s); got != tt.want {
				t.Errorf("ValidateIPv4Address() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateIPv6Address(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "omission is good",
			arg:  "2001:470:9b36:1::2",
			want: true,
		},
		{
			name: "not omission is good as well",
			arg:  "2001:cdba:0000:0000:0000:0000:3257:9652",
			want: true,
		},
		{
			name: "no leading 0s is fine",
			arg:  "2001:cdba:0:0:0:0:3257:9652",
			want: true,
		},
		{
			name: "upper and lower is fine",
			arg:  "2001:cdBA::3257:9652",
			want: true,
		},
		{
			name: "localhost is good",
			arg:  "[::1]",
			want: true,
		},
		{
			name: "omission is good with port",
			arg:  "[2001:470:9b36:1::2]:80",
			want: true,
		},
		{
			name: "not omission is good as well with port",
			arg:  "[2001:cdba:0000:0000:0000:0000:3257:9652]:8012",
			want: true,
		},
		{
			name: "no leading 0s is fine with port",
			arg:  "[2001:cdba:0:0:0:0:3257:9652]:65534",
			want: true,
		},
		{
			name: "upper and lower is fine with port",
			arg:  "[2001:cdBA::3257:9652]:10",
			want: true,
		},
		{
			name: "localhost with port is fine too",
			arg:  "[::1]:10",
			want: true,
		},
		{
			name: "two many omissions",
			arg:  "1200:AB00:1234:1:1:1:1:1:2552:7777:1313",
			want: false,
		},
		{
			name: "O is no bueno",
			arg:  "1200:0000:AB00:1234:O000:2552:7777:1313",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateIPv6Address(tt.arg); got != tt.want {
				t.Errorf("ValidateIPv6Address() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIntValue(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def = 30
	actual := GetIntValue(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetIntValue(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetIntValue(values, key, def)
	assert.Equal(int(-1), actual)

	var expected = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetIntValue(values, key, def)
	assert.Equal(expected, actual)
}

func TestGetInt64Value(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def int64 = 30
	actual := GetInt64Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetInt64Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetInt64Value(values, key, def)
	assert.Equal(int64(-1), actual)

	var expected int64 = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetInt64Value(values, key, def)
	assert.Equal(expected, actual)
}

func TestGetInt32Value(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def int32 = 30
	actual := GetInt32Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetInt32Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetInt32Value(values, key, def)
	assert.Equal(int32(-1), actual)

	var expected int32 = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetInt32Value(values, key, def)
	assert.Equal(expected, actual)
}

func TestGetInt16Value(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def int16 = 30
	actual := GetInt16Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetInt16Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetInt16Value(values, key, def)
	assert.Equal(int16(-1), actual)

	var expected int16 = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetInt16Value(values, key, def)
	assert.Equal(expected, actual)
}

func TestGetInt8Value(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def int8 = 30
	actual := GetInt8Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetInt8Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetInt8Value(values, key, def)
	assert.Equal(int8(-1), actual)

	var expected int8 = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetInt8Value(values, key, def)
	assert.Equal(expected, actual)
}

func TestGetUintValue(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def uint = 30
	actual := GetUintValue(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetUintValue(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetUintValue(values, key, def)
	assert.Equal(^uint(0), actual)

	var expected uint = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetUintValue(values, key, def)
	assert.Equal(expected, actual)
}

func TestGetUint64Value(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def uint64 = 30
	actual := GetUint64Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetUint64Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetUint64Value(values, key, def)
	assert.Equal(^uint64(0), actual)

	var expected uint64 = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetUint64Value(values, key, def)
	assert.Equal(expected, actual)
}

func TestGetUint32Value(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def uint32 = 30
	actual := GetUint32Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetUint32Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetUint32Value(values, key, def)
	assert.Equal(^uint32(0), actual)

	var expected uint32 = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetUint32Value(values, key, def)
	assert.Equal(expected, actual)
}

func TestGetUint16Value(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def uint16 = 30
	actual := GetUint16Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetUint16Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetUint16Value(values, key, def)
	assert.Equal(^uint16(0), actual)

	var expected uint16 = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetUint16Value(values, key, def)
	assert.Equal(expected, actual)
}

func TestGetUint8Value(t *testing.T) {
	assert := assert.New(t)

	values := url.Values{}

	key := "test"
	var def uint8 = 30
	actual := GetUint8Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, generator.String(4))
	actual = GetUint8Value(values, key, def)
	assert.Equal(def, actual)

	values.Set(key, "-1")
	actual = GetUint8Value(values, key, def)
	assert.Equal(^uint8(0), actual)

	var expected uint8 = 42
	values.Set(key, strconv.Itoa(int(expected)))
	actual = GetUint8Value(values, key, def)
	assert.Equal(expected, actual)
}
