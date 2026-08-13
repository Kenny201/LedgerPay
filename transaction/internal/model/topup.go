package model

import "time"

type TopUp struct {
	ID             int64
	AccountID      int64
	AmountCents    int64
	IdempotencyKey string
	CreatedAt      time.Time
}

type TopUpResult struct {
	AccountID    int64
	BalanceCents int64
	Applied      bool
}
