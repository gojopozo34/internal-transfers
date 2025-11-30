package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"github.com/riteshkumar/internal-transfers/internal/errors"
	"github.com/riteshkumar/internal-transfers/internal/models"
	"github.com/riteshkumar/internal-transfers/internal/repository"
)

type AccountService interface {
	CreateAccount(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error)
	GetAccount(ctx context.Context, id string) (*models.Account, error)
}

type AccountServiceImpl struct {
	accountRepo repository.AccountRepository
	auditRepo   repository.AuditRepository
	logger      *slog.Logger
}

func NewAccountService(accountRepo repository.AccountRepository, auditRepo repository.AuditRepository, logger *slog.Logger) *AccountServiceImpl {
	return &AccountServiceImpl{
		accountRepo: accountRepo,
		auditRepo:   auditRepo,
		logger:      logger,
	}
}

func (s *AccountServiceImpl) CreateAccount(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error) {
	if err := s.validateCreateRequest(req); err != nil {
		s.logger.Warn("invalid create account request",
			"account_id", req.ID,
			"error", err.Error(),
		)
		return nil, err
	}

	account := &models.Account{
		ID:      req.ID,
		Balance: req.InitialBalance,
	}

	if err := s.accountRepo.CreateAccount(ctx, account); err != nil {
		if errors.IsAlreadyExists(err) {
			s.logger.Warn("account already exists",
				"account_id", req.ID,
			)
			return nil, err
		}

		s.logger.Error("failed to create account",
			"account_id", req.ID,
			"error", err.Error(),
		)
		return nil, err
	}

	// Log audit entry for account creation
	if err := s.createAccoutAuditLog(ctx, account); err != nil {
		s.logger.Error("failed to create audit log for account creation",
			"account_id", req.ID,
			"error", err.Error(),
		)
	}
	s.logger.Info("account created successfully",
		"account_id", req.ID,
	)
	return account, nil
}

func (s *AccountServiceImpl) GetAccount(ctx context.Context, id string) (*models.Account, error) {
	if id == "" {
		return nil, errors.ErrInvalidAccountID
	}

	account, err := s.accountRepo.GetAccountByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			s.logger.Warn("account not found",
				"account_id", id,
			)
			return nil, err
		}
		s.logger.Error("failed to get account",
			"account_id", id,
			"error", err.Error(),
		)
		return nil, err
	}

	return account, nil
}

func (s *AccountServiceImpl) validateCreateRequest(req *models.CreateAccountRequest) error {
	if req.ID == "" {
		return errors.ErrInvalidAccountID
	}
	if req.InitialBalance < 0 {
		return errors.ErrNegativeBalance
	}
	return nil
}

func (s *AccountServiceImpl) createAccoutAuditLog(ctx context.Context, account *models.Account) error {
	snapshot := models.AccountBalanceSnapshot{
		ID:      account.ID,
		Balance: account.Balance,
	}

	newValue, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	auditLog := &models.AuditLog{
		EntityType: models.EntityTypeAccount,
		EntityID:   account.ID,
		Action:     models.AuditActionCreate,
		NewValue:   newValue,
	}

	return s.auditRepo.CreateWithDB(ctx, auditLog)
}

// This function retrieves an account with a lock for updae within a trnasaction
// This is used internally by the transaction service to ensure consistency during transfers
func GetAccountForUpdate(ctx context.Context, tx *sql.Tx, accountRepo repository.AccountRepository, id string) (*models.Account, error) {
	return accountRepo.GetAccountByIDForUpdate(ctx, tx, id)
}
