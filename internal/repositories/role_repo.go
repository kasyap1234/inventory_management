package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	Exec(ctx context.Context, sql string, args ...interface{}) error
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Role, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error)
}

type roleRepo struct {
	db *pgxpool.Pool
}

func NewRoleRepo(db *pgxpool.Pool) RoleRepository {
	return &roleRepo{db: db}
}

func (r *roleRepo) Create(ctx context.Context, role *models.Role) error {
	query := `
		INSERT INTO roles (id, tenant_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (tenant_id, name) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, role.ID, role.TenantID, role.Name, role.Description)
	return err
}

func (r *roleRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Role, error) {
	role := &models.Role{}
	query := `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *roleRepo) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	role := &models.Role{}
	query := `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND name = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, name).Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *roleRepo) Update(ctx context.Context, role *models.Role) error {
	query := `
		UPDATE roles
		SET name = $1, description = $2, updated_at = NOW()
		WHERE tenant_id = $3 AND id = $4
	`
	_, err := r.db.Exec(ctx, query, role.Name, role.Description, role.TenantID, role.ID)
	return err
}

func (r *roleRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *roleRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}