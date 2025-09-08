package svcauth

import "errors"

var (
	ErrInvalidPassword       = errors.New("password is invalid")
	ErrUserAlreadyRegistered = errors.New("user already registered")
)
