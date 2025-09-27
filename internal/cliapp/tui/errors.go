package tui

import "errors"

var (
	ErrAlreadySubmitting = errors.New("already submitting")
	ErrInvalidForm       = errors.New("invalid form")
	ErrUserAlreadyExists = errors.New("user already exists")
)
