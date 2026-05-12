package service

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
)

type FieldError struct {
	Err     error
	Message string
	Field   string
}

func (e FieldError) Error() string {
	return e.Message
}

func (e FieldError) Unwrap() error {
	return e.Err
}

func ValidationError(field string, message string) FieldError {
	return FieldError{Err: ErrInvalidInput, Message: message, Field: field}
}
