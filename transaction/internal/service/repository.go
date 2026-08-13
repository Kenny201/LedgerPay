package service

import (
	"context"

	"github.com/Kenny201/LedgerPay/transaction/internal/model"
)

type TopUpCommand struct {
	AccountID      int64
	AmountCents    int64
	IdempotencyKey string
}

// WalletRepository — то, что сервису нужно для пополнения.
type WalletRepository interface {
	TopUp(ctx context.Context, cmd TopUpCommand) (*model.TopUpResult, error)
}
