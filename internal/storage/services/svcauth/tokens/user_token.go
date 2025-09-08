package tokens

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	jwt.RegisteredClaims
	ID uuid.UUID
}

func GenerateUserJWT(id uuid.UUID, expire time.Duration, key []byte) (string, error) {
	userClaims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
		},
		ID: id,
	}
	token, err := generateJWT(userClaims, key)
	if err != nil {
		return "", fmt.Errorf("generating user jwt token: %s", err.Error())
	}
	return token, nil
}

func ValidateUserJWT(tokenString string, key []byte) (*jwt.Token, error) {
	token, err := validateJWT(tokenString, new(UserClaims), key)
	if err != nil {
		return nil, fmt.Errorf("validating user jwt token: %w", err)
	}

	_, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return token, nil
}
