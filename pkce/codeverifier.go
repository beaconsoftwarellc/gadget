package pkce

import (
	"crypto/sha256"
	"encoding/base64"
)

// SEE: RFC7636 Section 4.1
func codeVerifierIsValid(codeVerifier string) bool {
	if len(codeVerifier) < CodeVerifierMinimumLength ||
		len(codeVerifier) > CodeVerifierMaximumLength {
		return false
	}
	for _, r := range codeVerifier {
		if (r < 'A' || r > 'Z') && // A-Z
			(r < 'a' || r > 'z') && // a-z
			(r < '0' || r > '9') && // 0-9
			r != '-' && r != '.' && r != '_' && r != '~' {
			return false
		}
	}
	return true
}

// SEE: RFC7636 Section 4.2
func s256(codeVerifier string) string {
	//  S256
	// 	code_challenge = BASE64URL-ENCODE(SHA256(ASCII(code_verifier)))
	var (
		hash   = sha256.Sum256([]byte(codeVerifier))
		buffer = make([]byte, len(hash))
	)
	copy(buffer, hash[:])
	return base64.RawURLEncoding.EncodeToString(buffer)
}
