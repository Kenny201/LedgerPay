package domain

import "errors"

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrInvalidInput        = errors.New("invalid input")
	ErrIdempotencyConflict = errors.New("idempotency key reused with different payload")
)
