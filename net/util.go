package net

import (
	"math"
	nnet "net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/beaconsoftwarellc/gadget/generator"
	"github.com/beaconsoftwarellc/gadget/log"
	"github.com/beaconsoftwarellc/gadget/stringutil"
)

var (
	ipv4AddressRegEx = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])(:[0-9]{1,5})?\b`)
)

// ValidateIPv4Address with or without a port and return a boolean indicating success.
func ValidateIPv4Address(s string) bool {
	return ipv4AddressRegEx.MatchString(s)
}

func cleanIPv6(s string) (string, int) {
	index := strings.Index(s, "]:")
	port := 0
	var err error
	if index != -1 {
		// remove the trailing ] and port
		portString := stringutil.SafeSubstring(s, index+2, 0)
		port, err = strconv.Atoi(portString)
		if nil != err {
			log.Warn(err)
			return "", -1
		}
		if port <= 0 || port > math.MaxUint16 {
			log.Warnf("port '%d' is not a valid port number, must be uint16", port)
			return "", -1
		}
		s = stringutil.SafeSubstring(s, 1, index)
	}
	if s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}
	return s, port
}

// ValidateIPv6Address with or without a port and return a boolean indicating success.
func ValidateIPv6Address(s string) bool {
	if ValidateIPv4Address(s) {
		return false
	}
	s, _ = cleanIPv6(s)
	return nil != nnet.ParseIP(s)
}

// RandomizeIPArray for basic load balancing
func RandomizeIPArray(a []nnet.IP) []nnet.IP {
	randomized := make([]nnet.IP, len(a))
	// ignore output since we sized the array
	copy(randomized, a)
	// swap indices in the array randomly len times
	for i := 0; i < len(a); i++ {
		j := generator.Int() % len(a)
		swap := randomized[i]
		randomized[i] = randomized[j]
		randomized[j] = swap
	}
	return randomized
}

// GetIntValue attempts to retrieve the value from url.Values, otherwise returns the default
func GetIntValue(values url.Values, key string, def int) int {
	if 64 == strconv.IntSize {
		return int(GetInt64Value(values, key, int64(def)))
	}
	return int(GetInt32Value(values, key, int32(def)))
}

// GetInt64Value attempts to retrieve the value from url.Values, otherwise returns the default
func GetInt64Value(values url.Values, key string, def int64) int64 {
	valueStr, ok := values[key]
	if !ok {
		return def
	}
	v, err := strconv.ParseInt(valueStr[0], 10, 64)
	if err != nil {
		return def // if not a valid int64, return the default
	}
	return int64(v)
}

// GetInt32Value attempts to retrieve the value from url.Values, otherwise returns the default
func GetInt32Value(values url.Values, key string, def int32) int32 {
	valueStr, ok := values[key]
	if !ok {
		return def
	}
	v, err := strconv.ParseInt(valueStr[0], 10, 32)
	if err != nil {
		return def // if not a valid int32, return the default
	}
	return int32(v)
}

// GetInt16Value attempts to retrieve the value from url.Values, otherwise returns the default
func GetInt16Value(values url.Values, key string, def int16) int16 {
	valueStr, ok := values[key]
	if !ok {
		return def
	}
	v, err := strconv.ParseInt(valueStr[0], 10, 16)
	if err != nil {
		return def // if not a valid int16, return the default
	}
	return int16(v)
}

// GetInt8Value attempts to retrieve the value from url.Values, otherwise returns the default
func GetInt8Value(values url.Values, key string, def int8) int8 {
	valueStr, ok := values[key]
	if !ok {
		return def
	}
	v, err := strconv.ParseInt(valueStr[0], 10, 8)
	if err != nil {
		return def // if not a valid int8, return the default
	}
	return int8(v)
}

// GetUintValue attempts to retrieve the value from url.Values, otherwise returns the default
func GetUintValue(values url.Values, key string, def uint) uint {
	if 64 == strconv.IntSize {
		return uint(GetUint64Value(values, key, uint64(def)))
	}
	return uint(GetUint32Value(values, key, uint32(def)))
}

// GetUint64Value attempts to retrieve the value from url.Values, otherwise returns the default
func GetUint64Value(values url.Values, key string, def uint64) uint64 {
	valueStr, ok := values[key]
	if !ok {
		return def
	}
	v, err := strconv.ParseInt(valueStr[0], 10, 64)
	if err != nil {
		return def // if not a valid uint64, return the default
	}
	return uint64(v)
}

// GetUint32Value attempts to retrieve the value from url.Values, otherwise returns the default
func GetUint32Value(values url.Values, key string, def uint32) uint32 {
	valueStr, ok := values[key]
	if !ok {
		return def
	}
	v, err := strconv.ParseInt(valueStr[0], 10, 32)
	if err != nil {
		return def // if not a valid uint32, return the default
	}
	return uint32(v)
}

// GetUint16Value attempts to retrieve the value from url.Values, otherwise returns the default
func GetUint16Value(values url.Values, key string, def uint16) uint16 {
	valueStr, ok := values[key]
	if !ok {
		return def
	}
	v, err := strconv.ParseInt(valueStr[0], 10, 16)
	if err != nil {
		return def // if not a valid uint16, return the default
	}
	return uint16(v)
}

// GetUint8Value attempts to retrieve the value from url.Values, otherwise returns the default
func GetUint8Value(values url.Values, key string, def uint8) uint8 {
	valueStr, ok := values[key]
	if !ok {
		return def
	}
	v, err := strconv.ParseInt(valueStr[0], 10, 8)
	if err != nil {
		return def // if not a valid uint8, return the default
	}
	return uint8(v)
}
