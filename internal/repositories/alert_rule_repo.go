package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AlertRuleRepository interface for alert rule operations
type AlertRuleRepository interface {
	Create(ctx context.Context, rule *models.AlertRule) error
	Update(ctx context.Context, rule *models.AlertRule) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.AlertRule, error)
	List(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.AlertRule, error)
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	GetActiveRulesForEvent(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.AlertRule, error)
	UpdateTriggerStatus(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
}

// alertRuleRepo implements AlertRuleRepository
type alertRuleRepo struct {
	db *pgxpool.Pool
}

// NewAlertRuleRepo creates a new alert rule repository
func NewAlertRuleRepo(db *pgxpool.Pool) AlertRuleRepository {
	return &alertRuleRepo{db: db}
}

// Create creates a new alert rule
func (r *alertRuleRepo) Create(ctx context.Context, rule *models.AlertRule) error {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	query := `
		INSERT INTO alert_rules (
			id, tenant_id, name, event_type, conditions, actions,
			is_active, trigger_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
		)`

	_, err = r.db.Exec(ctx, query,
		rule.ID, rule.TenantID, rule.Name, rule.EventType,
		conditionsJSON, actionsJSON, rule.IsActive, rule.TriggerCount,
	)

	if err != nil {
		// Check for unique constraint violation (duplicate name)
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505") {
			return fmt.Errorf("alert rule with name '%s' already exists", rule.Name)
		}
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	return nil
}

// Update updates an existing alert rule
func (r *alertRuleRepo) Update(ctx context.Context, rule *models.AlertRule) error {
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	query := `
		UPDATE alert_rules SET
			name = $3, event_type = $4, conditions = $5, actions = $6,
			is_active = $7, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query,
		rule.TenantID, rule.ID, rule.Name, rule.EventType,
		conditionsJSON, actionsJSON, rule.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update alert rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert rule not found")
	}

	return nil
}

// GetByID retrieves an alert rule by ID
func (r *alertRuleRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.AlertRule, error) {
	query := `
		SELECT id, tenant_id, name, event_type, conditions, actions,
			   is_active, last_triggered_at, trigger_count, created_at, updated_at
		FROM alert_rules
		WHERE tenant_id = $1 AND id = $2`

	var rule models.AlertRule
	var conditionsJSON, actionsJSON []byte

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&rule.ID, &rule.TenantID, &rule.Name, &rule.EventType,
		&conditionsJSON, &actionsJSON, &rule.IsActive,
		&rule.LastTriggeredAt, &rule.TriggerCount,
		&rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("alert rule not found")
		}
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}

	if err := json.Unmarshal(actionsJSON, &rule.Actions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
	}

	return &rule, nil
}

// List retrieves alert rules, optionally filtered by event type
func (r *alertRuleRepo) List(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.AlertRule, error) {
	query := `
		SELECT id, tenant_id, name, event_type, conditions, actions,
			   is_active, last_triggered_at, trigger_count, created_at, updated_at
		FROM alert_rules
		WHERE tenant_id = $1`

	args := []interface{}{tenantID}

	if eventType != "" {
		query += " AND event_type = $2"
		args = append(args, eventType)
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*models.AlertRule

	for rows.Next() {
		var rule models.AlertRule
		var conditionsJSON, actionsJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Name, &rule.EventType,
			&conditionsJSON, &actionsJSON, &rule.IsActive,
			&rule.LastTriggeredAt, &rule.TriggerCount,
			&rule.CreatedAt, &rule.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}

		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		if err := json.Unmarshal(actionsJSON, &rule.Actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

// Delete deletes an alert rule
func (r *alertRuleRepo) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM alert_rules WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert rule not found")
	}

	return nil
}

// GetActiveRulesForEvent retrieves active alert rules for a specific event type
func (r *alertRuleRepo) GetActiveRulesForEvent(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.AlertRule, error) {
	query := `
		SELECT id, tenant_id, name, event_type, conditions, actions,
			   is_active, last_triggered_at, trigger_count, created_at, updated_at
		FROM alert_rules
		WHERE tenant_id = $1 AND event_type = $2 AND is_active = true
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, tenantID, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to get active alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*models.AlertRule

	for rows.Next() {
		var rule models.AlertRule
		var conditionsJSON, actionsJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Name, &rule.EventType,
			&conditionsJSON, &actionsJSON, &rule.IsActive,
			&rule.LastTriggeredAt, &rule.TriggerCount,
			&rule.CreatedAt, &rule.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}

		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		if err := json.Unmarshal(actionsJSON, &rule.Actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

// UpdateTriggerStatus updates the trigger status when an alert is triggered
func (r *alertRuleRepo) UpdateTriggerStatus(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `
		UPDATE alert_rules SET
			last_triggered_at = NOW(),
			trigger_count = trigger_count + 1,
			updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to update alert rule trigger status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert rule not found")
	}

	return nil
}