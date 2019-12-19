package crypto

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

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

// HashMD5 returns the MD5 sum of the passed bytes
func HashMD5(plaintext []byte) string {
	return fmt.Sprintf("%x", md5.Sum(plaintext))
}