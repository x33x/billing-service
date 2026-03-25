package domain

import (
	"errors"
	"testing"
	"time"
)

// CanDebit
func TestCanDebit_Success(t *testing.T) {
	acc := Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  500,
		Status:   AccountStatusActive,
	}

	if !acc.CanDebit(100) {
		t.Error("expected true, got false")
	}
}

func TestCanDebit_InsufficientFunds(t *testing.T) {
	acc := Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  500,
		Status:   AccountStatusActive,
	}

	if acc.CanDebit(1000) {
		t.Error("expected false, got true")
	}
}

func TestCanDebit_AccountBlocked(t *testing.T) {
	acc := Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  500,
		Status:   AccountStatusBlocked,
	}

	if acc.CanDebit(100) {
		t.Error("expected false, got true")
	}
}

// Apply
func TestApply_SuccessDebit(t *testing.T) {
	acc := Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  500,
		Status:   AccountStatusActive,
	}

	tx := Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    100,
		Type:      TxTypeDebit,
		Status:    TxStatusPending,
		CreatedAt: time.Now(),
	}

	err := acc.Apply(tx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if acc.Balance != 400 {
		t.Errorf("expected balance=400, got=%d", acc.Balance)
	}
}

func TestApply_SuccessCredit(t *testing.T) {
	acc := Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  500,
		Status:   AccountStatusActive,
	}

	tx := Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    100,
		Type:      TxTypeCredit,
		Status:    TxStatusPending,
		CreatedAt: time.Now(),
	}

	err := acc.Apply(tx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if acc.Balance != 600 {
		t.Errorf("expected balance=600, got=%d", acc.Balance)
	}
}

func TestApply_InsufficientFunds(t *testing.T) {
	acc := Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  500,
		Status:   AccountStatusActive,
	}

	tx := Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    1000,
		Type:      TxTypeDebit,
		Status:    TxStatusPending,
		CreatedAt: time.Now(),
	}

	err := acc.Apply(tx)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}

	if acc.Balance != 500 {
		t.Errorf("expected balance=500, got=%d", acc.Balance)
	}
}

func TestApply_AccountBlocked(t *testing.T) {
	acc := Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  500,
		Status:   AccountStatusBlocked,
	}

	tx := Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    100,
		Type:      TxTypeDebit,
		Status:    TxStatusPending,
		CreatedAt: time.Now(),
	}

	err := acc.Apply(tx)
	if !errors.Is(err, ErrAccountBlocked) {
		t.Errorf("expected ErrAccountBlocked, got %v", err)
	}

}

func TestApply_AccountClosed(t *testing.T) {
	acc := Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  500,
		Status:   AccountStatusClosed,
	}

	tx := Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    100,
		Type:      TxTypeDebit,
		Status:    TxStatusPending,
		CreatedAt: time.Now(),
	}

	err := acc.Apply(tx)
	if !errors.Is(err, ErrAccountClosed) {
		t.Errorf("expected ErrAccountClosed, got %v", err)
	}

}

func TestApply_InvalidStatus(t *testing.T) {
	acc := Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  500,
		Status:   AccountStatusActive,
	}

	tx := Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    100,
		Type:      TxTypeDebit,
		Status:    TxStatusFailed,
		CreatedAt: time.Now(),
	}

	err := acc.Apply(tx)
	if !errors.Is(err, ErrTransactionMustBePending) {
		t.Errorf("expected ErrTransactionMustBePending, got %v", err)
	}

}

// Validate
func TestValidate_InvalidAmount(t *testing.T) {
	tx := Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    -100,
		Type:      TxTypeDebit,
		Status:    TxStatusPending,
		CreatedAt: time.Now(),
	}

	err := tx.Validate()
	if !errors.Is(err, ErrAmountMustBeGreaterThanZero) {
		t.Errorf("expected ErrAmountMustBeGreaterThanZero, got %v", err)
	}

}

func TestValidate_InvalidType(t *testing.T) {
	tx := Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    100,
		Type:      "None",
		Status:    TxStatusPending,
		CreatedAt: time.Now(),
	}

	err := tx.Validate()
	if !errors.Is(err, ErrUnknownTransactionType) {
		t.Errorf("expected ErrUnknownTransactionType, got %v", err)
	}

}

func TestValidate_EmptyAccountID(t *testing.T) {
	tx := Transaction{
		ID:        "tx-1",
		AccountID: "",
		Amount:    100,
		Type:      TxTypeDebit,
		Status:    TxStatusPending,
		CreatedAt: time.Now(),
	}

	err := tx.Validate()
	if !errors.Is(err, ErrAccountIdIsRequired) {
		t.Errorf("expected ErrAccountIdIsRequired, got %v", err)
	}

}
