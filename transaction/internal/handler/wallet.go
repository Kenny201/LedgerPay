package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Kenny201/LedgerPay/transaction/internal/domain"
	"github.com/Kenny201/LedgerPay/transaction/internal/service"
)

type Wallet struct {
	wallet *service.Wallet
}

func NewWallet(wallet *service.Wallet) *Wallet {
	return &Wallet{wallet: wallet}
}

type topUpRequest struct {
	AccountID   int64 `json:"account_id"`
	AmountCents int64 `json:"amount_cents"`
}

type topUpResponse struct {
	AccountID    int64 `json:"account_id"`
	BalanceCents int64 `json:"balance_cents"`
	Applied      bool  `json:"applied"`
}

func (h *Wallet) TopUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req topUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	res, err := h.wallet.TopUp(r.Context(), service.TopUpCommand{
		AccountID:      req.AccountID,
		AmountCents:    req.AmountCents,
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, topUpResponse{
		AccountID:    res.AccountID,
		BalanceCents: res.BalanceCents,
		Applied:      res.Applied,
	})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrAccountNotFound):
		writeError(w, http.StatusNotFound, "account not found")
	case errors.Is(err, domain.ErrIdempotencyConflict):
		writeError(w, http.StatusConflict, "idempotency key reused with different amount")
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
