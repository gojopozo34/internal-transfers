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

type AccountHandler struct {
	accountService service.AccountService
	logger         *slog.Logger
}

func NewAccountHandler(accountService service.AccountService, logger *slog.Logger) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		logger:         logger,
	}
}

func (h *AccountHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/accounts", h.CreateAccount).Methods(http.MethodPost)
	router.HandleFunc("/accounts/{id}", h.GetAccount).Methods(http.MethodGet)
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid create account request", "error", err.Error())
		u.WriteError(w, http.StatusBadRequest, "invalid request payload", err.Error())
		return
	}

	account, err := h.accountService.CreateAccount(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err, "create account")
		return
	}

	u.WriteJSON(w, http.StatusCreated, models.AccountResponse{
		ID:      account.ID,
		Balance: account.Balance,
	})
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["id"]

	if accountID == "" {
		u.WriteError(w, http.StatusBadRequest, "id is required", "")
		return
	}

	account, err := h.accountService.GetAccount(r.Context(), accountID)
	if err != nil {
		h.handleServiceError(w, err, "get account")
		return
	}

	u.WriteJSON(w, http.StatusOK, models.AccountResponse{
		ID:      account.ID,
		Balance: account.Balance,
	})
}

func (h *AccountHandler) handleServiceError(w http.ResponseWriter, err error, operation string) {
	switch {
	case errors.IsNotFound(err):
		u.WriteError(w, http.StatusNotFound, "account not found", "")
	case errors.IsAlreadyExists(err):
		u.WriteError(w, http.StatusConflict, "account already exists", "")
	case errors.IsValidationError(err):
		u.WriteError(w, http.StatusBadRequest, "validation error", err.Error())
	case err == errors.ErrInvalidAccountID:
		u.WriteError(w, http.StatusBadRequest, "invalid account ID", "")
	case err == errors.ErrNegativeBalance:
		u.WriteError(w, http.StatusBadRequest, "negative balance not allowed", "")
	default:
		h.logger.Error("internal server error during "+operation, "error", err.Error())
		u.WriteError(w, http.StatusInternalServerError, "internal server error", "")
	}
}
