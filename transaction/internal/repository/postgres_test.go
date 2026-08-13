package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Kenny201/LedgerPay/transaction/internal/domain"
	"github.com/Kenny201/LedgerPay/transaction/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestWalletPostgres_TopUpIdempotent(t *testing.T) {
	pool := openTestPool(t)
	repo := NewWalletPostgres(pool)
	ctx := context.Background()

	accountID := insertTestAccount(t, pool)
	cmd := service.TopUpCommand{
		AccountID:      accountID,
		AmountCents:    1500,
		IdempotencyKey: fmt.Sprintf("itest-%d", time.Now().UnixNano()),
	}

	first, err := repo.TopUp(ctx, cmd)
	if err != nil {
		t.Fatalf("first TopUp: %v", err)
	}
	if !first.Applied || first.BalanceCents != 1500 {
		t.Fatalf("first: %+v", first)
	}

	second, err := repo.TopUp(ctx, cmd)
	if err != nil {
		t.Fatalf("second TopUp: %v", err)
	}
	if second.Applied || second.BalanceCents != 1500 {
		t.Fatalf("replay must not add twice: %+v", second)
	}

	var balance int64
	if err := pool.QueryRow(ctx, `SELECT balance_cents FROM accounts WHERE id = $1`, accountID).Scan(&balance); err != nil {
		t.Fatalf("select balance: %v", err)
	}
	if balance != 1500 {
		t.Fatalf("stored balance=%d want 1500", balance)
	}

	var ledgerCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM ledger_entries WHERE account_id = $1 AND kind = 'topup'`, accountID,
	).Scan(&ledgerCount); err != nil {
		t.Fatalf("count ledger: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("ledger entries=%d want 1", ledgerCount)
	}
}

func TestWalletPostgres_TopUpConflict(t *testing.T) {
	pool := openTestPool(t)
	repo := NewWalletPostgres(pool)
	ctx := context.Background()

	accountID := insertTestAccount(t, pool)
	key := fmt.Sprintf("conflict-%d", time.Now().UnixNano())
	cmd := service.TopUpCommand{AccountID: accountID, AmountCents: 100, IdempotencyKey: key}
	if _, err := repo.TopUp(ctx, cmd); err != nil {
		t.Fatalf("first: %v", err)
	}
	cmd.AmountCents = 999
	_, err := repo.TopUp(ctx, cmd)
	if !errors.Is(err, domain.ErrIdempotencyConflict) {
		t.Fatalf("err=%v want %v", err, domain.ErrIdempotencyConflict)
	}
}

func TestWalletPostgres_AccountNotFound(t *testing.T) {
	pool := openTestPool(t)
	repo := NewWalletPostgres(pool)

	_, err := repo.TopUp(context.Background(), service.TopUpCommand{
		AccountID:      9_999_999_999,
		AmountCents:    100,
		IdempotencyKey: fmt.Sprintf("missing-%d", time.Now().UnixNano()),
	})
	if !errors.Is(err, domain.ErrAccountNotFound) {
		t.Fatalf("err=%v want %v", err, domain.ErrAccountNotFound)
	}
}

func insertTestAccount(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	login := fmt.Sprintf("tx_user_%d", time.Now().UnixNano())
	var userID, accountID int64
	ctx := context.Background()
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (login, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		login, login+"@example.com", "hash",
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO accounts (user_id, currency, balance_cents) VALUES ($1, 'RUB', 0) RETURNING id`,
		userID,
	).Scan(&accountID); err != nil {
		t.Fatalf("insert account: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM ledger_entries WHERE account_id = $1`, accountID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM topups WHERE account_id = $1`, accountID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM accounts WHERE id = $1`, accountID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})
	return accountID
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
