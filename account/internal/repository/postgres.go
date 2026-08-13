package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kenny201/LedgerPay/account/internal/domain"
	"github.com/Kenny201/LedgerPay/account/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AccountPostgres — реализация AccountRepository на PostgreSQL.
type AccountPostgres struct {
	pool *pgxpool.Pool
}

func NewAccountPostgres(pool *pgxpool.Pool) *AccountPostgres {
	return &AccountPostgres{pool: pool}
}

func (r *AccountPostgres) Create(ctx context.Context, account *model.Account) error {
	currency := account.Currency
	if currency == "" {
		currency = "RUB"
	}

	const q = `
		INSERT INTO accounts (user_id, currency, balance_cents)
		VALUES ($1, $2, $3)
		RETURNING id, currency, balance_cents, updated_at`

	err := r.pool.QueryRow(ctx, q, account.UserID, currency, account.BalanceCents).
		Scan(&account.ID, &account.Currency, &account.BalanceCents, &account.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

func (r *AccountPostgres) GetByID(ctx context.Context, id int64) (*model.Account, error) {
	return r.findOne(ctx, `
		SELECT id, user_id, currency, balance_cents, updated_at
		FROM accounts
		WHERE id = $1`, id)
}

func (r *AccountPostgres) GetByUserID(ctx context.Context, userID int64) (*model.Account, error) {
	return r.findOne(ctx, `
		SELECT id, user_id, currency, balance_cents, updated_at
		FROM accounts
		WHERE user_id = $1`, userID)
}

func (r *AccountPostgres) findOne(ctx context.Context, q string, arg int64) (*model.Account, error) {
	var a model.Account
	err := r.pool.QueryRow(ctx, q, arg).
		Scan(&a.ID, &a.UserID, &a.Currency, &a.BalanceCents, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select account: %w", err)
	}
	return &a, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
