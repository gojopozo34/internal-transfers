package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/riteshkumar/internal-transfers/internal/errors"
	"github.com/riteshkumar/internal-transfers/internal/models"
	"github.com/riteshkumar/internal-transfers/internal/repository"
)

type TransactionService interface {
	Transfer(ctx context.Context, req *models.CreateTransactionRequest) (*models.Transaction, error)
}

type TransactionServiceImpl struct {
	db              *sql.DB
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
	auditRepo       repository.AuditRepository
	logger          *slog.Logger
}

func NewTransactionService(db *sql.DB, accountRepo repository.AccountRepository, transactionRepo repository.TransactionRepository, auditRepo repository.AuditRepository, logger *slog.Logger) *TransactionServiceImpl {
	return &TransactionServiceImpl{
		db:              db,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		auditRepo:       auditRepo,
		logger:          logger,
	}
}

// Transfer performs a money transfer b/w 2 accounts
// Uses db txns with row level locking to ensure consistency
func (s *TransactionServiceImpl) Transfer(ctx context.Context, req *models.CreateTransactionRequest) (*models.Transaction, error) {
	if err := s.validateTransferRequest(ctx, req); err != nil {
		s.logger.Warn("invalid transfer request",
			"source_account_id", req.SourceAccountID,
			"destination_account_id", req.DestinationAccountID,
			"amount", req.Amount,
			"error", err.Error(),
		)
		return nil, err
	}

	// Begin txn with SERIALIZABLE isolation level for strict consistency
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		s.logger.Error("failed to begin transaction",
			"error", err.Error(),
		)
		return nil, errors.NewTransactionError("begin", err)
	}

	// Ensure rollback on error
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// Lock and get source account
	sourceAccount, err := s.accountRepo.GetAccountByIDForUpdate(ctx, tx, req.SourceAccountID)
	if err != nil {
		if errors.IsNotFound(err) {
			s.logger.Error("source account not found",
				"source_account_id", req.SourceAccountID,
			)
			return nil, fmt.Errorf("source account: %w", err)
		}
		s.logger.Error("failed to get source account",
			"source_account_id", req.SourceAccountID,
			"error", err.Error(),
		)
		return nil, errors.NewTransactionError("get source account", err)
	}

	// Lock and get destination account
	destinationAccount, err := s.accountRepo.GetAccountByIDForUpdate(ctx, tx, req.DestinationAccountID)
	if err != nil {
		if errors.IsNotFound(err) {
			s.logger.Error("destination account not found",
				"destination_account_id", req.DestinationAccountID,
			)
			return nil, fmt.Errorf("destination account: %w", err)
		}
		s.logger.Error("failed to get destination account",
			"destination_account_id", req.DestinationAccountID,
			"error", err.Error(),
		)
		return nil, errors.NewTransactionError("get destination account", err)
	}

	// Check for sufficient balance
	if sourceAccount.Balance < req.Amount {
		s.logger.Warn("insufficient balance in source account",
			"source_account_id", req.SourceAccountID,
			"available_balance", sourceAccount.Balance,
			"requested_amount", req.Amount,
		)
		return nil, errors.ErrInsufficentBalance
	}

	// store old balance for audit
	oldSourceBalance := sourceAccount.Balance
	oldDestinationBalance := destinationAccount.Balance

	// calculate new balances
	newSourceBalance := sourceAccount.Balance - req.Amount
	newDestinationBalance := destinationAccount.Balance + req.Amount

	// Update source account balance
	if err := s.accountRepo.UpdateAccountBalance(ctx, tx, req.SourceAccountID, newSourceBalance); err != nil {
		s.logger.Error("failed to update source account balance",
			"source_account_id", req.SourceAccountID,
			"error", err.Error(),
		)
		return nil, errors.NewTransactionError("update source account balance", err)
	}

	// Update destination account balance
	if err := s.accountRepo.UpdateAccountBalance(ctx, tx, req.DestinationAccountID, newDestinationBalance); err != nil {
		s.logger.Error("failed to update destination account balance",
			"destination_account_id", req.DestinationAccountID,
			"error", err.Error(),
		)
		return nil, errors.NewTransactionError("update destination account balance", err)
	}

	// Create transaction record
	transaction := &models.Transaction{
		SourceAccountID:      req.SourceAccountID,
		DestinationAccountID: req.DestinationAccountID,
		Amount:               req.Amount,
	}

	if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
		s.logger.Error("failed to create transaction record",
			"source_account_id", req.SourceAccountID,
			"destination_account_id", req.DestinationAccountID,
			"amount", req.Amount,
			"error", err.Error(),
		)
		return nil, errors.NewTransactionError("create transaction record", err)
	}

	// Create audit logs for both accounts
	if err := s.createTransferAuditLog(ctx, tx, transaction, oldSourceBalance, newSourceBalance, oldDestinationBalance, newDestinationBalance); err != nil {
		s.logger.Error("failed to create audit logs for transfer",
			"transaction_id", transaction.ID,
			"error", err.Error(),
		)
		// continue with the tx even if audit loggin fails
	}

	// Commit txn
	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit transaction",
			"transaction_id", transaction.ID,
			"error", err.Error(),
		)
		return nil, errors.NewTransactionError("commit", err)
	}

	// Nullify tx to avoid rollback in defer
	tx = nil

	return transaction, nil
}

