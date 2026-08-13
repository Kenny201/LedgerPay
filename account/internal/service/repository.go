package service

import (
	"context"

	"github.com/Kenny201/LedgerPay/account/internal/model"
)

// AccountRepository — то, что сервису нужно от хранилища счетов.
type AccountRepository interface {
	Create(ctx context.Context, account *model.Account) error
	GetByID(ctx context.Context, id int64) (*model.Account, error)
	GetByUserID(ctx context.Context, userID int64) (*model.Account, error)
}
