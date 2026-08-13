package domain

import "errors"

var (
	ErrNotFound      = errors.New("account not found")
	ErrAlreadyExists = errors.New("account already exists")
	ErrInvalidInput  = errors.New("invalid input")
)
