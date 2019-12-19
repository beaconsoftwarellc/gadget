package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"

	"github.com/beaconsoftwarellc/gadget/errors"
)

const privateSlug = "RSA PRIVATE KEY"
const publicSlug = "RSA PUBLIC KEY"

// RSAPrivateKeyNotSetError  is returned when the RSA private key is not set and an operation needing a private
// key is called.
type RSAPrivateKeyNotSetError struct{ trace []string }

func (err *RSAPrivateKeyNotSetError) Error() string {
	return "RSA private key not set"
}

// Trace returns the stack trace for the error
func (err *RSAPrivateKeyNotSetError) Trace() []string {
	return err.trace
}

// NewRSAPrivateKeyNotSetError instantiates a RSAPrivateKeyNotSetError with a stack trace
func NewRSAPrivateKeyNotSetError() errors.TracerError {
	return &RSAPrivateKeyNotSetError{trace: errors.GetStackTrace()}
}

// RSAPublicKeyNotSetError  is returned when the RSA public key is not set and an operation needing a private
// key is called.
type RSAPublicKeyNotSetError struct{ trace []string }

func (err *RSAPublicKeyNotSetError) Error() string {
	return "RSA public key not set"
}

// Trace returns the stack trace for the error
func (err *RSAPublicKeyNotSetError) Trace() []string {
	return err.trace
}

// NewRSAPublicKeyNotSetError instantiates a RSAPublicKeyNotSetError with a stack trace
func NewRSAPublicKeyNotSetError() errors.TracerError {
	return &RSAPublicKeyNotSetError{trace: errors.GetStackTrace()}
}

// RSAEncryption provides 2048 bit rsa encryption with optional PSS Signing.
type RSAEncryption struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewRSAEncryption instance with no keys set.
func NewRSAEncryption() *RSAEncryption {
	return &RSAEncryption{}
}

// GetPrivateKey that is currently set on this instance of RSAEncryption
func (r *RSAEncryption) GetPrivateKey() *rsa.PrivateKey {
	return r.privateKey
}

// SetPrivateKey that will be used to decrypt and sign on this instance.
func (r *RSAEncryption) SetPrivateKey(key *rsa.PrivateKey) {
	r.privateKey = key
}

// GetPublicKey that is currently set on this instance.
func (r *RSAEncryption) GetPublicKey() *rsa.PublicKey {
	return r.publicKey
}

// SetPublicKey that will be used to encrypt and verify on this instance.
func (r *RSAEncryption) SetPublicKey(key rsa.PublicKey) {
	r.publicKey = &key
}

// GetType returns the cipher type this encryption instance provides.
func (r *RSAEncryption) GetType() CipherType {
	return RSA
}

// GenerateKey for 2048 bit rsa encryption.
func (r *RSAEncryption) GenerateKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if nil != err {
		panic(err)
	}
	return key
}

// MarshalPrivateKey data type (PKCS1) and return as bytes.
func (r *RSAEncryption) MarshalPrivateKey() ([]byte, error) {
	if nil == r.privateKey {
		return nil, NewRSAPrivateKeyNotSetError()
	}
	pkcs1key := x509.MarshalPKCS1PrivateKey(r.privateKey)
	block := &pem.Block{
		Type:  privateSlug,
		Bytes: pkcs1key,
	}
	pemdata := pem.EncodeToMemory(block)
	if nil == pemdata {
		return nil, errors.New("marshal private key failed on encode")
	}
	return pemdata, nil
}

