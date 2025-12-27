package api

import (
	"fmt"
)

type UnexpectedHTTPStatusCodeError struct {
	StatusCode int
	Message    string
}

func NewUnexpectedHTTPStatusCodeError(statusCode int, message string) *UnexpectedHTTPStatusCodeError {
	return &UnexpectedHTTPStatusCodeError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e *UnexpectedHTTPStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected http status code: %d, message: %s", e.StatusCode, e.Message)
}
