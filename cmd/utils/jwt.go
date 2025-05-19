package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserLogin    string
	UserPassword string
}

const TokenExp = time.Hour

func GenerateToken(userLogin string) (JWTtoken string, err error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserLogin:    userLogin,
	})

	JWTtoken, err = token.SignedString([]byte("secretKey"))
	if err != nil {
		return
	}

	return
}
