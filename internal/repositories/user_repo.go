package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error)
	GetTenantIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetByEmailGlobal(ctx context.Context, email string) (*models.User, error)
	UpdatePassword(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string) error
	UpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, status string) error
	FindUsersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error)
	UpdateGoogleID(ctx context.Context, tenantID, userID uuid.UUID, googleID string) error
	ListByStatus(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*models.User, error)
	// IsFirstUserInTenant atomically checks if a user is the first user in a tenant using FOR UPDATE SKIP LOCKED
	IsFirstUserInTenant(ctx context.Context, tenantID uuid.UUID) (bool, error)
	// IsPlatformAdmin checks if a user is a platform admin (super admin)
	IsPlatformAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	// SetPlatformAdmin sets the platform admin flag for a user
	SetPlatformAdmin(ctx context.Context, userID uuid.UUID, isPlatformAdmin bool) error
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	query := `
	INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, status, is_platform_admin, two_factor_secret, two_factor_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, user.ID, user.TenantID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Status, user.IsPlatformAdmin, user.TwoFactorSecret, user.TwoFactorEnabled)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, tenant_id, email, google_id, first_name, last_name, status, two_factor_secret, two_factor_enabled, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND id = $2
	`
	var twoFactorSecret sql.NullString
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&user.ID, &user.TenantID, &user.Email, &user.GoogleID, &user.FirstName, &user.LastName, &user.Status, &twoFactorSecret, &user.TwoFactorEnabled, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if twoFactorSecret.Valid {
		user.TwoFactorSecret = twoFactorSecret.String
	}
	return user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, tenant_id, email, google_id, password_hash, first_name, last_name, status, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND email = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, email).Scan(&user.ID, &user.TenantID, &user.Email, &user.GoogleID, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, status = $3, two_factor_secret = $4, two_factor_enabled = $5, updated_at = NOW()
		WHERE tenant_id = $6 AND id = $7
	`
	_, err := r.db.Exec(ctx, query, user.FirstName, user.LastName, user.Status, user.TwoFactorSecret, user.TwoFactorEnabled, user.TenantID, user.ID)
	return err
}

func (r *userRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM users WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *userRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, tenant_id, email, first_name, last_name, status, created_at, updated_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.FirstName, &user.LastName, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepo) GetTenantIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT tenant_id FROM users WHERE id = $1`
	var tenantID sql.NullString
	err := r.db.QueryRow(ctx, query, userID).Scan(&tenantID)
	if err != nil {
		return uuid.Nil, err
	}
	if !tenantID.Valid {
		return uuid.Nil, fmt.Errorf("tenant_id is NULL for user %s", userID)
	}
	id, err := uuid.Parse(tenantID.String)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid tenant_id format for user %s: %w", userID, err)
	}
	return id, nil
}

// GetByEmailGlobal searches for a user by email across all tenants.
// SECURITY NOTE: This function intentionally searches across all tenants for login purposes.
// It does not filter by tenant_id because:
//  1. Users may not know their tenant ID when logging in
//  2. Subdomain-based tenant resolution happens at a different layer
//
// IMPORTANT: Callers MUST implement timing-safe responses to prevent user enumeration attacks.
// The caller should return the same response time and message whether or not a user is found.
// Additionally, rate limiting should be applied to login endpoints.
func (r *userRepo) GetByEmailGlobal(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, tenant_id, email, google_id, password_hash, first_name, last_name, status, is_platform_admin, created_at, updated_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`
	var tenantID string
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &tenantID, &user.Email, &user.GoogleID, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Status, &user.IsPlatformAdmin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	parsedTenantID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id format for email %s: %w", email, err)
	}
	user.TenantID = parsedTenantID
	return user, nil
}

func (r *userRepo) UpdatePassword(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND id = $3
	`
	_, err := r.db.Exec(ctx, query, passwordHash, tenantID, userID)
	return err
}

func (r *userRepo) UpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, status string) error {
	query := `
		UPDATE users
		SET status = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND id = $3
	`
	_, err := r.db.Exec(ctx, query, status, tenantID, userID)
	return err
}

func (r *userRepo) FindUsersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT id, tenant_id, email, first_name, last_name, status, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND status = 'active'
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.FirstName, &user.LastName, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepo) UpdateGoogleID(ctx context.Context, tenantID, userID uuid.UUID, googleID string) error {
	query := `
		UPDATE users
		SET google_id = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND id = $3
	`
	_, err := r.db.Exec(ctx, query, googleID, tenantID, userID)
	return err
}

func (r *userRepo) ListByStatus(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, tenant_id, email, first_name, last_name, status, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.FirstName, &user.LastName, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// IsFirstUserInTenant atomically checks if the current signup is the first user in a tenant.
// Uses advisory locks to prevent race conditions where multiple signups could both become admin.
// This provides an atomic check at the database level that serializes concurrent requests.
// Note: This is called AFTER the user is created, so we check for count <= 1 (only the current user exists)
func (r *userRepo) IsFirstUserInTenant(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	// Use a transaction with an advisory lock based on tenant ID
	// Advisory locks are session-level and released when the transaction ends
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback if not committed

	// Convert tenantID to a lock key (use first 8 bytes as int64)
	lockKey := int64(tenantID[0])<<56 | int64(tenantID[1])<<48 | int64(tenantID[2])<<40 |
		int64(tenantID[3])<<32 | int64(tenantID[4])<<24 | int64(tenantID[5])<<16 |
		int64(tenantID[6])<<8 | int64(tenantID[7])

	// Acquire advisory lock - this will wait if another transaction holds it
	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey)
	if err != nil {
		return false, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	// Now safely check if only one user exists in this tenant (the one just created)
	// We check count <= 1 because the current user has already been created before this check
	var count int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE tenant_id = $1", tenantID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to count users: %w", err)
	}

	// Commit to release the advisory lock
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// If there's only 1 user (the current one), this is the first user
	return count <= 1, nil
}

// IsPlatformAdmin checks if a user is a platform admin (super admin)
func (r *userRepo) IsPlatformAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	var isPlatformAdmin bool
	query := `SELECT COALESCE(is_platform_admin, false) FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&isPlatformAdmin)
	if err != nil {
		return false, err
	}
	return isPlatformAdmin, nil
}

// SetPlatformAdmin sets the platform admin flag for a user
func (r *userRepo) SetPlatformAdmin(ctx context.Context, userID uuid.UUID, isPlatformAdmin bool) error {
	query := `UPDATE users SET is_platform_admin = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, isPlatformAdmin, userID)
	return err
}
