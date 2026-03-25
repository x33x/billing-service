package service

import (
	"testing"

	"github.com/x33x/billing-service/internal/domain"
)

func TestGetBalance_Exists(t *testing.T) {
	// 1. prepare data
	acc := &domain.Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  1000,
		Status:   domain.AccountStatusActive,
	}

	// 2. create processor with account
	processor := NewMemoryPaymentProcessor([]*domain.Account{acc})

	// 3. exec method GetBalance on account
	balance, err := processor.GetBalance("acc-1")

	// 4. check result
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if balance != 1000 {
		t.Errorf("expected 1000, got %d", balance)
	}
}

func TestProcess_SuccessDebit(t *testing.T) {
	acc := &domain.Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  1000,
		Status:   domain.AccountStatusActive,
	}

	processor := NewMemoryPaymentProcessor([]*domain.Account{acc})

	tx := domain.Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    300,
		Type:      domain.TxTypeDebit,
		Status:    domain.TxStatusPending,
	}

	if err := processor.Process(tx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if acc.Balance != 700 {
		t.Errorf("expected 700, got %d", acc.Balance)
	}
}

func TestProcess_AccountNotFound(t *testing.T) {
	processor := NewMemoryPaymentProcessor([]*domain.Account{})

	tx := domain.Transaction{
		ID:        "tx-1",
		AccountID: "acc-999",
		Amount:    100,
		Type:      domain.TxTypeDebit,
		Status:    domain.TxStatusPending,
	}

	if err := processor.Process(tx); err == nil {
		t.Error("expected error for empty AccountID, got nil")
	}
}

func TestProcess_DuplicateTransaction(t *testing.T) {
	acc := &domain.Account{
		ID:       "acc-1",
		Currency: "RUB",
		Balance:  1000,
		Status:   domain.AccountStatusActive,
	}

	processor := NewMemoryPaymentProcessor([]*domain.Account{acc})

	tx := domain.Transaction{
		ID:        "tx-1",
		AccountID: "acc-1",
		Amount:    100,
		Type:      domain.TxTypeDebit,
		Status:    domain.TxStatusPending,
	}

	if err := processor.Process(tx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := processor.Process(tx); err == nil {
		t.Error("expected error for duplicate tx-1, got nil")
	} else if err.Error() != "duplicate transaction: tx-1" {
		t.Errorf("unexpected error message: %v", err)
	}

}
