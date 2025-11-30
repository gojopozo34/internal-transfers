package models

import (
	"encoding/json"
	"time"
)

type Account struct {
	ID        string    `json:"id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID                   string    `json:"id"`
	SourceAccountID      string    `json:"source_account_id"`
	DestinationAccountID string    `json:"destination_account_id"`
	Amount               float64   `json:"amount"`
	CreatedAt            time.Time `json:"created_at"`
}

type AuditLog struct {
	ID         string          `json:"id"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Action     string          `json:"action"`
	OldValue   json.RawMessage `json:"old_value"`
	NewValue   json.RawMessage `json:"new_value"`
	CreatedAt  time.Time       `json:"created_at"`
}

const (
	AuditActionCreate   = "CREATE"
	AuditActionUpdate   = "UPDATE"
	AuditActionTransfer = "TRANSFER"
)

const (
	EntityTypeAccount     = "ACCOUNT"
	EntityTypeTransaction = "TRANSACTION"
)

type CreateAccountRequest struct {
	ID             string  `json:"id"`
	InitialBalance float64 `json:"initial_balance"`
}

type AccountResponse struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
}

type CreateTransactionRequest struct {
	SourceAccountID      string  `json:"source_account_id"`
	DestinationAccountID string  `json:"destination_account_id"`
	Amount               float64 `json:"amount"`
}

type TransactionResponse struct {
	ID                   string    `json:"id"`
	SourceAccountID      string    `json:"source_account_id"`
	DestinationAccountID string    `json:"destination_account_id"`
	Amount               float64   `json:"amount"`
	CreatedAt            time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type AccountBalanceSnapshot struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
}
