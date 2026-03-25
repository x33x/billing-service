package service

import "github.com/x33x/billing-service/internal/domain"

type PaymentProcessor interface {
	Process(tx domain.Transaction) error
	GetBalance(accountID string) (int64, error)
}
