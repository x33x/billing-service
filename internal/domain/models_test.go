package domain

import (
	"testing"
	"time"
)

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

	if err == nil {
		t.Error("expected error for empty AccountID, got nil")
	}
}
