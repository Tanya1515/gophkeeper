package crypto

import "crypto/cipher"

// Crypto - structure, that containts object for generating
// key for encryption/decryption data.
type Crypto struct {
	Aesgcm cipher.AEAD
}
