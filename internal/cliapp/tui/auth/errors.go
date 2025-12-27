package auth

import "errors"

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrAlreadyLoggedIn       = errors.New("you are already logged in")
	ErrLoginCredentialsWrong = errors.New("username or password is wrong")
	ErrIncorrectFieldsFormat = errors.New("incorrect fields format")
	ErrUnknown               = errors.New("unknown error")
)
