package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var inputStrings = []string{
	"foo",
	"",
	"!aASDFa ads adsfl;ka !#",
	"!@#(AMADSLFA jafdk al1)",
	"pynes is a 🤡",
	"Hello, 世界",
}

func TestHashAndSalt(t *testing.T) {
	assert := assert.New(t)
	for _, password := range inputStrings {
		hash, salt := HashAndSalt(password)
		assert.NotEmpty(hash)
		assert.NotEmpty(salt)

		actual := Hash(password, salt)
		assert.Equal(hash, actual)

		actual = Hash(password, salt+"asf")
		assert.NotEqual(hash, actual)

		actual = Hash(password+"asf", salt)
		assert.NotEqual(hash, actual)
	}
}
