package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/Kenny201/LedgerPay/transaction/internal/domain"
	"github.com/Kenny201/LedgerPay/transaction/internal/model"
	"github.com/Kenny201/LedgerPay/transaction/internal/service"
)

func TestWallet_TopUpIdempotent(t *testing.T) {
	t.Parallel()

	repo := newMemoryWallet(1, 0)
	svc := service.NewWallet(repo)
	cmd := service.TopUpCommand{AccountID: 1, AmountCents: 1000, IdempotencyKey: "topup-1"}

	first, err := svc.TopUp(context.Background(), cmd)
	if err != nil {
		t.Fatalf("first TopUp: %v", err)
	}
	if !first.Applied || first.BalanceCents != 1000 {
		t.Fatalf("first: %+v", first)
	}

	second, err := svc.TopUp(context.Background(), cmd)
	if err != nil {
		t.Fatalf("second TopUp: %v", err)
	}
	if second.Applied || second.BalanceCents != 1000 {
		t.Fatalf("second must not apply again: %+v", second)
	}
}

func TestWallet_TopUpDifferentKeysAddUp(t *testing.T) {
	t.Parallel()

	svc := service.NewWallet(newMemoryWallet(1, 0))
	if _, err := svc.TopUp(context.Background(), service.TopUpCommand{AccountID: 1, AmountCents: 500, IdempotencyKey: "a"}); err != nil {
		t.Fatalf("first: %v", err)
	}
	got, err := svc.TopUp(context.Background(), service.TopUpCommand{AccountID: 1, AmountCents: 700, IdempotencyKey: "b"})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if got.BalanceCents != 1200 {
		t.Fatalf("balance=%d want 1200", got.BalanceCents)
	}
}

func TestWallet_TopUpConflict(t *testing.T) {
	t.Parallel()

	svc := service.NewWallet(newMemoryWallet(1, 0))
	cmd := service.TopUpCommand{AccountID: 1, AmountCents: 100, IdempotencyKey: "same"}
	if _, err := svc.TopUp(context.Background(), cmd); err != nil {
		t.Fatalf("first: %v", err)
	}
	cmd.AmountCents = 200
	_, err := svc.TopUp(context.Background(), cmd)
	if !errors.Is(err, domain.ErrIdempotencyConflict) {
		t.Fatalf("err=%v want %v", err, domain.ErrIdempotencyConflict)
	}
}

func TestWallet_InvalidInput(t *testing.T) {
	t.Parallel()

	svc := service.NewWallet(newMemoryWallet(1, 0))
	_, err := svc.TopUp(context.Background(), service.TopUpCommand{AccountID: 1, AmountCents: 0, IdempotencyKey: "k"})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err=%v want %v", err, domain.ErrInvalidInput)
	}
}

type memoryWallet struct {
	mu       sync.Mutex
	balances map[int64]int64
	applied  map[string]int64 // key accountID|idem -> amount
}

func newMemoryWallet(accountID, balance int64) *memoryWallet {
	return &memoryWallet{
		balances: map[int64]int64{accountID: balance},
		applied:  make(map[string]int64),
	}
}

func (m *memoryWallet) TopUp(_ context.Context, cmd service.TopUpCommand) (*model.TopUpResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.balances[cmd.AccountID]; !ok {
		return nil, domain.ErrAccountNotFound
	}

	key := fmtKey(cmd.AccountID, cmd.IdempotencyKey)
	if prev, ok := m.applied[key]; ok {
		if prev != cmd.AmountCents {
			return nil, domain.ErrIdempotencyConflict
		}
		return &model.TopUpResult{AccountID: cmd.AccountID, BalanceCents: m.balances[cmd.AccountID], Applied: false}, nil
	}

	m.applied[key] = cmd.AmountCents
	m.balances[cmd.AccountID] += cmd.AmountCents
	return &model.TopUpResult{AccountID: cmd.AccountID, BalanceCents: m.balances[cmd.AccountID], Applied: true}, nil
}

func fmtKey(accountID int64, idem string) string {
	return fmt.Sprintf("%d|%s", accountID, idem)
}
