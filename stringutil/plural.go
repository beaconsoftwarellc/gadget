package stringutil

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/collection"
)

var (
	vowels     = collection.NewSet("a", "e", "i", "o", "u")
	consonants = collection.NewSet("b", "c", "d", "f", "g", "h", "j", "k", "l", "m", "n", "p", "q", "r", "s", "t", "v", "w", "x", "y", "z")
	esSuffix   = collection.NewSet("s", "ss", "sh", "ch", "x", "z")
)

// Plural of the passed english noun.
// This is a naive approach assuming a non-loan word, well-behaved noun. Words must be > 2 in length.
func Plural(s string) string {
	if len(s) < 3 {
		return s
	}
	var (
		last       = strings.ToLower(string(s[len(s)-1]))
		secondLast = strings.ToLower(string(s[len(s)-2]))
	)
	if last == "z" && vowels.Contains(secondLast) {
		return fmt.Sprintf("%szes", s)
	}
	if esSuffix.Contains(last) || esSuffix.Contains(secondLast+last) {
		return fmt.Sprintf("%ses", s)
	}
	if consonants.Contains(secondLast) && last == "y" {
		return fmt.Sprintf("%sies", s[:len(s)-1])
	}
	if consonants.Contains(secondLast) && last == "o" {
		return fmt.Sprintf("%ses", s)
	}
	if last == "y" {
		return fmt.Sprintf("%ss", s)
	}
	if last == "f" {
		return fmt.Sprintf("%sves", s[:len(s)-1])
	}
	if secondLast+last == "fe" {
		return fmt.Sprintf("%sves", s[:len(s)-2])
	}
	// we could:
	//		- include a static list of words that are their own plural
	// 		- include a list of words that end in 'f' or 'ff' and are pluralized to 'fs' or 'ffs'
	// 		- loan words that violate these rules
	// for now just assume 's'
	return fmt.Sprintf("%ss", s)
}
