package domain

import "errors"

var (
	ErrAccountNotFound             = errors.New("account not found")
	ErrDuplicateTransaction        = errors.New("duplicate transaction")
	ErrInsufficientFunds           = errors.New("insufficient funds")
	ErrAccountBlocked              = errors.New("account is blocked")
	ErrAccountClosed               = errors.New("account is closed")
	ErrTransactionMustBePending    = errors.New("transaction must be pending")
	ErrAmountMustBeGreaterThanZero = errors.New("amount must be greater than zero")
	ErrAccountIdIsRequired         = errors.New("accountID is required")
	ErrUnknownTransactionType      = errors.New("unknown transaction type")
)
