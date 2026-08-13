package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kenny201/LedgerPay/auth/internal/domain"
	"github.com/Kenny201/LedgerPay/auth/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserPostgres — реализация UserRepository на PostgreSQL.
type UserPostgres struct {
	pool *pgxpool.Pool
}

func NewUserPostgres(pool *pgxpool.Pool) *UserPostgres {
	return &UserPostgres{pool: pool}
}

func (r *UserPostgres) Create(ctx context.Context, user *model.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertUser = `
		INSERT INTO users (login, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err = tx.QueryRow(ctx, insertUser, user.Login, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	const insertAccount = `
		INSERT INTO accounts (user_id, currency, balance_cents)
		VALUES ($1, 'RUB', 0)`
	if _, err := tx.Exec(ctx, insertAccount, user.ID); err != nil {
		return fmt.Errorf("insert account: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit register: %w", err)
	}
	return nil
}

func (r *UserPostgres) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	return r.findOne(ctx, `
		SELECT id, login, email, password_hash, created_at
		FROM users
		WHERE login = $1`, login)
}

func (r *UserPostgres) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.findOne(ctx, `
		SELECT id, login, email, password_hash, created_at
		FROM users
		WHERE email = $1`, email)
}

func (r *UserPostgres) findOne(ctx context.Context, q string, arg any) (*model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx, q, arg).
		Scan(&u.ID, &u.Login, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select user: %w", err)
	}
	return &u, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
