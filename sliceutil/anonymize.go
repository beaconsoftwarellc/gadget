package sliceutil

func Anonymize[T any](src []T) []interface{} {
	dst := make([]interface{}, len(src))

	for i, s := range src {
		dst[i] = s
	}

	return dst
}

func String[T ~string](src []T) []string {
	dst := make([]string, len(src))

	for i, s := range src {
		dst[i] = string(s)
	}

	return dst
}
