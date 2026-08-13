package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Kenny201/LedgerPay/auth/internal/domain"
	"github.com/Kenny201/LedgerPay/auth/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIsUniqueViolation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "other", err: errors.New("boom"), want: false},
		{name: "unique", err: &pgconn.PgError{Code: "23505"}, want: true},
		{name: "fk", err: &pgconn.PgError{Code: "23503"}, want: false},
		{
			name: "wrapped unique",
			err:  fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505"}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isUniqueViolation(tt.err); got != tt.want {
				t.Fatalf("isUniqueViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserPostgres_CreateAndFind(t *testing.T) {
	pool := openTestPool(t)
	repo := NewUserPostgres(pool)
	ctx := context.Background()

	login := fmt.Sprintf("user_%d", time.Now().UnixNano())
	email := login + "@example.com"

	u := &model.User{
		Login:        login,
		Email:        email,
		PasswordHash: "hash",
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("Create: expected non-zero id")
	}
	if u.CreatedAt.IsZero() {
		t.Fatal("Create: expected created_at")
	}
	t.Cleanup(func() { cleanupUser(pool, u.ID) })

	got, err := repo.FindByLogin(ctx, login)
	if err != nil {
		t.Fatalf("FindByLogin: %v", err)
	}
	if got.ID != u.ID || got.Email != email || got.PasswordHash != "hash" {
		t.Fatalf("FindByLogin: got %+v", got)
	}

	got, err = repo.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if got.ID != u.ID {
		t.Fatalf("FindByEmail: got id %d, want %d", got.ID, u.ID)
	}

	var accounts int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM accounts WHERE user_id = $1`, u.ID).Scan(&accounts); err != nil {
		t.Fatalf("count accounts: %v", err)
	}
	if accounts != 1 {
		t.Fatalf("register should create 1 account, got %d", accounts)
	}
}

func TestUserPostgres_CreateDuplicate(t *testing.T) {
	pool := openTestPool(t)
	repo := NewUserPostgres(pool)
	ctx := context.Background()

	login := fmt.Sprintf("dup_%d", time.Now().UnixNano())
	u := &model.User{Login: login, Email: login + "@example.com", PasswordHash: "h"}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { cleanupUser(pool, u.ID) })

	dup := &model.User{Login: login, Email: "other_" + login + "@example.com", PasswordHash: "h"}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("Create duplicate login: err = %v, want %v", err, domain.ErrAlreadyExists)
	}
}

func TestUserPostgres_FindNotFound(t *testing.T) {
	pool := openTestPool(t)
	repo := NewUserPostgres(pool)
	ctx := context.Background()

	_, err := repo.FindByLogin(ctx, "missing-user-that-does-not-exist")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("FindByLogin: err = %v, want %v", err, domain.ErrNotFound)
	}
}

func cleanupUser(pool *pgxpool.Pool, userID int64) {
	_, _ = pool.Exec(context.Background(), `DELETE FROM accounts WHERE user_id = $1`, userID)
	_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
}

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://ledger:ledger@localhost:5432/ledgerpay?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("БД недоступна (%v) — подними make up-infra && make migrate", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("БД недоступна (%v) — подними make up-infra && make migrate", err)
	}

	t.Cleanup(pool.Close)
	return pool
}
