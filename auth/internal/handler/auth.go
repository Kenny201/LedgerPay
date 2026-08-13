package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Kenny201/LedgerPay/auth/internal/domain"
	"github.com/Kenny201/LedgerPay/auth/internal/service"
)

type Auth struct {
	auth *service.Auth
}

func NewAuth(auth *service.Auth) *Auth {
	return &Auth{auth: auth}
}

type registerRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	UserID      int64  `json:"user_id"`
	ExpiresAt   string `json:"expires_at"`
}

func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	u, err := h.auth.Register(r.Context(), service.RegisterInput{
		Login:    req.Login,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, userResponse{
		ID:    u.ID,
		Login: u.Login,
		Email: u.Email,
	})
}

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	tok, err := h.auth.Login(r.Context(), service.LoginInput{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken: tok.AccessToken,
		UserID:      tok.UserID,
		ExpiresAt:   tok.ExpiresAt.Format(time.RFC3339),
	})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "user already exists")
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
