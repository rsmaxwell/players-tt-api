package basic

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	mySigningKey = []byte("<SESSION_SECRET_KEY>")
)

type MyJwtClaims struct {
	ID      int `json:"id"`
	Request int `json:"request"`
	jwt.StandardClaims
}

// GenerateToken generates a jwt token
func GenerateToken(id int, request int, expiresAfter time.Duration) (string, error) {

	claims := MyJwtClaims{
		ID:      id,
		Request: request,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiresAfter).Unix(),
			Issuer:    "test",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}

// ValidateToken validates the jwt token
func ValidateToken(signedToken string) (*MyJwtClaims, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&MyJwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return mySigningKey, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyJwtClaims)
	if !ok {
		err = errors.New("couldn't parse jwt claims")
		return nil, err
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		err = errors.New("jwt is expired")
		return nil, err
	}

	return claims, nil
}
