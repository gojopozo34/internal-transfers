package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/riteshkumar/internal-transfers/internal/models"
)

type AuditRepository interface {
	Create(ctx context.Context, tx *sql.Tx, log *models.AuditLog) error
	CreateWithDB(ctx context.Context, log *models.AuditLog) error
	GetByEntityID(ctx context.Context, entityType, entityID string) ([]*models.AuditLog, error)
}

type PostgresAuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *PostgresAuditRepository {
	return &PostgresAuditRepository{db: db}
}

// Create inserts a new audit log entry within a db transaction.
func (r *PostgresAuditRepository) Create(ctx context.Context, tx *sql.Tx, log *models.AuditLog) error {
	query := `INSERT INTO audit_logs (entity_type, entity_id, action, old_value, new_value, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		RETURNING id, created_at`

	var oldValue interface{}
	if log.OldValue != nil {
		oldValue = log.OldValue
	}
	err := tx.QueryRowContext(ctx, query,
		log.EntityType,
		log.EntityID,
		log.Action,
		oldValue,
		log.NewValue,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// CreateWithDB inserts a new audit log entry using the db connection directly
// Used for operations that don't require a transaction (e.g., logging account creation)
func (r *PostgresAuditRepository) CreateWithDB(ctx context.Context, log *models.AuditLog) error {
	query := `INSERT INTO audit_logs (entity_type, entity_id, action, old_value, new_value, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		RETURNING id, created_at`

	var oldValue interface{}
	if log.OldValue != nil {
		oldValue = log.OldValue
	}

	err := r.db.QueryRowContext(ctx, query,
		log.EntityType,
		log.EntityID,
		log.Action,
		oldValue,
		log.NewValue,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByEntityID retrieves audit logs for a specific entity type and ID.
func (r *PostgresAuditRepository) GetByEntityID(ctx context.Context, entityType, entityID string) ([]*models.AuditLog, error) {
	query := `SELECT id, entity_type, entity_id, action, old_value, new_value, created_at
		FROM audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by entity ID: %w", err)
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		log := &models.AuditLog{}
		var oldValue, newValue []byte

		err := rows.Scan(
			&log.ID, &log.EntityType, &log.EntityID, &log.Action, &oldValue, &newValue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if oldValue != nil {
			log.OldValue = json.RawMessage(oldValue)
		}
		log.NewValue = json.RawMessage(newValue)

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over audit logs: %w", err)
	}
	return logs, nil
}
