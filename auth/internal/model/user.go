package model

import "time"

// User — пользователь сервиса auth (таблица users).
type User struct {
	ID           int64
	Login        string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
