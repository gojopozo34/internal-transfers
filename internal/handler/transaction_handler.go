package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/riteshkumar/internal-transfers/internal/errors"
	"github.com/riteshkumar/internal-transfers/internal/models"
	"github.com/riteshkumar/internal-transfers/internal/service"
	u "github.com/riteshkumar/internal-transfers/internal/utils"
)

type TransactionHandler struct {
	transactionService service.TransactionService
	logger             *slog.Logger
}

func NewTransactionHandler(transactionService service.TransactionService, logger *slog.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		logger:             logger,
	}
}

func (h *TransactionHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/transactions", h.CreateTransaction).Methods(http.MethodPost)
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid create transaction request", "error", err.Error())
		u.WriteError(w, http.StatusBadRequest, "invalid request payload", err.Error())
		return
	}

	transaction, err := h.transactionService.Transfer(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err, "create transaction")
		return
	}

	u.WriteJSON(w, http.StatusCreated, models.TransactionResponse{
		ID:                   transaction.ID,
		SourceAccountID:      transaction.SourceAccountID,
		DestinationAccountID: transaction.DestinationAccountID,
		Amount:               transaction.Amount,
		CreatedAt:            transaction.CreatedAt,
	})
}

func (h *TransactionHandler) handleServiceError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.IsNotFound(err):
		u.WriteError(w, http.StatusNotFound, "acount not found", err.Error())
	case errors.IsInsufficientBalance(err):
		u.WriteError(w, http.StatusBadRequest, "insufficient balance", "source account does not have enough funds for txn")
	case errors.IsValidationError(err):
		u.WriteError(w, http.StatusBadRequest, "validation error", err.Error())
	case err == errors.ErrSameAccount:
		u.WriteError(w, http.StatusBadRequest, "same source and destination account", err.Error())
	case err == errors.ErrInvalidAmount:
		u.WriteError(w, http.StatusBadRequest, "invalid amount", err.Error())
	default:
		h.logger.Error("internal server error during "+action, "error", err.Error())
		u.WriteError(w, http.StatusInternalServerError, "internal server error", "")
	}
}
