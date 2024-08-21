package pkce

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

// InvalidGrantError is returned when the provided authorization grant
// (e.g., authorization, code, resource owner credentials) or refresh token is
// invalid, expired, revoked, does not match the redirection
// URI used in the authorization request, or was issued to
// another client.
// Ref. RFC6749 Section 5.2
var InvalidGrantError = errors.New("invalid_grant")

const (
	// CodeVerifierMinimumLength in bytes
	CodeVerifierMinimumLength = 43
	// CodeVerifierMaximumLength in bytes
	CodeVerifierMaximumLength = 128
	// S256Transformation uses SHA256 as the hashing method
	S256Transformation = "S256"
	// PlainTransformation identifies no hashing method being used
	// Do not use plain in normal operation.
	// FROM RFC7636 Section 4.2:
	//    If the client is capable of using "S256", it MUST use "S256", as
	//    "S256" is Mandatory To Implement (MTI) on the server.  Clients are
	//    permitted to use "plain" only if they cannot support "S256" for some
	//    technical reason and know via out-of-band configuration that the
	//    server supports "plain".

	//    The plain transformation is for compatibility with existing
	//    deployments and for constrained environments that can't use the S256
	//    transformation.
	PlainTransformation = "plain"
)

// GenerateCodeVerifier generates a high-entropy cryptographic
// random STRING using the unreserved characters
//
//	[A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
//
// from Section 2.3 of [RFC3986], with a minimum length of 43 characters
// and a maximum length of 128 characters.
func GenerateCodeVerifier(length int) (string, error) {
	// NOTE: The code verifier SHOULD have enough entropy to make it
	// impractical to guess the value. It is RECOMMENDED that the output of
	// a suitable random number generator be used to create a 32-octet
	// sequence.  The octet sequence is then base64url-encoded to produce a
	// 43-octet URL safe string to use as the code verifier.
	if length < CodeVerifierMinimumLength || length > CodeVerifierMaximumLength {
		return "", errors.New("code verifier length must be between 43 and 128 characters")
	}
	// Base64 => we need 6 bits to represent all 64 characters (2^6=64)
	// So each byte has 2 bits 'extra' so we can fit 4 characters in 3 bytes
	// So the resultant number of bytes required is 4/3 * n = length
	// n = 3 * length / 4
	var (
		// Since the inverse of this is not surjective on [43, 128]
		// we need to take the ceiling so we always have at least one
		// extra byte that we can truncate.
		// using ceil(x/y) = (x+y-1)/y when the cast is a floor
		sequenceLength = (3*length + 4 - 1) / 4
		octetSequence  = make([]byte, sequenceLength)
		n              int
		err            error
	)
	n, err = rand.Reader.Read(octetSequence)
	if err != nil {
		return "", err
	}
	if n < len(octetSequence) {
		return "", errors.New("failed to read enough random bytes")
	}
	return base64.RawURLEncoding.EncodeToString(octetSequence)[:length], nil
}

// GenerateCodeChallenge generates a code challenge from a code
// verifier using the specified transformation.
func GenerateCodeChallenge(transformation,
	codeVerifier string) (string, error) {
	if transformation != S256Transformation && transformation != PlainTransformation {
		return "", errors.New("transformation must be 'S256' or 'plain'")
	}
	if !codeVerifierIsValid(codeVerifier) {
		return "", errors.New("code verifier is not valid see: RFC7636 Section 4.1")
	}
	var (
		codeChallenge = codeVerifier
	)
	if transformation == S256Transformation {
		codeChallenge = s256(codeVerifier)
	}
	return codeChallenge, nil
}

// VerifyCodeVerifier returns an error if the codeVerifier does not
// equal the codeChallenge when transformed using the specified
// method. This function only returns error if the verification fails.
// If the values are not equal or the verifier is not valid, an error
// response indicating "invalid_grant" as described in Section 5.2 of
// [RFC6749] is returned.
func VerifyCodeVerifier(
	transformation, codeVerifier, codeChallenge string) error {
	if transformation != S256Transformation && transformation != PlainTransformation {
		// this error code text is dictated by RFC6749 Section 5.2
		return InvalidGrantError
	}
	if !codeVerifierIsValid(codeVerifier) {
		return InvalidGrantError
	}
	var (
		challenge = codeVerifier
	)
	if transformation == S256Transformation {
		challenge = s256(codeVerifier)
	}
	if x := subtle.ConstantTimeCompare(
		[]byte(challenge),
		[]byte(codeChallenge),
	); x != 1 {
		return InvalidGrantError
	}
	return nil
}