func (s *TransactionServiceImpl) validateTransferRequest(ctx context.Context, req *models.CreateTransactionRequest) error {
	if req.SourceAccountID == "" {
		return errors.NewValidationError("source_account_id", "must be non-empty")
	}
	if req.DestinationAccountID == "" {
		return errors.NewValidationError("destination_account_id", "must be non-empty")
	}
	if req.SourceAccountID == req.DestinationAccountID {
		return errors.ErrSameAccount
	}
	if req.Amount <= 0 {
		return errors.ErrInvalidAmount
	}
	return nil
}

func (s *TransactionServiceImpl) createTransferAuditLog(ctx context.Context, tx *sql.Tx, transaction *models.Transaction, oldSourceBalance, newSourceBalance, oldDestinationBalance, newDestinationBalance float64) error {
	sourceOldSnapshot := models.AccountBalanceSnapshot{
		ID:      transaction.SourceAccountID,
		Balance: oldSourceBalance,
	}

	sourceNewSnapshot := models.AccountBalanceSnapshot{
		ID:      transaction.SourceAccountID,
		Balance: newSourceBalance,
	}

	sourceOldValue, _ := json.Marshal(sourceOldSnapshot)
	sourceNewValue, _ := json.Marshal(sourceNewSnapshot)

	sourceAuditLog := &models.AuditLog{
		EntityType: "account",
		EntityID:   transaction.SourceAccountID,
		Action:     "debit",
		OldValue:   sourceOldValue,
		NewValue:   sourceNewValue,
	}

	if err := s.auditRepo.Create(ctx, tx, sourceAuditLog); err != nil {
		return fmt.Errorf("failed to create source account audit log: %w", err)
	}

	destinationOldSnapshot := models.AccountBalanceSnapshot{
		ID:      transaction.DestinationAccountID,
		Balance: oldDestinationBalance,
	}

	destinationNewSnapshot := models.AccountBalanceSnapshot{
		ID:      transaction.DestinationAccountID,
		Balance: newDestinationBalance,
	}

	destinationOldValue, _ := json.Marshal(destinationOldSnapshot)
	destinationNewValue, _ := json.Marshal(destinationNewSnapshot)

	destinationAuditLog := &models.AuditLog{
		EntityType: "account",
		EntityID:   transaction.DestinationAccountID,
		Action:     "credit",
		OldValue:   destinationOldValue,
		NewValue:   destinationNewValue,
	}

	if err := s.auditRepo.Create(ctx, tx, destinationAuditLog); err != nil {
		return fmt.Errorf("failed to create destination account audit log: %w", err)
	}

	// audit log for the tx itself
	txSnapshot := struct {
		ID                   string  `json:"id"`
		SourceAccountID      string  `json:"source_account_id"`
		DestinationAccountID string  `json:"destination_account_id"`
		Amount               float64 `json:"amount"`
	}{
		ID:                   transaction.ID,
		SourceAccountID:      transaction.SourceAccountID,
		DestinationAccountID: transaction.DestinationAccountID,
		Amount:               transaction.Amount,
	}

	txValue, _ := json.Marshal(txSnapshot)

	txAuditLog := &models.AuditLog{
		EntityType: "transaction",
		EntityID:   transaction.ID,
		Action:     "transfer",
		NewValue:   txValue,
	}

	if err := s.auditRepo.Create(ctx, tx, txAuditLog); err != nil {
		return fmt.Errorf("failed to create transaction audit log: %w", err)
	}

	return nil
}
