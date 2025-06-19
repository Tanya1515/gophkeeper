package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func CreateInitVector() (aesgcm cipher.AEAD, err error) {
	key := make([]byte, 2*aes.BlockSize)

	key = []byte("Gophekepeer encrypt/decrypt key.")

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error while creating new cipher.Block: %w", err)
	}

	aesgcm, err = cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("error while creating given 128-bit block cipher: %w", err)
	}

	return
}

func (s *GophkeeperServer) DecryptData(sensetiveData string, initVector []byte) (string, error) {
	decodedData, err := base64.StdEncoding.DecodeString(sensetiveData)
	if err != nil {
		return "", fmt.Errorf("error, while decoding string: %w", err)
	}
	result, err := s.Crypto.aesgcm.Open(nil, initVector, decodedData, nil)
	if err != nil {
		s.Logger.Errorf("Error while decrypting data: %s\n", err)
		return "", err
	}

	return string(result), err
}

func (s *GophkeeperServer) EncryptData(sensetiveData string) (string, []byte) {
	incomingData := []byte(sensetiveData)

	vectInit := make([]byte, s.Crypto.aesgcm.NonceSize())

	_, err := rand.Read(vectInit)
	if err != nil {
		s.Logger.Errorln("Error while generating init vector: %s", err)
		return "", nil
	}

	result := s.Crypto.aesgcm.Seal(nil, vectInit, incomingData, nil)

	return base64.StdEncoding.EncodeToString(result), vectInit
}
