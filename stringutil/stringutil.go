// Package stringutil contains utility functions for working with strings.
package stringutil

import (
	"crypto/subtle"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/beaconsoftwarellc/gadget/intutil"
	"github.com/beaconsoftwarellc/gadget/log"
)

// AnonymizeRunes converts an array of runes to an array of anonymous interfaces
func AnonymizeRunes(arr []rune) []interface{} {
	ia := make([]interface{}, len(arr))
	for i, s := range arr {
		ia[i] = s
	}
	return ia
}

// Reverse returns its argument string reversed rune-wise left to right.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// RuneAtIndex returns the rune located at the specified index in the passed
// string. Supports negative indexing.
func RuneAtIndex(s string, i int) rune {
	var runed = []rune(s)
	if i < 0 {
		i = len(runed) + i
	}
	if i < 0 || i > len(runed) {
		var r rune
		return r
	}
	return runed[i]
}

// LastRune returns the last rune in a string. If the string is Empty
// the default value for rune is returned.
func LastRune(s string) rune {
	return RuneAtIndex(s, -1)
}

// IsEmpty returns a bool indicating that the passed string is Empty. used
// primarily as a filter function.
func IsEmpty(s string) bool {
	return len(s) == 0
}

// IsWhiteSpace returns a bool indicating whether the passed string is
// composed entirely of whitespace characters.
func IsWhiteSpace(s string) bool {
	// TrimSpace is all whitespace
	return IsEmpty(strings.TrimSpace(s))
}

// CleanWhiteSpace removes all strings in the passed slice that contain only
// white space.
func CleanWhiteSpace(s []string) []string {
	return Filter(s, IsWhiteSpace)
}

// Clean removes empty strings from a slice of strings and returns the new
// slice.
func Clean(s []string) []string {
	return Filter(s, IsEmpty)
}

// Filter removes all strings from a slice of strings that do not have a
// value of 'true' when passed to the filter function.
func Filter(ss []string, filter func(string) bool) []string {
	// This is fast and suitable for 'small' slices since we will
	// be duplicating the data in memory.
	var filtered = make([]string, 0)
	for _, s := range ss {
		if !filter(s) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// SafeSubstring returns a substring that should be safe to use with strings
// that contain non-ascii characters, with python style indexing. If end equals
// 0 it will be interpreted as the end of the string.
func SafeSubstring(value string, start int, end int) string {
	if len(value) == 0 {
		return ""
	}
	runed := []rune(value)
	if start < 0 {
		start = len(runed) + start
	}
	start = intutil.Max(start, 0)
	if end <= 0 {
		end = len(runed) + end
	}

	if end < start {
		swap := end
		end = start
		start = swap
	}
	start = intutil.Max(start, 0)
	end = intutil.Min(end, len(runed))
	if start == end {
		return ""
	}
	return string(runed[start:end])
}

// SprintHex the passed byte array as a string of Hexadecimal numbers space
// separated.
func SprintHex(b []byte) string {
	return fmt.Sprintf("% X", b)
}

// ByteToHexASCII returns a byte slice containing the hex representation of the
// passed byte array in ASCII characters.
func ByteToHexASCII(b []byte) []byte {
	r := []byte{}
	for _, b := range b {
		for _, c := range fmt.Sprintf("%.2X", b) {
			r = append(r, byte(c))
		}
	}
	return r
}

// MakeASCIIZeros in a byte array of the passed size.
func MakeASCIIZeros(count uint) []byte {
	b := make([]byte, count)
	for i := 0; i < int(count); i++ {
		b[i] = '0'
	}
	return b
}

// NullTerminatedString from the passed byte array.
// Note: this only works with ASCII or UTF-8
func NullTerminatedString(b []byte) string {
	var i int
	n := rune(0)
	r := []rune(string(b))
	for i = 0; i < len(r); i++ {
		if r[i] == n {
			break
		}
	}
	return string(b[:i])
}

// AppendIfMissing adds a string to a slice if it's not already in it
func AppendIfMissing(slice []string, i string) []string {
	if Contains(slice, i) {
		return slice
	}
	return append(slice, i)
}

// Contains checks if a string is in a slice
func Contains(slice []string, i string) bool {
	for _, ele := range slice {
		if ele == i {
			return true
		}
	}
	return false
}

// ConstantTimeComparison evaluates strings in a constant time to avoid hack attempts based on string comparison
// response rates
func ConstantTimeComparison(expected, actual string) bool {
	equal := len(expected) == len(actual)
	minLength := intutil.Min(len(expected), len(actual))
	for i := 0; i < minLength; i++ {
		check := subtle.ConstantTimeByteEq(byte(expected[i]), byte(actual[i])) == 1
		equal = equal && check
	}
	return equal
}

// GetSchemaName from an object for use in the schema field on a routed message payload.
func GetSchemaName(obj proto.Message) (name string) {
	if nil == obj {
		log.Infof("GetSchemaName called with nil")
		return "<nil>"
	}
	ptrtype := reflect.TypeOf(obj)
	// we want the actual type not '*type'.
	name = ptrtype.Elem().String()
	return
}

// Anonymize converts an array of strings to an array of anonymous interfaces
func Anonymize(arr []string) []interface{} {
	ia := make([]interface{}, len(arr))
	for i, s := range arr {
		ia[i] = s
	}
	return ia
}

// PSPrint pretty prints the passed map to string.
func PSPrint(prefix string, m map[string]string) string {
	sa := make([]string, len(m))
	i := 0
	for k, v := range m {
		sa[i] = fmt.Sprintf("%s'%s': '%s'", prefix, k, v)
		i++
	}
	return strings.Join(sa, "\n")
}

// Pointer converts a string to a string pointer
func Pointer(str string) *string {
	return &str
}
