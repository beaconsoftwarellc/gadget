package generator

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// AlphaRunes is the runes [a-zA-Z]
	AlphaRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	// HexRunes are those digits and characters that can occur as part of
	// a hexadecimal number representation (0-9, A-F)
	HexRunes = []rune("0123456789ABCDEF")
	// DigitRunes are the 10 digits [0-9]
	DigitRunes = []rune("0123456789")
	// SymbolRunes are the special characters [!@#$%`'^&*()_+]
	SymbolRunes = []rune("!@#$%`'^&*()_+")
	// UnreservedSymbolRunes are the symbols that are not reserved in URLs
	UnreservedSymbolRunes = ("-._~")
	base32Runes           = []rune("13456789abcdefghijkmnopqrstuwxyz")
)

/*
	 Excludes alphanumerics that are not easily distinguishable in some common fonts:
		0,O
		1,l,I
		2,Z
		5,S
		8,B
*/
var distinguishables = []rune("34679abcdefghijkmnopqrstuvwxyzACDEFGHJKMNPQRTUVWXY")

// Password generates a random password of a given length
func Password(length int) string {
	source := append(AlphaRunes, SymbolRunes...)
	source = append(source, DigitRunes...)
	return random(length, source)
}

func random(length int, source []rune) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	maxInt := len(source) - 1

	randomString := make([]rune, length)
	for i := range randomString {
		randomString[i] = source[r.Intn(maxInt)]
	}
	return string(randomString)
}

// IDPrefix is the human readable prefix that is attached to a generated ID
type IDPrefix string

// IDSizeBytes is the length of an identifier in bytes in the system.
const IDSizeBytes = 32

// Base32IDSizeBytes is the length of a base 32 identifier in bytes in the system.
const Base32IDSizeBytes = 18

// MaxPrefix is the maximum ID prefix length
const MaxPrefix = 8

// ID creates a random id starting with prefix to idSize length
func ID(prefix IDPrefix) string {
	if len(prefix) > MaxPrefix {
		panic(fmt.Sprintf("%s is too long of a prefix ID", prefix))
	}
	n := IDSizeBytes - len(prefix) - 1
	s := base64.StdEncoding.EncodeToString([]byte(strings.Replace(uuid.New().String(), "-", "", -1)))
	return fmt.Sprintf("%s_%s", prefix, s[:n])
}

// Base32ID creates a random id starting with prefix to idSize length
func Base32ID(prefix IDPrefix) string {
	if len(prefix) > MaxPrefix {
		panic(fmt.Sprintf("%s is too long of a prefix ID", prefix))
	}
	n := Base32IDSizeBytes - len(prefix) - 1
	return fmt.Sprintf("%s_%s", prefix, random(n, base32Runes))
}

const secretLength = 32

// Secret returns a string that is suitable as a Salt or a Client Secret
func Secret() string {
	return Password(secretLength)
}

// Code returns random letters & numbers of the requested length
func Code(length int) string {
	return random(length, append(AlphaRunes, DigitRunes...))
}

// Hex returns a hex string of the requested length
func Hex(length int) string {
	return random(length, HexRunes)
}

/*
   Random generators to assist with testing
*/

// TestID returns an ID with a random prefix for testing purposes
func TestID() string {
	return ID(IDPrefix(String(3)))
}

// String returns random letters of the requested length
func String(length int) string {
	return random(length, AlphaRunes)
}

// HexColor returns a 7 length string from #000000 - #FFFFFF
func HexColor() string {
	return "#" + random(6, HexRunes)
}

// Year returns random year between 1976 and 2017
func Year() int32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return int32(r.Intn(41) + 1976)
}

// Email returns a fake email address for testing
func Email() string {
	return fmt.Sprintf("fake+%s@gadget.com", random(10, AlphaRunes))
}

// Name returns a fake name for testing
func Name() string {
	return fmt.Sprintf("%s %s", random(6, AlphaRunes), random(12, AlphaRunes))
}

// Distinguishable returns random alphanumerics of the requested length,
// excluding similar characters (e.g. 0, O, 1, l, I)
func Distinguishable(length int) string {
	return random(length, distinguishables)
}
