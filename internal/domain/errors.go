package domain

import "errors"

var (
	// ErrInvalidInput is returned when a request payload is invalid.
	ErrInvalidInput = errors.New("invalid input")

	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")
)
