package domain

import (
	"fmt"
	"time"
)

type Account struct {
	ID       string
	Currency string
	Balance  int64
	Status   string
}

type Transaction struct {
	ID        string
	AccountID string
	Amount    int64
	Type      string
	Status    string
	CreatedAt time.Time
}

func (a *Account) CanDebit(amount int64) bool {
	if a.Status != AccountStatusActive {
		return false
	}

	return a.Balance >= amount
}

func (a *Account) Apply(tx Transaction) error {
	if a.Status == AccountStatusBlocked || a.Status == AccountStatusClosed {
		return fmt.Errorf("account is %s", a.Status)
	}

	if tx.Status != TxStatusPending {
		return fmt.Errorf("transaction must be pending")
	}

	if tx.Type == TxTypeDebit && !a.CanDebit(tx.Amount) {
		return fmt.Errorf("insufficient funds: balance %d, amount %d", a.Balance, tx.Amount)
	}

	if tx.Type == TxTypeDebit {
		a.Balance = a.Balance - tx.Amount
	} else {
		a.Balance = a.Balance + tx.Amount
	}

	return nil
}

func (tx *Transaction) Validate() error {
	if tx.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	if tx.Type != TxTypeDebit && tx.Type != TxTypeCredit {
		return fmt.Errorf("unknown type: %s", tx.Type)
	}

	if tx.AccountID == "" {
		return fmt.Errorf("accountID is required")
	}

	return nil
}

const (
	AccountStatusActive  = "active"
	AccountStatusBlocked = "blocked"
	AccountStatusClosed  = "closed"

	TxTypeDebit  = "debit"
	TxTypeCredit = "credit"

	TxStatusPending   = "pending"
	TxStatusCompleted = "completed"
	TxStatusFailed    = "failed"
)
