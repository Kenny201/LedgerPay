package model

import "time"

// Account — денежный счёт пользователя (таблица accounts).
type Account struct {
	ID           int64
	UserID       int64
	Currency     string
	BalanceCents int64
	UpdatedAt    time.Time
}
