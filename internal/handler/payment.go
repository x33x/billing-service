package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/x33x/billing-service/internal/domain"
)

type PaymentService interface {
	ProcessPayment(ctx context.Context, tx domain.Transaction) error
	GetBalance(ctx context.Context, accountID string) (int64, error)
	GetTransactions(ctx context.Context, accountID string) ([]domain.Transaction, error)
}

type PaymentHandler struct {
	service PaymentService
}

func NewPaymentHandler(service PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

type CreatePaymentRequest struct {
	AccountID      string  `json:"account_id"`
	Amount         int64   `json:"amount"`
	Type           string  `json:"type"`
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest // create empty structure not pointer (on nil) to write into it
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	tx := domain.Transaction{
		AccountID:      req.AccountID,
		Amount:         req.Amount,
		Type:           req.Type,
		IdempotencyKey: req.IdempotencyKey,
	}

	if err := h.service.ProcessPayment(r.Context(), tx); err != nil {
		WriteJSON(w, errorToStatus(err), APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	WriteJSON(w, http.StatusCreated, APIResponse{Success: true})
}

func (h *PaymentHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")

	balance, err := h.service.GetBalance(r.Context(), accountID)
	if err != nil {
		WriteJSON(w, errorToStatus(err), APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	WriteJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]any{
			"balance": balance,
		},
	})
}

func (h *PaymentHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")

	txs, err := h.service.GetTransactions(r.Context(), accountID)
	if err != nil {
		WriteJSON(w, errorToStatus(err), APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	WriteJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    txs,
	})
}

func errorToStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return http.StatusNotFound // 404
	case errors.Is(err, domain.ErrInsufficientFunds),
		errors.Is(err, domain.ErrAccountBlocked),
		errors.Is(err, domain.ErrAccountClosed):
		return http.StatusUnprocessableEntity // 422
	case errors.Is(err, domain.ErrDuplicateTransaction):
		return http.StatusConflict // 409
	default:
		return http.StatusInternalServerError // 500
	}
}
