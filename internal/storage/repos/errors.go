package repos

import "errors"

var (
	ErrRecordNotFound    = errors.New("record not found")
	ErrPasswordMissMatch = errors.New("password mismatch")
	ErrDuplicateKey      = errors.New("duplicate key")
	ErrIsolation         = errors.New("isolation error")
	ErrUnknown           = errors.New("unknown error")
)
