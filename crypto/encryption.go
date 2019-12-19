package crypto

// CipherType represents how the message body will be encrypted.
type CipherType uint8

const (
	// None specifies no encryption. Suitable only for Negotiate requests.
	None CipherType = 0
	// AES symmetric encryption
	AES CipherType = 1
	// RSA asymmetric small message encryption
	RSA CipherType = 2
)

func (ct CipherType) String() string {
	switch ct {
	case None:
		return "None"
	case AES:
		return "AES"
	case RSA:
		return "RSA"
	default:
		return "Unknown"
	}
}

// Encryption interface provides the necessary methods for an encryption provider.
type Encryption interface {
	GetType() CipherType
	Encrypt(plaintext []byte) (ciphertext []byte, err error)
	Decrypt(ciphertext []byte) (plaintext []byte, err error)
	Sign(plaintext []byte) (signature []byte, err error)
	Verify(plaintext []byte, signature []byte) (err error)
}

// NewNoEncryption returns an instance of NoEncryption which can be used as a pass through.
func NewNoEncryption() Encryption {
	return &NoEncryption{}
}

// NoEncryption provides a passthrough for when you need an Encryption object but don't actually want
// encryption.
type NoEncryption struct{}

// GetType of cipher on this Encryption.
func (ne *NoEncryption) GetType() CipherType {
	return None
}

// Encrypt returns the plaintext
func (ne *NoEncryption) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	return plaintext, nil
}

// Decrypt returns the ciphertext
func (ne *NoEncryption) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	return ciphertext, nil
}

// Sign the passed plaintext and return a signature that can be used to verify that the
// data was signed using this instance of encryptions key.
func (ne *NoEncryption) Sign(plaintext []byte) (signature []byte, err error) {
	return []byte{}, nil
}

// Verify the passed signature against the key on this instance. Returns err on failure.
func (ne *NoEncryption) Verify(plaintext []byte, signature []byte) (err error) {
	return nil
}
