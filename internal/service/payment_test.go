package service

import (
	"context"
	"errors"
	"testing"

	"github.com/x33x/billing-service/internal/domain"
)

type mockAccountRepo struct {
	accounts map[string]*domain.Account
}

type mockTxRepo struct {
	byIdempotencyKey map[string]*domain.Transaction
	created          []domain.Transaction
	createErr        error
}

func (m *mockAccountRepo) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	acc, ok := m.accounts[id]
	if !ok {
		return nil, domain.ErrAccountNotFound
	}

	return acc, nil
}

func (m *mockTxRepo) Create(ctx context.Context, tx domain.Transaction) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.created = append(m.created, tx)

	return nil
}

func (m *mockTxRepo) GetByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	return nil, nil
}

func (m *mockTxRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	tx, ok := m.byIdempotencyKey[key]
	if !ok {
		return nil, nil
	}

	return tx, nil
}

func newService(acc *domain.Account, feeRate float64) (*PaymentService, *mockTxRepo) {
	accRepo := &mockAccountRepo{
		accounts: map[string]*domain.Account{acc.ID: acc},
	}
	txRepo := &mockTxRepo{
		byIdempotencyKey: make(map[string]*domain.Transaction),
	}
	svc := NewPaymentService(accRepo, txRepo, domain.FeeConfig{Rate: feeRate})

	return svc, txRepo
}

func TestProcessPayment_SuccessDebit(t *testing.T) {
	acc := &domain.Account{
		ID:      "acc-1",
		Balance: 10000,
		Status:  domain.AccountStatusActive,
	}

	svc, txRepo := newService(acc, 0.015)

	tx := domain.Transaction{
		AccountID: "acc-1",
		Amount:    3000,
		Type:      domain.TxTypeDebit,
	}

	err := svc.ProcessPayment(context.Background(), tx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(txRepo.created) != 2 {
		t.Errorf("expected 2 transactions (payment + fee), got %d", len(txRepo.created))
	}
}

func TestProcessPayment_InsufficientFunds(t *testing.T) {
	acc := &domain.Account{
		ID:      "acc-1",
		Balance: 1000,
		Status:  domain.AccountStatusActive,
	}

	svc, _ := newService(acc, 0.015)

	tx := domain.Transaction{
		AccountID: "acc-1",
		Amount:    5000,
		Type:      domain.TxTypeDebit,
	}

	err := svc.ProcessPayment(context.Background(), tx)

	if !errors.Is(err, domain.ErrInsufficientFunds) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcessPayment_AccountBlocked(t *testing.T) {
	acc := &domain.Account{
		ID:      "acc-1",
		Balance: 100,
		Status:  domain.AccountStatusBlocked,
	}

	svc, _ := newService(acc, 0.015)

	tx := domain.Transaction{
		AccountID: "acc-1",
		Amount:    300,
		Type:      domain.TxTypeDebit,
	}

	err := svc.ProcessPayment(context.Background(), tx)

	if !errors.Is(err, domain.ErrAccountBlocked) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcessPayment_Idempotency(t *testing.T) {
	acc := &domain.Account{
		ID:      "acc-1",
		Balance: 10000,
		Status:  domain.AccountStatusActive,
	}

	svc, txRepo := newService(acc, 0.015)

	key := "order-1"

	tx := domain.Transaction{
		AccountID:      "acc-1",
		Amount:         300,
		Type:           domain.TxTypeDebit,
		IdempotencyKey: &key,
	}

	err := svc.ProcessPayment(context.Background(), tx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	txRepo.byIdempotencyKey[key] = &domain.Transaction{ID: "tx-exists"}

	err = svc.ProcessPayment(context.Background(), tx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(txRepo.created) != 2 {
		t.Errorf("expected 2 transactions (payment + fee), got %d", len(txRepo.created))
	}
}

func TestProcessPayment_FeeCalculation(t *testing.T) {
	acc := &domain.Account{
		ID:      "acc-1",
		Balance: 10000,
		Status:  domain.AccountStatusActive,
	}

	svc, txRepo := newService(acc, 0.015)

	tx := domain.Transaction{
		AccountID: "acc-1",
		Amount:    10000,
		Type:      domain.TxTypeDebit,
	}

	err := svc.ProcessPayment(context.Background(), tx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if txRepo.created[1].Amount != 150 {
		t.Errorf("expected fee amount equal to 150, got %d", txRepo.created[1].Amount)
	}

	if txRepo.created[1].Type != domain.TxTypeFee {
		t.Errorf("expected type %q, got %q", domain.TxTypeFee, txRepo.created[1].Type)
	}
}
