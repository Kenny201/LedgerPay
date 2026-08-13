package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Kenny201/LedgerPay/account/internal/domain"
	"github.com/Kenny201/LedgerPay/account/internal/model"
	"github.com/Kenny201/LedgerPay/account/internal/service"
)

type Account struct {
	accounts *service.Account
}

func NewAccount(accounts *service.Account) *Account {
	return &Account{accounts: accounts}
}

type createRequest struct {
	UserID int64 `json:"user_id"`
}

type accountResponse struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	Currency     string `json:"currency"`
	BalanceCents int64  `json:"balance_cents"`
}

func (h *Account) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	acc, err := h.accounts.Create(r.Context(), req.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(acc))
}

func (h *Account) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		writeError(w, http.StatusBadRequest, "user_id query required")
		return
	}

	acc, err := h.accounts.GetByUserID(r.Context(), userID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(acc))
}

func toResponse(acc *model.Account) accountResponse {
	return accountResponse{
		ID:           acc.ID,
		UserID:       acc.UserID,
		Currency:     acc.Currency,
		BalanceCents: acc.BalanceCents,
	}
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "account already exists")
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "account not found")
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
