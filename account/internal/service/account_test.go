package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Kenny201/LedgerPay/account/internal/domain"
	"github.com/Kenny201/LedgerPay/account/internal/model"
	"github.com/Kenny201/LedgerPay/account/internal/service"
)

func TestAccount_CreateAndGet(t *testing.T) {
	t.Parallel()

	svc := service.NewAccount(newMemoryAccounts())
	acc, err := svc.Create(context.Background(), 42)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if acc.ID == 0 || acc.UserID != 42 || acc.BalanceCents != 0 || acc.Currency != "RUB" {
		t.Fatalf("Create: unexpected %+v", acc)
	}

	got, err := svc.GetByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if got.ID != acc.ID || got.BalanceCents != 0 {
		t.Fatalf("GetByUserID: got %+v", got)
	}
}

func TestAccount_CreateDuplicate(t *testing.T) {
	t.Parallel()

	svc := service.NewAccount(newMemoryAccounts())
	if _, err := svc.Create(context.Background(), 7); err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, err := svc.Create(context.Background(), 7)
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("Create dup: err=%v want %v", err, domain.ErrAlreadyExists)
	}
}

func TestAccount_GetNotFound(t *testing.T) {
	t.Parallel()

	svc := service.NewAccount(newMemoryAccounts())
	_, err := svc.GetByUserID(context.Background(), 99)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByUserID: err=%v want %v", err, domain.ErrNotFound)
	}
}

func TestAccount_InvalidUserID(t *testing.T) {
	t.Parallel()

	svc := service.NewAccount(newMemoryAccounts())
	_, err := svc.Create(context.Background(), 0)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("Create: err=%v want %v", err, domain.ErrInvalidInput)
	}
}

type memoryAccounts struct {
	mu     sync.Mutex
	byID   map[int64]*model.Account
	nextID int64
}

func newMemoryAccounts() *memoryAccounts {
	return &memoryAccounts{byID: make(map[int64]*model.Account), nextID: 1}
}

func (m *memoryAccounts) Create(_ context.Context, account *model.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.byID {
		if a.UserID == account.UserID {
			return domain.ErrAlreadyExists
		}
	}
	cp := *account
	cp.ID = m.nextID
	m.nextID++
	m.byID[cp.ID] = &cp
	account.ID = cp.ID
	account.Currency = cp.Currency
	return nil
}

func (m *memoryAccounts) GetByID(_ context.Context, id int64) (*model.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (m *memoryAccounts) GetByUserID(_ context.Context, userID int64) (*model.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.byID {
		if a.UserID == userID {
			cp := *a
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}
