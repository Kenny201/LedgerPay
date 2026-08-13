package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kenny201/LedgerPay/account/internal/domain"
	"github.com/Kenny201/LedgerPay/account/internal/model"
)

// Account — создание счёта и чтение баланса.
type Account struct {
	repo AccountRepository
}

func NewAccount(repo AccountRepository) *Account {
	return &Account{repo: repo}
}

func (s *Account) Create(ctx context.Context, userID int64) (*model.Account, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: user_id required", domain.ErrInvalidInput)
	}

	acc := &model.Account{
		UserID:       userID,
		Currency:     "RUB",
		BalanceCents: 0,
	}
	if err := s.repo.Create(ctx, acc); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil, domain.ErrAlreadyExists
		}
		return nil, fmt.Errorf("create account: %w", err)
	}
	return acc, nil
}

func (s *Account) GetByUserID(ctx context.Context, userID int64) (*model.Account, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: user_id required", domain.ErrInvalidInput)
	}

	acc, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get account: %w", err)
	}
	return acc, nil
}
