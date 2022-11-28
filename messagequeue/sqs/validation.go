package sqs

import (
	"strings"
	"unicode/utf8"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

const (
	minNameCharacters = 1
	maxNameCharacters = 256
	minBodyCharacters = 1
	maxBodyKilobytes  = 255
	prohibitedAWS     = "aws"
	prohibitedAmazon  = "amazon"
	period            = "."
)

var (
	prohibitedPrefixError = errors.
				New("name has invalid prefix (%s|%s)", prohibitedAmazon, prohibitedAWS)
	dotError         = errors.New("name cannot begin, end, or contain sequences of '%s'", period)
	bodyMinimumError = "minimum character count is 1"
)

// NameIsValid for use as an attribute or system attribute name
func NameIsValid(s string) error {
	/* TODO: [COR-544] SQS Name Validation is incomplete
	Name – The message attribute name can contain the following characters:
		A-Z
		a-z
		0-9
		underscore (_)
		hyphen (-)
		period (.)
	The following restrictions apply:
		- Can be up to 256 characters long
		- Can't start with AWS. or Amazon. (or any casing variations)
		- Is case-sensitive
		- Must be unique among all attribute names for the message
		- Must not start or end with a period
		- Must not have periods in a sequence
	*/
	runeCount := utf8.RuneCountInString(s)
	if runeCount < minBodyCharacters || runeCount > maxNameCharacters {
		return errors.New("name character count out of bounds [%d, %d] (%d)",
			minNameCharacters, maxNameCharacters, runeCount)
	}
	if strings.HasPrefix(s, period) || strings.HasSuffix(s, period) ||
		strings.Contains(s, period+period) {
		return dotError
	}
	low := strings.ToLower(s)
	if strings.HasPrefix(low, prohibitedAmazon) || strings.HasPrefix(low, prohibitedAWS) {
		return prohibitedPrefixError
	}
	return nil
}

// BodyIsValid for use as a attribute value or a message body
func BodyIsValid(s string) error {
	/* TODO: [COR-545] SQS Body Validation is incomplete
	The minimum size is one character. The maximum size is 256 KB.

	A message can include only XML, JSON, and unformatted text. The following
	Unicode characters are allowed:

	#x9 | #xA | #xD | #x20 to #xD7FF | #xE000 to #xFFFD | #x10000 to #x10FFFF

	Any characters not included in this list will be rejected. For more information,
	see the W3C specification for characters (http://www.w3.org/TR/REC-xml/#charsets).
	*/
	if utf8.RuneCountInString(s) == 0 {
		return errors.New(bodyMinimumError)
	}
	if len(s) > maxBodyKilobytes {
		return errors.New("body cannot exceed %d kilobytes (was %d kb)",
			maxBodyKilobytes, len(s)/1024)
	}
	return nil
}