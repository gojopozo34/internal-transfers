package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/riteshkumar/internal-transfers/internal/models"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error
	GetByID(ctx context.Context, id string) (*models.Transaction, error)
	GetByAccountID(ctx context.Context, accountID string) ([]*models.Transaction, error)
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error {
	// Generate UUID if not set
	if transaction.ID == "" {
		transaction.ID = uuid.New().String()
	}

	query := `INSERT INTO transactions (id, source_account_id, destination_account_id, amount)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	err := tx.QueryRowContext(ctx, query,
		transaction.ID,
		transaction.SourceAccountID,
		transaction.DestinationAccountID,
		transaction.Amount,
	).Scan(&transaction.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

func (r *PostgresTransactionRepository) GetByID(ctx context.Context, id string) (*models.Transaction, error) {
	query := `SELECT id, source_account_id, destination_account_id, amount, created_at
		FROM transactions WHERE id = $1`

	transaction := &models.Transaction{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&transaction.ID, &transaction.SourceAccountID, &transaction.DestinationAccountID, &transaction.Amount, &transaction.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}
	return transaction, nil
}

func (r *PostgresTransactionRepository) GetByAccountID(ctx context.Context, accountID string) ([]*models.Transaction, error) {
	query := `SELECT id, source_account_id, destination_account_id, amount, created_at
		FROM transactions
		WHERE source_account_id = $1 OR destination_account_id = $1
		ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by account ID: %w", err)
	}
	defer rows.Close()
	var transactions []*models.Transaction
	for rows.Next() {
		transaction := &models.Transaction{}
		err := rows.Scan(&transaction.ID, &transaction.SourceAccountID, &transaction.DestinationAccountID, &transaction.Amount, &transaction.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over transactions: %w", err)
	}
	return transactions, nil
}