// UnmarshallPrivateKey from the passed bytes created from `MarshalPrivateKey`
// and set it on this instance.
func (r *RSAEncryption) UnmarshallPrivateKey(bytes []byte) error {
	// extra bytes are ignored
	block, _ := pem.Decode(bytes)
	if nil == block {
		return errors.New("unmarshal private key failed on decode")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if nil == err {
		r.privateKey = key
	}
	return err
}

func (r *RSAEncryption) marshalPublicKey(key *rsa.PublicKey) ([]byte, error) {
	pubasn1, err := x509.MarshalPKIXPublicKey(key)
	if nil != err {
		return nil, err
	}
	block := &pem.Block{
		Type:  publicSlug,
		Bytes: pubasn1,
	}
	return pem.EncodeToMemory(block), nil
}

// MarshalPrivatePublicKey to data type PubASN1 PEM format and return as bytes.
func (r *RSAEncryption) MarshalPrivatePublicKey() ([]byte, error) {
	if nil == r.privateKey {
		return nil, NewRSAPrivateKeyNotSetError()
	}
	return r.marshalPublicKey(&r.privateKey.PublicKey)
}

// MarshalPublicKey data type (PubASN1) and return as bytes.
func (r *RSAEncryption) MarshalPublicKey() ([]byte, error) {
	if nil == r.publicKey {
		return nil, NewRSAPublicKeyNotSetError()
	}
	return r.marshalPublicKey(r.publicKey)
}

// UnmarshallPublicKey from the passed bytes created using MarshalPublicKey and
// set it on this instance.
func (r *RSAEncryption) UnmarshallPublicKey(bytes []byte) error {
	block, _ := pem.Decode(bytes)
	if nil == block {
		return errors.New("unmarshal public key failed on decode")
	}
	anon, err := x509.ParsePKIXPublicKey(block.Bytes)
	if nil != err {
		return err
	}
	key, ok := anon.(*rsa.PublicKey)
	if !ok {
		return errors.New("public key was not rsa")
	}
	r.publicKey = key
	return nil
}

// Encrypt the passed plaintext using the passed public key.
func (r *RSAEncryption) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	if nil == r.publicKey {
		return nil, NewRSAPublicKeyNotSetError()
	}
	hash := crypto.SHA256.New()
	ciphertext, err = rsa.EncryptOAEP(hash, rand.Reader, r.publicKey, plaintext, []byte("kasita"))
	return ciphertext, err
}

// Sign with RSASSA-PSS
func (r *RSAEncryption) Sign(plaintext []byte) (signed []byte, err error) {
	if nil == r.privateKey {
		return nil, NewRSAPrivateKeyNotSetError()
	}
	hashed := sha256.Sum256(plaintext)
	signed, err = rsa.SignPSS(rand.Reader, r.privateKey, crypto.SHA256, hashed[:], nil)
	return
}

// EncryptAndSign the passed plaintext with the passed encryption key and signing key.
func (r *RSAEncryption) EncryptAndSign(plaintext []byte) (
	signature []byte, ciphertext []byte, err error) {
	ciphertext, err = r.Encrypt(plaintext)
	if nil != err {
		return
	}
	signature, err = r.Sign(plaintext)
	return
}

// Decrypt the passed ciphertext using the passed private key.
func (r *RSAEncryption) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	if nil == r.privateKey {
		return nil, NewRSAPrivateKeyNotSetError()
	}
	plaintext, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, r.privateKey, ciphertext, []byte("kasita"))
	return plaintext, err
}

// Verify that the passed signature matches the signature of the plaintext encrypted using the
// private key corresponding to the passed public key.
func (r *RSAEncryption) Verify(plaintext []byte, signature []byte) error {
	if nil == r.publicKey {
		return NewRSAPublicKeyNotSetError()
	}
	hashed := sha256.Sum256(plaintext)
	return rsa.VerifyPSS(r.publicKey, crypto.SHA256, hashed[:], signature, nil)
}

// DecryptAndVerify decrypts the passed ciphertext and verifies the signature.
func (r *RSAEncryption) DecryptAndVerify(ciphertext []byte, signature []byte) (
	plaintext []byte, err error) {
	plaintext, err = r.Decrypt(ciphertext)
	if nil != err {
		return
	}
	err = r.Verify(plaintext, signature)
	return
}
