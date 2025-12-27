package svcauth

import "errors"

var (
	ErrInvalidCredentials    = errors.New("login or password is invalid")
	ErrUserAlreadyRegistered = errors.New("user already registered")
)
