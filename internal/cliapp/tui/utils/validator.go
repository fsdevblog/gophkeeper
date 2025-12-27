package utils

import (
	"errors"
	"strings"
)

const (
	maxUsernameLength = 16
	maxUsernameBytes  = 64
	maxPasswordBytes  = 72
)

func ValidateUsernameField(s string) error {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return errors.New("username required")
	}
	if len(s) > maxUsernameLength || len([]byte(s)) > maxUsernameBytes {
		return errors.New("username too long")
	}
	return nil
}

func ValidatePasswordField(s string) error {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return errors.New("password required")
	}
	if len([]byte(s)) > maxPasswordBytes {
		return errors.New("password too long")
	}
	return nil
}
