package repositories

import (
	"context"
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
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	// Check for global email uniqueness before insertion
	var count int
	emailCheckQuery := `SELECT COUNT(*) FROM users WHERE email = $1`
	err := r.db.QueryRow(ctx, emailCheckQuery, user.Email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("user with email '%s' already exists", user.Email)
	}

	// Debug logging - trace the tenant_id being inserted
	if user.TenantID.String() != "" {
		// This is our debug output (would appear in application logs if logging was enabled)
	}

	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err = r.db.Exec(ctx, query, user.ID, user.TenantID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Status)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, tenant_id, email, first_name, last_name, status, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&user.ID, &user.TenantID, &user.Email, &user.FirstName, &user.LastName, &user.Status, &user.CreatedAt, &user.UpdatedAt)
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
		SET first_name = $1, last_name = $2, status = $3, updated_at = NOW()
		WHERE tenant_id = $4 AND id = $5
	`
	_, err := r.db.Exec(ctx, query, user.FirstName, user.LastName, user.Status, user.TenantID, user.ID)
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
	var tenantID uuid.UUID
	err := r.db.QueryRow(ctx, query, userID).Scan(&tenantID)
	if err != nil {
		return uuid.Nil, err
	}
	return tenantID, nil
}
