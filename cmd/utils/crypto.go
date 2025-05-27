package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

func CreateInitVector() (vectInit []byte, err error) {
	key := make([]byte, 2*aes.BlockSize)

	_, err = rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("error while creating key for encrypting sensetive data: %w", err)
	}

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error while creating new cipher.Block: %w", err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("error while creating given 128-bit block cipher: %w", err)
	}

	vectInit = make([]byte, aesgcm.NonceSize())

	_, err = rand.Read(vectInit)
	if err != nil {
		return nil, fmt.Errorf("error while creating initialization vector: %w", err)
	}

	return
}

func DecryptData(sensetiveData string, vectInit []byte) (result string, err error) {

	// result, err = aesgcm.Open(nil, vectInit, sensetiveData, nil) // расшифровываем
	// if err != nil {
	// 	fmt.Printf("error: %v\n", err)
	// 	return
	// }

	return
}

func EncryptData(sensetiveData string, vectInit []byte) (result string, err error) {

	// result = aesgcm.Seal(nil, vectInit, sensetiveData, nil)

	return
}
