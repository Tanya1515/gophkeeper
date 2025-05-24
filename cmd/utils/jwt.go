package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserLogin string
}

const TokenExp = time.Hour

func GenerateToken(userLogin string) (JWTtoken string, err error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserLogin: userLogin,
	})

	JWTtoken, err = token.SignedString([]byte("secretKey"))
	if err != nil {
		return
	}

	return
}

func ProcessJWTToken(JWTToken string) (userLogin string, err error) {
	claims := Claims{}

	jwt.ParseWithClaims(JWTToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("secretKey"), nil
	})

	err = claims.Valid()
	if err != nil {
		return "", err
	}

	return claims.UserLogin, err

}

func SaveJWT(JWTToken, userLogin string) error {

	var err error
	var userJWTByte []byte
	userJWT := make(map[string]string)

	userJWTByte, err = os.ReadFile(".configs/configJWT.json")
	if err != nil {
		return fmt.Errorf("error while reading user JWT tokens from file: %w", err)
	}

	if len(userJWTByte) != 0 {
		err = json.Unmarshal(userJWTByte, &userJWT)
		if err != nil {
			return fmt.Errorf("error while unmarhalling data: %w", err)
		}
	}

	_, exists := userJWT[userLogin]

	if exists {
		userJWT[userLogin] = JWTToken
		userJWTByte, err = json.Marshal(userJWT)
		if err != nil {
			return err
		}

		err = os.WriteFile(".configs/configJWT.json", userJWTByte, fs.ModeAppend)
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

	err = os.WriteFile(".configs/configJWT.json", userJWTByte, fs.ModeAppend)
	if err != nil {
		return err
	}

	return nil
}

func GetJWT(userLogin string) (JWTToken string, err error) {

	var userJWTByte []byte
	userJWT := make(map[string]string)

	userJWTByte, err = os.ReadFile(".configs/configJWT.json")
	if err != nil {
		return "", fmt.Errorf("error while reading file with JWT tokens: %w", err)
	}

	err = json.Unmarshal(userJWTByte, &userJWT)
	if err != nil {
		return "", fmt.Errorf("error while unmarshalling data from file: %w", err)
	}

	JWTToken, exists := userJWT[userLogin]

	if exists {
		return "", fmt.Errorf("jwttoken for user %s does not exist, please login or register", userLogin)
	}

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
