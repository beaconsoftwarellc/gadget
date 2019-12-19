package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/generator"
)

func TestGenerateKey(t *testing.T) {
	assert := assert.New(t)
	rsaEncryption := NewRSAEncryption()
	expected := rsaEncryption.GenerateKey()
	assert.NotNil(expected)
}

func TestMarshalPrivateKey(t *testing.T) {
	assert := assert.New(t)
	rsae := &RSAEncryption{}
	rsae.SetPrivateKey(rsae.GenerateKey())
	assert.NotNil(rsae.GetPrivateKey())
	assert.Nil(rsae.GetPublicKey())
	actual, err := rsae.MarshalPrivateKey()
	assert.NoError(err)
	assert.NotNil(actual)
	stringActual := string(actual)
	assert.True(strings.Contains(stringActual, "BEGIN RSA PRIVATE KEY"))
	assert.True(strings.Contains(stringActual, "END RSA PRIVATE KEY"))
}

func TestMarshalUnmarshalPrivate(t *testing.T) {
	assert := assert.New(t)
	rsae := &RSAEncryption{}
	expected := rsae.GenerateKey()
	rsae.SetPrivateKey(expected)
	bytes, err := rsae.MarshalPrivateKey()
	assert.NoError(err)
	assert.NoError(rsae.UnmarshallPrivateKey(bytes))
	assert.NoError(err)
	assert.Equal(expected, rsae.GetPrivateKey())
}

func TestMarshalPublicKey(t *testing.T) {
	assert := assert.New(t)
	rsae := &RSAEncryption{}
	key := rsae.GenerateKey()
	assert.NotNil(key)
	rsae.SetPublicKey(key.PublicKey)
	actual, err := rsae.MarshalPublicKey()
	assert.NoError(err)
	assert.NotNil(actual)
	stringActual := string(actual)
	assert.True(strings.Contains(stringActual, "BEGIN RSA PUBLIC KEY"))
	assert.True(strings.Contains(stringActual, "END RSA PUBLIC KEY"))
}

func TestMarshalUnmarshalPublic(t *testing.T) {
	assert := assert.New(t)
	rsae := &RSAEncryption{}
	key := rsae.GenerateKey()
	expected := key.PublicKey
	rsae.SetPublicKey(expected)
	bytes, err := rsae.MarshalPublicKey()
	assert.NoError(err)
	assert.NoError(rsae.UnmarshallPublicKey(bytes))
	assert.NoError(err)
	assert.Equal(&expected, rsae.publicKey)
}

func TestEncrypt(t *testing.T) {
	assert := assert.New(t)
	rsae := &RSAEncryption{}
	rsae.SetPublicKey(rsae.GenerateKey().PublicKey)
	plaintext := generator.String(190) // this is the MAX for our keys
	ciphertext, err := rsae.Encrypt([]byte(plaintext))
	assert.NoError(err)
	assert.NotNil(ciphertext)
}

func TestEncryptDecrypt(t *testing.T) {
	assert := assert.New(t)
	encrypt := &RSAEncryption{}
	encrypt.SetPrivateKey(encrypt.GenerateKey())
	decrypt := NewRSAEncryption()
	decrypt.SetPrivateKey(decrypt.GenerateKey())

	encrypt.SetPublicKey(decrypt.GetPrivateKey().PublicKey)
	decrypt.SetPublicKey(encrypt.GetPrivateKey().PublicKey)

	expected := []byte(generator.String(190)) // this is the MAX for our keys
	ciphertext, err := encrypt.Encrypt(expected)
	assert.NoError(err)
	actual, err := decrypt.Decrypt(ciphertext)
	assert.NoError(err)
	assert.Equal(expected, actual)
}

func TestSign(t *testing.T) {
	assert := assert.New(t)
	rsae := NewRSAEncryption()
	signKey := rsae.GenerateKey()
	rsae.SetPrivateKey(signKey)
	plaintext := []byte(generator.String(100))
	signature, err := rsae.Sign(plaintext)
	assert.NoError(err)
	assert.NotNil(signature)
}

func TestSignVerify(t *testing.T) {
	assert := assert.New(t)
	rsae := &RSAEncryption{}
	signKey := rsae.GenerateKey()
	rsae.SetPrivateKey(signKey)
	rsae.SetPublicKey(signKey.PublicKey)
	plaintext := []byte(generator.String(100))
	signature, err := rsae.Sign(plaintext)
	assert.NoError(err)
	err = rsae.Verify(plaintext, signature)
	assert.NoError(err)
}

func TestVerifyFail(t *testing.T) {
	assert := assert.New(t)
	r1 := &RSAEncryption{}
	r2 := &RSAEncryption{}
	r1.SetPrivateKey(r1.GenerateKey())
	r2.SetPrivateKey(r2.GenerateKey())
	r1.SetPublicKey(r2.GetPrivateKey().PublicKey)
	r2.SetPublicKey(r1.GetPrivateKey().PublicKey)

	plaintext := []byte(generator.String(100))

	signature, err := r1.Sign(plaintext)
	assert.NoError(err)
	plaintext[10] = plaintext[10] + 1
	err = r2.Verify(plaintext, signature)
	assert.EqualError(err, "crypto/rsa: verification error")
}

func TestEncryptAndSign(t *testing.T) {
	assert := assert.New(t)
	r1 := &RSAEncryption{}
	r2 := &RSAEncryption{}
	r1.SetPrivateKey(r1.GenerateKey())
	r2.SetPrivateKey(r2.GenerateKey())
	r1.SetPublicKey(r2.GetPrivateKey().PublicKey)
	r2.SetPublicKey(r1.GetPrivateKey().PublicKey)

	plaintext := []byte(generator.String(100))
	signature, ciphertext, err := r1.EncryptAndSign(plaintext)
	assert.NoError(err)
	assert.NotNil(signature)
	assert.NotNil(ciphertext)
}

func TestEncryptAndSignDecryptAndVerify(t *testing.T) {
	assert := assert.New(t)
	r1 := &RSAEncryption{}
	r2 := &RSAEncryption{}
	r1.SetPrivateKey(r1.GenerateKey())
	r2.SetPrivateKey(r2.GenerateKey())
	r1.SetPublicKey(r2.GetPrivateKey().PublicKey)
	r2.SetPublicKey(r1.GetPrivateKey().PublicKey)

	plaintext := []byte(generator.String(90))
	signature, ciphertext, err := r1.EncryptAndSign(plaintext)
	assert.NoError(err)

	actual, err := r2.DecryptAndVerify(ciphertext, signature)
	assert.NoError(err)
	assert.Equal(plaintext, actual)
}
