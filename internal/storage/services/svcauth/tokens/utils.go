package tokens

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func generateJWT(claims jwt.Claims, key []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("generating jwt token: %s", err.Error())
	}

	return tokenString, nil
}

func validateJWT(tokenString string, claims jwt.Claims, key []byte) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (any, error) {
		return key, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("parsing jwt token `%s`: %w", tokenString, err)
	}

	return token, nil
}
