package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kenny201/LedgerPay/transaction/internal/domain"
	"github.com/Kenny201/LedgerPay/transaction/internal/model"
	"github.com/Kenny201/LedgerPay/transaction/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WalletPostgres — пополнение в одной DB-транзакции.
type WalletPostgres struct {
	pool *pgxpool.Pool
}

func NewWalletPostgres(pool *pgxpool.Pool) *WalletPostgres {
	return &WalletPostgres{pool: pool}
}

func (r *WalletPostgres) TopUp(ctx context.Context, cmd service.TopUpCommand) (*model.TopUpResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var topupID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO topups (account_id, amount_cents, idempotency_key)
		VALUES ($1, $2, $3)
		RETURNING id`,
		cmd.AccountID, cmd.AmountCents, cmd.IdempotencyKey,
	).Scan(&topupID)

	if isUniqueViolation(err) {
		res, replayErr := r.replayTopUp(ctx, tx, cmd)
		if replayErr != nil {
			return nil, replayErr
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit replay: %w", err)
		}
		return res, nil
	}
	if err != nil {
		return nil, fmt.Errorf("insert topup: %w", err)
	}

	var balance int64
	err = tx.QueryRow(ctx, `
		SELECT balance_cents FROM accounts WHERE id = $1 FOR UPDATE`,
		cmd.AccountID,
	).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("lock account: %w", err)
	}

	balance += cmd.AmountCents
	if _, err := tx.Exec(ctx, `
		UPDATE accounts SET balance_cents = $2, updated_at = now() WHERE id = $1`,
		cmd.AccountID, balance,
	); err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO ledger_entries (account_id, kind, amount_cents)
		VALUES ($1, 'topup', $2)`,
		cmd.AccountID, cmd.AmountCents,
	); err != nil {
		return nil, fmt.Errorf("insert ledger: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit topup: %w", err)
	}

	return &model.TopUpResult{
		AccountID:    cmd.AccountID,
		BalanceCents: balance,
		Applied:      true,
	}, nil
}

func (r *WalletPostgres) replayTopUp(ctx context.Context, tx pgx.Tx, cmd service.TopUpCommand) (*model.TopUpResult, error) {
	var prevAmount int64
	err := tx.QueryRow(ctx, `
		SELECT amount_cents FROM topups
		WHERE account_id = $1 AND idempotency_key = $2`,
		cmd.AccountID, cmd.IdempotencyKey,
	).Scan(&prevAmount)
	if err != nil {
		return nil, fmt.Errorf("select existing topup: %w", err)
	}
	if prevAmount != cmd.AmountCents {
		return nil, domain.ErrIdempotencyConflict
	}

	var balance int64
	err = tx.QueryRow(ctx, `SELECT balance_cents FROM accounts WHERE id = $1`, cmd.AccountID).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("select balance: %w", err)
	}

	return &model.TopUpResult{
		AccountID:    cmd.AccountID,
		BalanceCents: balance,
		Applied:      false,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
