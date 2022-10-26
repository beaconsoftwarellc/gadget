package sliceutil

// Anonymize casts slice to slice of interfaces
func Anonymize[T any](src []T) []interface{} {
	dst := make([]interface{}, len(src))

	for i, s := range src {
		dst[i] = s
	}

	return dst
}

// ToStringSlice casts slice of string type aliases to slice of string type
func ToStringSlice[T ~string](src []T) []string {
	dst := make([]string, len(src))

	for i, s := range src {
		dst[i] = string(s)
	}

	return dst
}
