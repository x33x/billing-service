package domain

import (
	"fmt"
	"time"
)

type Account struct {
	ID       string `json:"id"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance"`
	Status   string `json:"status"`
}

type Transaction struct {
	ID             string    `json:"id"`
	AccountID      string    `json:"account_id"`
	Amount         int64     `json:"amount"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	IdempotencyKey *string   `json:"idempotency_key,omitempty"` // may be nill (null in db) or has value
	CreatedAt      time.Time `json:"created_at"`
}

type LedgerEntry struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	AccountID     string    `json:"account_id"`
	Amount        int64     `json:"amount"`
	Direction     string    `json:"direction"`
	CreatedAt     time.Time `json:"created_at"`
}

type FeeConfig struct {
	Rate float64
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

const (
	LedgerDirectionDebit  = "debit"
	LedgerDirectionCredit = "credit"
)

const (
	TxTypeFee = "fee"
)

func (a *Account) CanDebit(amount int64) bool {
	if a.Status != AccountStatusActive {
		return false
	}

	return a.Balance >= amount
}

func (a *Account) Apply(tx Transaction) error {
	if a.Status == AccountStatusBlocked {
		return fmt.Errorf("Apply: %w", ErrAccountBlocked)
	}

	if a.Status == AccountStatusClosed {
		return fmt.Errorf("Apply: %w", ErrAccountClosed)
	}

	if tx.Status != TxStatusPending {
		return fmt.Errorf("Apply: %w", ErrTransactionMustBePending)
	}

	if tx.Type == TxTypeDebit && !a.CanDebit(tx.Amount) {
		return fmt.Errorf("Apply: insufficient funds (balance %d, amount %d): %w", a.Balance, tx.Amount, ErrInsufficientFunds)
	}

	if tx.Type == TxTypeDebit {
		a.Balance -= tx.Amount
	} else {
		a.Balance += tx.Amount
	}

	return nil
}

func (tx *Transaction) Validate() error {
	if tx.Amount <= 0 {
		return fmt.Errorf("Validate: %w", ErrAmountMustBeGreaterThanZero)
	}

	if tx.Type != TxTypeDebit && tx.Type != TxTypeCredit && tx.Type != TxTypeFee {
		return fmt.Errorf("Validate: unknown transaction type %q: %w", tx.Type, ErrUnknownTransactionType) // ("unknown type: %s", tx.Type)
	}

	if tx.AccountID == "" {
		return fmt.Errorf("Validate: %w", ErrAccountIdIsRequired)
	}

	return nil
}

func (f FeeConfig) Calculate(amount int64) int64 {
	return int64(float64(amount) * f.Rate)
}
