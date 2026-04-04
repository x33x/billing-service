package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/x33x/billing-service/internal/db"
	"github.com/x33x/billing-service/internal/domain"
)

type TransactionRepository struct {
	db     *db.DB
	ledger *LedgerRepository
}

func NewTransactionRepository(db *db.DB, ledger *LedgerRepository) *TransactionRepository {
	return &TransactionRepository{
		db:     db,
		ledger: ledger,
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tx domain.Transaction) error {
	// begin tran
	dbTx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	// rollback if something wrong
	defer dbTx.Rollback(ctx)

	// save transaction in table
	query := "insert into transactions (id, account_id, amount, type, status, idempotency_key) values ($1, $2, $3, $4, $5, $6)"
	_, err = dbTx.Exec(ctx, query, tx.ID, tx.AccountID, tx.Amount, tx.Type, tx.Status, tx.IdempotencyKey)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // error unique violation
			return fmt.Errorf("Create: %w", domain.ErrDuplicateTransaction)
		}

		return fmt.Errorf("insert: %w", err)
	}

	// update balance in table
	switch tx.Type {
	case domain.TxTypeDebit:
		query = "update accounts set balance = balance - $1 where id = $2"
	case domain.TxTypeCredit:
		query = "update accounts set balance = balance + $1 where id = $2"
	}
	_, err = dbTx.Exec(ctx, query, tx.Amount, tx.AccountID)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	// insert ledger entries
	now := time.Now().UnixNano()
	entries := []domain.LedgerEntry{
		{
			ID:            fmt.Sprintf("ledger_%d_d", now),
			TransactionID: tx.ID,
			AccountID:     tx.AccountID,
			Amount:        tx.Amount,
			Direction:     domain.LedgerDirectionDebit,
		},
		{
			ID:            fmt.Sprintf("ledger_%d_c", now),
			TransactionID: tx.ID,
			AccountID:     tx.AccountID,
			Amount:        tx.Amount,
			Direction:     domain.LedgerDirectionCredit,
		},
	}

	if err := r.ledger.createBatch(ctx, dbTx, entries); err != nil {
		return fmt.Errorf("Create: ledger: %w", err)
	}

	// commit
	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func (r *TransactionRepository) GetByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	query := "select idempotency_key, id, account_id, amount, type, status, created_at from transactions where account_id = $1 order by created_at desc"
	rows, err := r.db.Pool().Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("GetByAccountID: %w", err)
	}
	defer rows.Close()

	txs := make([]domain.Transaction, 0) // init empty slice, not nil
	for rows.Next() {
		tx := domain.Transaction{}
		if err := rows.Scan(
			&tx.IdempotencyKey,
			&tx.ID,
			&tx.AccountID,
			&tx.Amount,
			&tx.Type,
			&tx.Status,
			&tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		txs = append(txs, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetByAccountID: %w", err)
	}

	return txs, nil
}

func (r *TransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	tx := &domain.Transaction{}

	query := "select id, account_id, amount, type, status, created_at, idempotency_key from transactions where idempotency_key = $1"
	err := r.db.Pool().QueryRow(ctx, query, key).Scan(
		&tx.ID,
		&tx.AccountID,
		&tx.Amount,
		&tx.Type,
		&tx.Status,
		&tx.CreatedAt,
		&tx.IdempotencyKey,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("GetByIdempotencyKey: %w", err)
	}

	return tx, nil
}
