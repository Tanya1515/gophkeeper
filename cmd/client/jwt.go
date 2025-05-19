package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func SaveJWT(JWTToken, userLogin string) error {
	file, err := os.OpenFile(".configs/configJWT.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	var userJWTByte []byte
	userJWT := make(map[string]string)

	_, err = file.Read(userJWTByte)
	if err != nil {
		return fmt.Errorf("error while reading user JWT tokens from file: %w", err)
	}

	err = json.Unmarshal(userJWTByte, &userJWT)
	if err != nil {
		return err
	}

	_, exists := userJWT[userLogin]

	if exists {
		userJWT[userLogin] = JWTToken
		userJWTByte, err = json.Marshal(userJWT)
		if err != nil {
			return err
		}

		_, err = file.Write(userJWTByte)
		if err != nil {
			return err
		}
		return nil
	}

	userJWT[userLogin] = JWTToken

	userJWTByte, err = json.Marshal(userJWT)
	if err != nil {
		return err
	}
	// перезатереть содержимое файла ?
	_, err = file.Write(userJWTByte)
	if err != nil {
		return err
	}

	defer file.Close()
	return nil
}

func GetJWT(userLogin string) (JWTToken string, err error) {

	file, err := os.OpenFile(".configs/configJWT.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	var userJWTByte []byte
	userJWT := make(map[string]string)

	_, err = file.Read(userJWTByte)
	if err != nil {
		return
	}

	err = json.Unmarshal(userJWTByte, &userJWT)
	if err != nil {
		return
	}

	JWTToken, exists := userJWT[userLogin]

	if exists {
		return
	}

	defer file.Close()
	return
}

func CreateJWTPath() error {
	path := ".configs"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			fmt.Println("Error while creating directory for saving JWT tokens: ", err)
			return err
		}
	}

	fileJWT := ".configs/configJWT.json"
	if _, err := os.Stat(fileJWT); errors.Is(err, os.ErrNotExist) {
		file, err := os.OpenFile(fileJWT, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			fmt.Println("Error while creating file for saving JWT tokens: ", err)
			return err
		}

		file.WriteString("File for saving user JWT tokens in format: user:JWTtoken\n")
		defer file.Close()
	}
	return nil
}
