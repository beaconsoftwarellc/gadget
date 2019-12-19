package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/generator"
)

func TestNewAES(t *testing.T) {
	assert := assert.New(t)
	encryption, err := NewAES(nil)
	aes := encryption.(*AESEncryption)
	assert.NoError(err)
	assert.NotNil(aes)
	assert.NotNil(aes.key)
	assert.NotNil(aes.block)
	assert.NotNil(aes.gcm)
	assert.NotEmpty(aes.GetKey())
	// uncomment this line to get a new key
	// assert.Fail(base64.StdEncoding.EncodeToString(aes.GetKey()))
}

func TestNewAESPassedKey(t *testing.T) {
	assert := assert.New(t)
	key := AESEncryption{}.GenerateKey()
	encryption, err := NewAES(key)
	aes := encryption.(*AESEncryption)
	assert.NoError(err)
	assert.NotNil(aes)
	assert.NotNil(aes.key)
	assert.NotNil(aes.block)
	assert.NotNil(aes.gcm)
	assert.NotEmpty(aes.GetKey())
}

func TestNewAESBadKeyPanics(t *testing.T) {
	assert := assert.New(t)
	_, err := NewAES(generator.Bytes(AES256KeySize + 5))
	assert.EqualError(err, "crypto/aes: invalid key size 37")
}

func TestAESEncrypt(t *testing.T) {
	assert := assert.New(t)
	encryption, err := NewAES(nil)
	aes := encryption.(*AESEncryption)
	assert.NoError(err)
	plaintext := generator.Bytes(500)
	ciphertext, err := aes.Encrypt(plaintext)
	assert.NoError(err)
	assert.NotEqual(plaintext, ciphertext)
}

func TestAESDecryptNonsense(t *testing.T) {
	assert := assert.New(t)
	encryption, err := NewAES(nil)
	aes := encryption.(*AESEncryption)
	assert.NoError(err)
	plaintext, err := aes.Decrypt(generator.Bytes(100))
	assert.Nil(plaintext)
	assert.EqualError(err, "cipher: message authentication failed")
}

func TestAESDecryptModified(t *testing.T) {
	assert := assert.New(t)
	encryption, err := NewAES(nil)
	aes := encryption.(*AESEncryption)
	assert.NoError(err)
	plaintext := generator.Bytes(100)
	ciphertext, err := aes.Encrypt(plaintext)
	assert.NoError(err)
	ciphertext[30] = ciphertext[30] + 2
	actual, err := aes.Decrypt(ciphertext)
	assert.Nil(actual)
	assert.EqualError(err, "cipher: message authentication failed")
}

func TestAESEncryptDecrypt(t *testing.T) {
	assert := assert.New(t)
	encryption, err := NewAES(nil)
	aes := encryption.(*AESEncryption)
	assert.NoError(err)
	expected := generator.Bytes(100)
	ciphertext, err := aes.Encrypt(expected)
	assert.NoError(err)
	actual, err := aes.Decrypt(ciphertext)
	assert.Equal(expected, actual)
	assert.NoError(err)
}

func TestAESEncryptDecryptDifferentInstances(t *testing.T) {
	assert := assert.New(t)
	encryption, err := NewAES(nil)
	aes := encryption.(*AESEncryption)
	assert.NoError(err)
	aes2, err := NewAES(aes.GetKey())
	assert.NoError(err)
	expected := generator.Bytes(4000)
	ciphertext, err := aes.Encrypt(expected)
	assert.NoError(err)
	actual, err := aes2.Decrypt(ciphertext)
	assert.Equal(expected, actual)
	assert.NoError(err)
}
