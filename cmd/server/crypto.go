package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

func CreateInitVector() (aesgcm cipher.AEAD, vectInit []byte, err error) {
	key := make([]byte, 2*aes.BlockSize)

	_, err = rand.Read(key)
	if err != nil {
		return nil, nil, fmt.Errorf("error while creating key for encrypting sensetive data: %w", err)
	}

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("error while creating new cipher.Block: %w", err)
	}

	aesgcm, err = cipher.NewGCM(aesblock)
	if err != nil {
		return nil, nil, fmt.Errorf("error while creating given 128-bit block cipher: %w", err)
	}

	vectInit = make([]byte, aesgcm.NonceSize())

	_, err = rand.Read(vectInit)
	if err != nil {
		return nil, nil, fmt.Errorf("error while creating initialization vector: %w", err)
	}

	return
}

func (s *GophkeeperServer) DecryptData(sensetiveData string) (string, error) {
	incomingData := []byte(sensetiveData)
	result, err := s.Crypto.aesgcm.Open(nil, s.Crypto.InitVect, incomingData, nil)
	if err != nil {
		s.Logger.Errorf("Error while decrypting data: %s\n", err)
		return "", err
	}

	return string(result), err
}

func (s *GophkeeperServer) EncryptData(sensetiveData string) string {
	incomingData := []byte(sensetiveData)

	result := s.Crypto.aesgcm.Seal(nil, s.Crypto.InitVect, incomingData, nil)

	return string(result)
}
