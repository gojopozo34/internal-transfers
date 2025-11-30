package errors

import (
	"errors"
	"fmt"
)

// Domain error type for internal transfer application
var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrInsufficentBalance   = errors.New("insufficient balance")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvalidAccountID     = errors.New("invalid account ID")
	ErrSameAccount          = errors.New("source and destination accounts cannot be the same")
	ErrNegativeBalance      = errors.New("balance cannot be negative")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

type TransactionError struct {
	Operation string
	Cause     error
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf("transaction error during '%s': %v", e.Operation, e.Cause)
}

func (e *TransactionError) Unwrap() error {
	return e.Cause
}

func NewTransactionError(operation string, cause error) error {
	return &TransactionError{
		Operation: operation,
		Cause:     cause,
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrAccountNotFound)
}

func IsInsufficientBalance(err error) bool {
	return errors.Is(err, ErrInsufficentBalance)
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAccountAlreadyExists)
}
