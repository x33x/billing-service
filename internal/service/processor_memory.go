package service

import (
	"fmt"

	"github.com/x33x/billing-service/internal/domain"
)

type MemoryPaymentProcessor struct {
	accounts     map[string]*domain.Account
	transactions map[string]struct{}
}

// var _ PaymentProcessor = (*MemoryPaymentProcessor)(nil) // test interface

func NewMemoryPaymentProcessor(accounts []*domain.Account) *MemoryPaymentProcessor {
	m := map[string]*domain.Account{}

	for _, acc := range accounts {
		m[acc.ID] = acc
	}

	return &MemoryPaymentProcessor{
		accounts:     m,
		transactions: make(map[string]struct{}),
	}
}

func (p *MemoryPaymentProcessor) Process(tx domain.Transaction) error {
	if err := tx.Validate(); err != nil {
		return err // if validate is failed then return err
	}

	if _, exists := p.transactions[tx.ID]; exists {
		return fmt.Errorf("duplicate transaction: %s", tx.ID)
	}

	acc, ok := p.accounts[tx.AccountID] // search in map (acc is value, ok - bool (true - index found, false - index NOT found))
	if !ok {                            // if not found
		return fmt.Errorf("account not found: %s", tx.AccountID)
	}

	// exec method Apply on real account
	if err := acc.Apply(tx); err != nil {
		return err
	}

	// check transaction as applied
	p.transactions[tx.ID] = struct{}{}

	return nil
}

func (p *MemoryPaymentProcessor) GetBalance(accountID string) (int64, error) {
	acc, ok := p.accounts[accountID] // search in map (acc is value, ok - bool (true - index found, false - index NOT found))

	if !ok { // if not found
		return 0, fmt.Errorf("account not found: %s", accountID)
	}

	return acc.Balance, nil // if account found then return balance and error=nil
}
