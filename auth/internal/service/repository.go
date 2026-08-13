package service

import (
	"context"

	"github.com/Kenny201/LedgerPay/auth/internal/model"
)

// UserRepository — то, что сервису нужно от хранилища пользователей.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByLogin(ctx context.Context, login string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}
