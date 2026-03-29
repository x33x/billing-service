package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/x33x/billing-service/internal/db"
	"github.com/x33x/billing-service/internal/domain"
)

type AccountRepository struct {
	db *db.DB
}

func NewAccountRepository(db *db.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	acc := &domain.Account{}

	query := "select id, currency, balance, status from accounts where id = $1"
	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&acc.ID,
		&acc.Currency,
		&acc.Balance,
		&acc.Status,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}

		return nil, fmt.Errorf("GetByID: %w", err)
	}

	return acc, nil
}

func (r *AccountRepository) Create(ctx context.Context, acc domain.Account) error {
	query := "insert into accounts (id, currency, balance, status) values ($1, $2, $3, $4)"
	_, err := r.db.Pool().Exec(ctx, query, acc.ID, acc.Currency, acc.Balance, acc.Status)

	if err != nil {
		return fmt.Errorf("Create: %w", err)
	}

	return nil
}
