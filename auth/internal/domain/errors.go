package domain

import "errors"

var (
	ErrNotFound           = errors.New("user not found")
	ErrAlreadyExists      = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
)
