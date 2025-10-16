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
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	query := `
	INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, status, two_factor_secret, two_factor_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, user.ID, user.TenantID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Status, user.TwoFactorSecret, user.TwoFactorEnabled)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, tenant_id, email, first_name, last_name, status, two_factor_secret, two_factor_enabled, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&user.ID, &user.TenantID, &user.Email, &user.FirstName, &user.LastName, &user.Status, &user.TwoFactorSecret, &user.TwoFactorEnabled, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, status, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND email = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, email).Scan(&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Status, &user.CreatedAt, &user.UpdatedAt)
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

func (r *userRepo) GetByEmailGlobal(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var tenantID string
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &tenantID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Status, &user.CreatedAt, &user.UpdatedAt)
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
	return users, nil
}
