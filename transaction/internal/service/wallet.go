package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kenny201/LedgerPay/transaction/internal/domain"
	"github.com/Kenny201/LedgerPay/transaction/internal/model"
)

// Wallet — пополнение счёта.
type Wallet struct {
	repo WalletRepository
}

func NewWallet(repo WalletRepository) *Wallet {
	return &Wallet{repo: repo}
}

func (s *Wallet) TopUp(ctx context.Context, cmd TopUpCommand) (*model.TopUpResult, error) {
	cmd.IdempotencyKey = strings.TrimSpace(cmd.IdempotencyKey)
	if cmd.AccountID <= 0 || cmd.AmountCents <= 0 || cmd.IdempotencyKey == "" {
		return nil, fmt.Errorf("%w: account_id, amount_cents > 0 and Idempotency-Key required", domain.ErrInvalidInput)
	}
	return s.repo.TopUp(ctx, cmd)
}
