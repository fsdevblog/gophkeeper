package svcauth

import "errors"

var (
	ErrInvalidPassword       = errors.New("password is invalid")
	ErrInvalidDevice         = errors.New("device is invalid")
	ErrUserAlreadyRegistered = errors.New("user already registered")
)
