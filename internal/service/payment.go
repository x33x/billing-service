package service

import (
	"context"
	"fmt"
	"time"

	"github.com/x33x/billing-service/internal/domain"
)

type AccountReader interface {
	GetByID(ctx context.Context, id string) (*domain.Account, error)
}

type TransactionWriter interface {
	Create(ctx context.Context, tx domain.Transaction) error
}

type TransactionReader interface {
	GetByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error)
}

type TransactionRepository interface {
	TransactionWriter
	TransactionReader
}

type PaymentService struct {
	accounts     AccountReader
	transactions TransactionRepository
	feeConfig    domain.FeeConfig
}

func NewPaymentService(
	accounts AccountReader,
	transactions TransactionRepository,
	feeConfig domain.FeeConfig,
) *PaymentService {
	return &PaymentService{
		accounts:     accounts,
		transactions: transactions,
		feeConfig:    feeConfig,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, tx domain.Transaction) error {
	tx.ID = fmt.Sprintf("tx_%d", time.Now().UnixNano())
	tx.Status = domain.TxStatusPending

	if tx.IdempotencyKey != nil {

		txExist, err := s.transactions.GetByIdempotencyKey(ctx, *tx.IdempotencyKey)

		if err != nil {
			return fmt.Errorf("ProcessPayment: check idempotency: %w", err)
		}

		if txExist != nil {
			return nil
		}
	}

	acc, err := s.accounts.GetByID(ctx, tx.AccountID)
	if err != nil {
		return fmt.Errorf("ProcessPayment: %w", err)
	}

	if err := acc.Apply(tx); err != nil {
		return fmt.Errorf("ProcessPayment: %w", err)
	}

	if err := s.transactions.Create(ctx, tx); err != nil {
		return fmt.Errorf("ProcessPayment: %w", err)
	}

	if tx.Type == domain.TxTypeDebit {
		if err := s.applyFee(ctx, tx); err != nil {
			return fmt.Errorf("ProcessPayment: %w", err)
		}
	}

	return nil
}

func (s *PaymentService) GetBalance(ctx context.Context, accountID string) (int64, error) {
	acc, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("GetBalance: %w", err)
	}

	return acc.Balance, nil
}

func (s *PaymentService) GetTransactions(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	txs, err := s.transactions.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("GetTransactions: %w", err)
	}

	return txs, nil
}

func (s *PaymentService) applyFee(ctx context.Context, original domain.Transaction) error {
	feeAmount := s.feeConfig.Calculate(original.Amount)

	if feeAmount == 0 {
		return nil
	}

	feeTx := domain.Transaction{
		ID:        "fee_" + original.ID,
		AccountID: original.AccountID,
		Amount:    feeAmount,
		Type:      domain.TxTypeFee,
		Status:    domain.TxStatusPending,
	}

	if err := s.transactions.Create(ctx, feeTx); err != nil {
		return fmt.Errorf("applyFee: %w", err)
	}

	return nil
}
