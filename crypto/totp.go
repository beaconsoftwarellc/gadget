package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/skip2/go-qrcode"
)

// NewOTPKey for use with HOTP or TOTP as a base32 encoded string
func NewOTPKey() (string, error) {
	key := make([]byte, sha1.Size)
	random := rand.Reader
	_, err := io.ReadFull(random, key)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(key), err
}

// DynamicTruncate as described in RFC4226
//
//	"The Truncate function performs Step 2 and Step 3, i.e., the dynamic
//	truncation and then the reduction modulo 10^Digit.  The purpose of
//	the dynamic offset truncation technique is to extract a 4-byte
//	dynamic binary code from a 160-bit (20-byte) HMAC-SHA-1 result.
//
//	 DT(String) // String = String[0]...String[19]
//	  Let OffsetBits be the low-order 4 bits of String[19]
//	  Offset = StToNum(OffsetBits) // 0 <= OffSet <= 15
//	  Let P = String[OffSet]...String[OffSet+3]
//	  Return the Last 31 bits of P"
func DynamicTruncate(hmacResult []byte, digits int) string {
	offset := int(hmacResult[len(hmacResult)-1] & 0xF)
	binCode := []byte{
		hmacResult[offset] & 0x7f,
		hmacResult[offset+1],
		hmacResult[offset+2],
		hmacResult[offset+3],
	}
	return fmt.Sprintf("%0"+strconv.Itoa(digits)+"d", binary.BigEndian.Uint32(binCode)%uint32(math.Pow10(digits)))
}

// HOTP for the passed key and counter with the specified number of digits (min 6, max 8)
func HOTP(key string, counter uint64, length int) (string, error) {
	keyBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(key)
	if nil != err {
		return "", err
	}
	counterBytes := make([]byte, 8)
	if length < 6 || length > 8 {
		return "", errors.New("length must be within interval [6,8]")
	}
	binary.BigEndian.PutUint64(counterBytes, counter)
	cipher := hmac.New(sha1.New, keyBytes)
	n, err := cipher.Write(counterBytes)
	if nil != err {
		return "", err
	}
	if n != len(counterBytes) {
		return "", errors.New("unable to generate HOTP, unexpected number of bytes written (%d, %d)",
			n, len(counterBytes))
	}
	return DynamicTruncate(cipher.Sum(nil), length), nil
}

// HOTPCompare the HOTP for the specified key and the passed challenge
func HOTPCompare(key string, counter uint64, length int, challenge string) (bool, error) {
	hotp, err := HOTP(key, counter, length)
	if nil != err {
		return false, err
	}
	return subtle.ConstantTimeCompare([]byte(hotp), []byte(challenge)) == 1, nil
}

// TOTP for the passed key with the specified period (step size) and number of digits,
// step will be adjusted using the passed 'vary'
func TOTP(key string, period time.Duration, vary int, length int) (string, error) {
	currentStep := uint64(math.Floor(float64(time.Now().Unix()) / period.Seconds()))
	return HOTP(key, currentStep+uint64(vary), length)
}

// TOTPCompare the challenge to TOTP for a specific step dictated by period and adjust.
func TOTPCompare(key string, period time.Duration, adjust int, length int, challenge string) (bool, error) {
	totp, err := TOTP(key, period, adjust, length)
	if nil != err {
		return false, err
	}
	return subtle.ConstantTimeCompare([]byte(totp), []byte(challenge)) == 1, nil
}

// TOTPCompareWithVariance the expected TOTP calculation with the challenge in constant time.
// If variance is greater than 0, abs(variance) frames will be compared on either
// side of the 0 frame.
// Example:
//
//	Given the following values and offsets
//		TOTP():	|   A   |    B   |   C   |   D   |   E   |
//		offset:	|  -2   |   -1   |   0   |   1   |   2   |
//
// The following arguments would have the specified results:
//
//			Variance == ABS(Variance)
//	     Challenge	Variance 	Return
//			A			0		False
//			A			1		False
//			A			2		True
//			B			0		False
//			B			1		True
//			B			2		True
//			C			0		True
//			C			1		True
//			C			2		True
//			E			0		False
//			E			1		False
//			E			2		True
func TOTPCompareWithVariance(key string, period time.Duration, length int,
	variance uint, challenge string) (ok bool, err error) {
	ok, _, err = TOTPCompareAndGetDrift(key, period, length, variance, challenge, 0)
	return
}

// TOTPCompareAndGetDrift the expected TOTP calculation with the challenge in
// constant time.
func TOTPCompareAndGetDrift(key string, period time.Duration, length int,
	variance uint, challenge string, drift int) (bool, int, error) {
	matched := false
	driftActual := 0
	for i := drift - int(variance); i < drift+int(variance)+1; i++ {
		eq, err := TOTPCompare(key, period, i, length, challenge)
		if nil != err {
			return false, 0, err
		}
		if eq {
			matched = true
			driftActual = i
		}
	}
	return matched, driftActual, nil
}

// TOTPCompareAndGetDriftWithResynchronization will check the 0 drift case before
// comparing the passed drift. Executes in constants time in non-error conditions.
func TOTPCompareAndGetDriftWithResynchronization(key string, period time.Duration, length int,
	variance uint, challenge string, drift int) (bool, int, error) {
	resyncEqual, resyncdrift, err := TOTPCompareAndGetDrift(key, period, length,
		variance, challenge, 0)
	if nil != err {
		return false, 0, err
	}
	// always do both so we get constant time
	driftedEqual, driftedDrift, err := TOTPCompareAndGetDrift(key, period, length, variance, challenge, drift)
	if nil != err {
		return false, 0, err
	}
	if resyncEqual {
		return resyncEqual, resyncdrift, nil
	}
	return driftedEqual, driftedDrift, err
}

// GenerateTOTPURI for use in a QR code for registration with an authenticator application
func GenerateTOTPURI(key, issuer, user string, period time.Duration, length int) string {
	issuer = url.QueryEscape(issuer)
	user = url.QueryEscape(user)
	return fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%2.f",
		user, key, issuer, length, period.Seconds())
}

// GenerateTOTPQRCodePNG that can be served directly using content type header with 'image/png' or written to file.
func GenerateTOTPQRCodePNG(key, issuer, user string, period time.Duration, length int) ([]byte, error) {
	return qrcode.Encode(GenerateTOTPURI(key, issuer, user, period, length), qrcode.High, 256)
}
