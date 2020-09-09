package stringutil

import (
	"strings"
)

// TagOptions is the string following a comma in a struct field's
// tag, or the empty string. It does not include the leading comma.
type TagOptions []string

// ParseTag splits a struct field's tag into its name and
// comma-separated options.
func ParseTag(tag string) (string, TagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], parseOptions(tag[idx+1:])
	}
	return tag, make(TagOptions, 0)
}

func parseOptions(options string) TagOptions {
	optionList := strings.Split(options, ",")
	return CleanWhiteSpace(optionList)
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o TagOptions) Contains(optionName string) bool {
	for _, val := range o {
		if val == optionName {
			return true
		}
	}
	return false
}
