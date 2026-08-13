package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Kenny201/LedgerPay/auth/internal/domain"
	"github.com/Kenny201/LedgerPay/auth/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLen = 8
	tokenTTL       = 24 * time.Hour
	bcryptCost     = bcrypt.DefaultCost
)

// Auth — регистрация и выдача JWT.
type Auth struct {
	users     UserRepository
	jwtSecret []byte
}

func NewAuth(users UserRepository, jwtSecret string) *Auth {
	return &Auth{
		users:     users,
		jwtSecret: []byte(jwtSecret),
	}
}

type RegisterInput struct {
	Login    string
	Email    string
	Password string
}

type LoginInput struct {
	Login    string
	Password string
}

type TokenResult struct {
	AccessToken string
	UserID      int64
	ExpiresAt   time.Time
}

func (s *Auth) Register(ctx context.Context, in RegisterInput) (*model.User, error) {
	login := strings.TrimSpace(in.Login)
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if login == "" || email == "" || len(in.Password) < minPasswordLen {
		return nil, fmt.Errorf("%w: login/email required, password min %d chars", domain.ErrInvalidInput, minPasswordLen)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &model.User{
		Login:        login,
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.users.Create(ctx, u); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil, domain.ErrAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (s *Auth) Login(ctx context.Context, in LoginInput) (*TokenResult, error) {
	login := strings.TrimSpace(in.Login)
	if login == "" || in.Password == "" {
		return nil, domain.ErrInvalidInput
	}

	u, err := s.users.FindByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	expiresAt := time.Now().UTC().Add(tokenTTL)
	token, err := s.issueToken(u.ID, u.Login, expiresAt)
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken: token,
		UserID:      u.ID,
		ExpiresAt:   expiresAt,
	}, nil
}

// Validate разбирает access token и возвращает user_id (для других сервисов позже).
func (s *Auth) Validate(accessToken string) (int64, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("%w: unexpected signing method", domain.ErrInvalidCredentials)
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, domain.ErrInvalidCredentials
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return 0, domain.ErrInvalidCredentials
	}

	var userID int64
	if _, err := fmt.Sscan(sub, &userID); err != nil || userID <= 0 {
		return 0, domain.ErrInvalidCredentials
	}
	return userID, nil
}

func (s *Auth) issueToken(userID int64, login string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", userID),
		"login": login,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().UTC().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}
