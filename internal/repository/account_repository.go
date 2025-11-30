package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/riteshkumar/internal-transfers/internal/errors"
	"github.com/riteshkumar/internal-transfers/internal/models"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, account *models.Account) error
	GetAccountByID(ctx context.Context, id string) (*models.Account, error)
	GetAccountByIDForUpdate(ctx context.Context, tx *sql.Tx, id string) (*models.Account, error)
	UpdateAccountBalance(ctx context.Context, tx *sql.Tx, id string, newBalance float64) error
	AccountExists(ctx context.Context, id string) (bool, error)
}

type PostgresAccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) CreateAccount(ctx context.Context, account *models.Account) error {
	query := `INSERT INTO accounts (id, balance, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, account.ID, account.Balance).
		Scan(&account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.ErrAccountAlreadyExists
		}
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

func (r *PostgresAccountRepository) GetAccountByID(ctx context.Context, id string) (*models.Account, error) {
	query := `SELECT id, balance, created_at, updated_at FROM accounts WHERE id = $1`

	account := &models.Account{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&account.ID, &account.Balance, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}
	return account, nil
}

func (r *PostgresAccountRepository) GetAccountByIDForUpdate(ctx context.Context, tx *sql.Tx, id string) (*models.Account, error) {
	query := `SELECT id, balance, created_at, updated_at FROM accounts WHERE id = $1 FOR UPDATE`

	account := &models.Account{}
	err := tx.QueryRowContext(ctx, query, id).
		Scan(&account.ID, &account.Balance, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account by ID for update: %w", err)
	}

	return account, nil
}

func (r *PostgresAccountRepository) UpdateAccountBalance(ctx context.Context, tx *sql.Tx, id string, newBalance float64) error {
	query := `UPDATE accounts SET balance = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := tx.ExecContext(ctx, query, newBalance, id)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after updating account balance: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrAccountNotFound
	}

	return nil
}

func (r *PostgresAccountRepository) AccountExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if account exists: %w", err)
	}

	return exists, nil
}
