package generator

import "crypto/rand"

// Bytes array of length n with random values.
func Bytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
