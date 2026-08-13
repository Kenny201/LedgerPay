package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Kenny201/LedgerPay/account/internal/domain"
	"github.com/Kenny201/LedgerPay/account/internal/model"

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

func TestAccountPostgres_CreateAndGet(t *testing.T) {
	pool := openTestPool(t)
	repo := NewAccountPostgres(pool)
	ctx := context.Background()

	userID := insertTestUser(t, pool)
	acc := &model.Account{UserID: userID, BalanceCents: 0}
	if err := repo.Create(ctx, acc); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if acc.ID == 0 {
		t.Fatal("Create: expected non-zero id")
	}
	if acc.Currency != "RUB" {
		t.Fatalf("Create: currency = %q, want RUB", acc.Currency)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM accounts WHERE id = $1`, acc.ID)
	})

	byID, err := repo.GetByID(ctx, acc.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if byID.UserID != userID || byID.BalanceCents != 0 {
		t.Fatalf("GetByID: got %+v", byID)
	}

	byUser, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if byUser.ID != acc.ID {
		t.Fatalf("GetByUserID: id = %d, want %d", byUser.ID, acc.ID)
	}
}

func TestAccountPostgres_CreateDuplicateUser(t *testing.T) {
	pool := openTestPool(t)
	repo := NewAccountPostgres(pool)
	ctx := context.Background()

	userID := insertTestUser(t, pool)
	acc := &model.Account{UserID: userID}
	if err := repo.Create(ctx, acc); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM accounts WHERE id = $1`, acc.ID)
	})

	dup := &model.Account{UserID: userID}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("Create duplicate: err = %v, want %v", err, domain.ErrAlreadyExists)
	}
}

func TestAccountPostgres_GetNotFound(t *testing.T) {
	pool := openTestPool(t)
	repo := NewAccountPostgres(pool)

	_, err := repo.GetByID(context.Background(), 9_999_999_999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByID: err = %v, want %v", err, domain.ErrNotFound)
	}
}

func insertTestUser(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	login := fmt.Sprintf("acc_user_%d", time.Now().UnixNano())
	var id int64
	err := pool.QueryRow(
		context.Background(),
		`INSERT INTO users (login, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		login, login+"@example.com", "hash",
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert test user: %v (нужна таблица users — make migrate-auth)", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})
	return id
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
