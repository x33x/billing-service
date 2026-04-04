package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/x33x/billing-service/internal/db"
	"github.com/x33x/billing-service/internal/domain"
)

type LedgerRepository struct {
	db *db.DB
}

func NewLedgerRepository(db *db.DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

// insert records in the same tran DB as insert transaction and update balance
func (r *LedgerRepository) createBatch(ctx context.Context, dbTx pgx.Tx, entries []domain.LedgerEntry) error {
	query := "insert into ledger_entries (id, transaction_id, account_id, amount, direction) values ($1, $2, $3, $4, $5)"

	for _, entry := range entries {
		_, err := dbTx.Exec(ctx, query, entry.ID, entry.TransactionID, entry.AccountID, entry.Amount, entry.Direction)

		if err != nil {
			return fmt.Errorf("createBatch: %w", err)
		}
	}

	return nil
}
