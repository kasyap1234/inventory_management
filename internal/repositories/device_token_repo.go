package repositories

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeviceTokenRepository interface for device token operations
type DeviceTokenRepository interface {
	RegisterToken(ctx context.Context, token *models.DeviceToken) error
	GetTokensByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.DeviceToken, error)
	GetActiveTokensByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.DeviceToken, error)
	DeactivateToken(ctx context.Context, tenantID uuid.UUID, deviceToken string) error
	UpdateLastUsed(ctx context.Context, tenantID uuid.UUID, deviceToken string) error
	GetByToken(ctx context.Context, tenantID uuid.UUID, deviceToken string) (*models.DeviceToken, error)
	DeleteToken(ctx context.Context, tenantID uuid.UUID, deviceToken string) error
}

// deviceTokenRepository implements DeviceTokenRepository
type deviceTokenRepository struct {
	db *pgxpool.Pool
}

// NewDeviceTokenRepository creates a new device token repository
func NewDeviceTokenRepository(db *pgxpool.Pool) DeviceTokenRepository {
	return &deviceTokenRepository{db: db}
}

// RegisterToken registers or updates a device token
func (r *deviceTokenRepository) RegisterToken(ctx context.Context, token *models.DeviceToken) error {
	query := `
		INSERT INTO device_tokens (
			id, tenant_id, user_id, device_token, device_type, 
			device_name, app_version, is_active, last_used_at, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tenant_id, user_id, device_token) 
		DO UPDATE SET 
			device_name = EXCLUDED.device_name,
			app_version = EXCLUDED.app_version,
			is_active = true,
			last_used_at = EXCLUDED.last_used_at,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}
	if token.UpdatedAt.IsZero() {
		token.UpdatedAt = time.Now()
	}
	if token.LastUsedAt.IsZero() {
		token.LastUsedAt = time.Now()
	}

	err := r.db.QueryRow(ctx, query,
		token.ID,
		token.TenantID,
		token.UserID,
		token.DeviceToken,
		token.DeviceType,
		token.DeviceName,
		token.AppVersion,
		token.IsActive,
		token.LastUsedAt,
		token.CreatedAt,
		token.UpdatedAt,
	).Scan(&token.ID, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to register device token: %w", err)
	}

	return nil
}

// GetTokensByUser retrieves all device tokens for a user
func (r *deviceTokenRepository) GetTokensByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.DeviceToken, error) {
	query := `
		SELECT id, tenant_id, user_id, device_token, device_type, 
		       device_name, app_version, is_active, last_used_at, 
		       created_at, updated_at, deleted_at
		FROM device_tokens
		WHERE tenant_id = $1 AND user_id = $2 AND deleted_at IS NULL
		ORDER BY last_used_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*models.DeviceToken
	for rows.Next() {
		token := &models.DeviceToken{}
		err := rows.Scan(
			&token.ID,
			&token.TenantID,
			&token.UserID,
			&token.DeviceToken,
			&token.DeviceType,
			&token.DeviceName,
			&token.AppVersion,
			&token.IsActive,
			&token.LastUsedAt,
			&token.CreatedAt,
			&token.UpdatedAt,
			&token.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// GetActiveTokensByUser retrieves only active device tokens for a user
func (r *deviceTokenRepository) GetActiveTokensByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.DeviceToken, error) {
	query := `
		SELECT id, tenant_id, user_id, device_token, device_type, 
		       device_name, app_version, is_active, last_used_at, 
		       created_at, updated_at, deleted_at
		FROM device_tokens
		WHERE tenant_id = $1 AND user_id = $2 
		  AND is_active = true AND deleted_at IS NULL
		ORDER BY last_used_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active device tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*models.DeviceToken
	for rows.Next() {
		token := &models.DeviceToken{}
		err := rows.Scan(
			&token.ID,
			&token.TenantID,
			&token.UserID,
			&token.DeviceToken,
			&token.DeviceType,
			&token.DeviceName,
			&token.AppVersion,
			&token.IsActive,
			&token.LastUsedAt,
			&token.CreatedAt,
			&token.UpdatedAt,
			&token.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// DeactivateToken marks a device token as inactive
func (r *deviceTokenRepository) DeactivateToken(ctx context.Context, tenantID uuid.UUID, deviceToken string) error {
	query := `
		UPDATE device_tokens 
		SET is_active = false, updated_at = NOW()
		WHERE tenant_id = $1 AND device_token = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, tenantID, deviceToken)
	if err != nil {
		return fmt.Errorf("failed to deactivate device token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("device token not found")
	}

	return nil
}

// UpdateLastUsed updates the last used timestamp for a device token
func (r *deviceTokenRepository) UpdateLastUsed(ctx context.Context, tenantID uuid.UUID, deviceToken string) error {
	query := `
		UPDATE device_tokens 
		SET last_used_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND device_token = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, tenantID, deviceToken)
	if err != nil {
		return fmt.Errorf("failed to update last used timestamp: %w", err)
	}

	return nil
}

// GetByToken retrieves a device token by its token string
func (r *deviceTokenRepository) GetByToken(ctx context.Context, tenantID uuid.UUID, deviceToken string) (*models.DeviceToken, error) {
	query := `
		SELECT id, tenant_id, user_id, device_token, device_type, 
		       device_name, app_version, is_active, last_used_at, 
		       created_at, updated_at, deleted_at
		FROM device_tokens
		WHERE tenant_id = $1 AND device_token = $2 AND deleted_at IS NULL
	`

	token := &models.DeviceToken{}
	err := r.db.QueryRow(ctx, query, tenantID, deviceToken).Scan(
		&token.ID,
		&token.TenantID,
		&token.UserID,
		&token.DeviceToken,
		&token.DeviceType,
		&token.DeviceName,
		&token.AppVersion,
		&token.IsActive,
		&token.LastUsedAt,
		&token.CreatedAt,
		&token.UpdatedAt,
		&token.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get device token: %w", err)
	}

	return token, nil
}

// DeleteToken soft deletes a device token
func (r *deviceTokenRepository) DeleteToken(ctx context.Context, tenantID uuid.UUID, deviceToken string) error {
	query := `
		UPDATE device_tokens 
		SET deleted_at = NOW(), is_active = false, updated_at = NOW()
		WHERE tenant_id = $1 AND device_token = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, tenantID, deviceToken)
	if err != nil {
		return fmt.Errorf("failed to delete device token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("device token not found")
	}

	return nil
}
