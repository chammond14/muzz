package server

import "errors"

var (
	ErrInvalidRequest  = errors.New("invalid request")
	ErrMustBeLoggedIn  = errors.New("please log in to your account")
	ErrValidationError = errors.New("request body contained unexpected values")
	ErrUnexpectedError = errors.New("unexpected error occurred")
)
