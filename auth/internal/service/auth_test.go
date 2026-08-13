package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Kenny201/LedgerPay/auth/internal/domain"
	"github.com/Kenny201/LedgerPay/auth/internal/model"
	"github.com/Kenny201/LedgerPay/auth/internal/service"
)

func TestAuth_RegisterAndLogin(t *testing.T) {
	t.Parallel()

	repo := newMemoryUsers()
	auth := service.NewAuth(repo, "test-secret-at-least-32-characters!!")

	u, err := auth.Register(context.Background(), service.RegisterInput{
		Login:    "vitaly",
		Email:    "vitaly@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if u.ID == 0 || u.PasswordHash == "" || u.PasswordHash == "password123" {
		t.Fatalf("Register: unexpected user %+v", u)
	}

	tok, err := auth.Login(context.Background(), service.LoginInput{
		Login:    "vitaly",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if tok.AccessToken == "" || tok.UserID != u.ID {
		t.Fatalf("Login: unexpected token %+v", tok)
	}

	userID, err := auth.Validate(tok.AccessToken)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if userID != u.ID {
		t.Fatalf("Validate: userID=%d want %d", userID, u.ID)
	}
}

func TestAuth_LoginWrongPassword(t *testing.T) {
	t.Parallel()

	repo := newMemoryUsers()
	auth := service.NewAuth(repo, "test-secret-at-least-32-characters!!")
	_, err := auth.Register(context.Background(), service.RegisterInput{
		Login: "u1", Email: "u1@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, err = auth.Login(context.Background(), service.LoginInput{Login: "u1", Password: "wrong-pass"})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("Login: err=%v want %v", err, domain.ErrInvalidCredentials)
	}
}

func TestAuth_RegisterDuplicate(t *testing.T) {
	t.Parallel()

	repo := newMemoryUsers()
	auth := service.NewAuth(repo, "test-secret-at-least-32-characters!!")
	in := service.RegisterInput{Login: "u1", Email: "u1@example.com", Password: "password123"}
	if _, err := auth.Register(context.Background(), in); err != nil {
		t.Fatalf("Register: %v", err)
	}
	_, err := auth.Register(context.Background(), in)
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("Register dup: err=%v want %v", err, domain.ErrAlreadyExists)
	}
}

func TestAuth_RegisterShortPassword(t *testing.T) {
	t.Parallel()

	auth := service.NewAuth(newMemoryUsers(), "test-secret-at-least-32-characters!!")
	_, err := auth.Register(context.Background(), service.RegisterInput{
		Login: "u1", Email: "u1@example.com", Password: "short",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("Register: err=%v want %v", err, domain.ErrInvalidInput)
	}
}

type memoryUsers struct {
	mu     sync.Mutex
	byID   map[int64]*model.User
	nextID int64
}

func newMemoryUsers() *memoryUsers {
	return &memoryUsers{
		byID:   make(map[int64]*model.User),
		nextID: 1,
	}
}

func (m *memoryUsers) Create(_ context.Context, user *model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.byID {
		if u.Login == user.Login || u.Email == user.Email {
			return domain.ErrAlreadyExists
		}
	}
	cp := *user
	cp.ID = m.nextID
	m.nextID++
	m.byID[cp.ID] = &cp
	user.ID = cp.ID
	return nil
}

func (m *memoryUsers) FindByLogin(_ context.Context, login string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.byID {
		if u.Login == login {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *memoryUsers) FindByEmail(_ context.Context, email string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.byID {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}
