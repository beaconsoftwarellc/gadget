package crypto

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/generator"
)

// AES256KeySize for AWS256 Encryption
const AES256KeySize = 32

// IncompleteDataError  returned when an incomplete ciphertext is passed to decrypt.
type IncompleteDataError struct{ trace []string }

func (err *IncompleteDataError) Error() string {
	return "ciphertext data was incomplete"
}

// Trace returns the stack trace for the error
func (err *IncompleteDataError) Trace() []string {
	return err.trace
}

// NewIncompleteDataError instantiates a IncompleteDataError with a stack trace
func NewIncompleteDataError() errors.TracerError {
	return &IncompleteDataError{trace: errors.GetStackTrace()}
}

// AESEncryption provides AES256 Encryption with GCM tampering detection.
type AESEncryption struct {
	key   []byte
	block cipher.Block
	gcm   cipher.AEAD
}

// NewAES using the passed key, if nil is passed a new key will be generated.
func NewAES(key []byte) (Encryption, error) {
	a := &AESEncryption{}
	var err error
	if nil == key || len(key) == 0 {
		a.RotateKey()
	} else {
		err = a.SetKey(key)
	}
	return a, err
}

// GetType returns the cipher type this instance of encryption provides.
func (a *AESEncryption) GetType() CipherType {
	return AES
}

// GenerateKey will create a new key to use with this instance of AES
func (AESEncryption) GenerateKey() []byte {
	return generator.Bytes(AES256KeySize)
}

// GetKey currently being used by this instance of AES
func (a *AESEncryption) GetKey() []byte {
	return a.key
}

// RotateKey generates a new AES256 key and sets for use on this instance and returns it.
func (a *AESEncryption) RotateKey() []byte {
	a.SetKey(a.GenerateKey())
	return a.key
}

// SetKey for use on this instance of AES256.
func (a *AESEncryption) SetKey(key []byte) error {
	var err error
	block, err := aes.NewCipher(key)
	if nil != err {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if nil != err {
		return err
	}
	a.key = key
	a.block = block
	a.gcm = gcm
	return nil
}

// Encrypt with AES256-GCM
func (a *AESEncryption) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	nonce := generator.Bytes(a.gcm.NonceSize())
	return a.gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt data using AES256-GCM
func (a *AESEncryption) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	if len(ciphertext) < a.gcm.NonceSize() {
		return nil, NewIncompleteDataError()
	}

	return a.gcm.Open(nil,
		ciphertext[:a.gcm.NonceSize()],
		ciphertext[a.gcm.NonceSize():],
		nil,
	)
}

// Sign does nothing with AES
func (a *AESEncryption) Sign(plaintext []byte) (signature []byte, err error) {
	return []byte{}, nil
}

// Verify does nothing with AES
func (a *AESEncryption) Verify(plaintext []byte, signature []byte) (err error) {
	return nil
}
