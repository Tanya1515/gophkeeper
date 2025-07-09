package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// CreateInitVector - function for generating aesgm object for encrypting/decrypting
// sensetive data.
func (c *Crypto) CreateInitVector() (err error) {
	key := make([]byte, 2*aes.BlockSize)

	key = []byte("Gophekepeer encrypt/decrypt key.")

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("error while creating new cipher.Block: %w", err)
	}

	c.Aesgcm, err = cipher.NewGCM(aesblock)
	if err != nil {
		return fmt.Errorf("error while creating given 128-bit block cipher: %w", err)
	}

	return
}

// DecryptData - function for decrypting user sensetive data.
func (c *Crypto) DecryptData(sensetiveData string, initVector []byte) (string, error) {
	decodedData, err := base64.StdEncoding.DecodeString(sensetiveData)
	if err != nil {
		return "", fmt.Errorf("error, while decoding string: %w", err)
	}
	result, err := c.Aesgcm.Open(nil, initVector, decodedData, nil)
	if err != nil {
		return "", fmt.Errorf("error while decrypting data: %w\n", err)
	}

	return string(result), err
}

// EncryptData - function, that generates init vector and ecnode sensetive data
// for current user.
func (c *Crypto) EncryptData(sensetiveData string) (string, []byte, error) {
	incomingData := []byte(sensetiveData)

	vectInit := make([]byte, c.Aesgcm.NonceSize())

	_, err := rand.Read(vectInit)
	if err != nil {
		return "", nil, fmt.Errorf("error while generating init vector: %w", err)
	}

	result := c.Aesgcm.Seal(nil, vectInit, incomingData, nil)

	return base64.StdEncoding.EncodeToString(result), vectInit, nil
}
