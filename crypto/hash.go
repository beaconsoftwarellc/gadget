package crypto

import (
	"crypto/sha512"
	"encoding/hex"

	"github.com/google/uuid"
)

// HashAndSalt generates a Hash and a Salt for a given string
func HashAndSalt(password string) (hash string, salt string) {
	salt = uuid.New().String()
	hash = Hash(password, salt)
	return hash, salt
}

// Hash returns the hash of a given string and salt
func Hash(password string, salt string) string {
	b := sha512.Sum512([]byte(password + salt))
	return hex.EncodeToString(b[:])
}
