package middlewares

import "errors"

var (
	// ErrTokenNotExist is returned when the authorization token is missing from the request
	// or has an invalid format.
	ErrTokenNotExist = errors.New("token not exist")
)
