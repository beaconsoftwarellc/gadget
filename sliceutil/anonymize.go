package sliceutil

func Anonymize[T any](src []T) []interface{} {
	dst := make([]interface{}, len(src))

	for i, s := range src {
		dst[i] = s
	}

	return dst
}
