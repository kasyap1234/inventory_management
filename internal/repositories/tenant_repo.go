package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Tenant, error)
}

type tenantRepo struct {
	db *pgxpool.Pool
}

func NewTenantRepo(db *pgxpool.Pool) TenantRepository {
	return &tenantRepo{db: db}
}

func (r *tenantRepo) Create(ctx context.Context, tenant *models.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, subdomain, license_number, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, tenant.ID, tenant.Name, tenant.Subdomain, tenant.License, tenant.Status)
	return err
}

func (r *tenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	tenant := &models.Tenant{}
	query := `
		SELECT id, name, subdomain, license_number, status, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(&tenant.ID, &tenant.Name, &tenant.Subdomain, &tenant.License, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func (r *tenantRepo) GetBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error) {
	tenant := &models.Tenant{}
	query := `
		SELECT id, name, subdomain, license_number, status, created_at, updated_at
		FROM tenants
		WHERE subdomain = $1
	`
	err := r.db.QueryRow(ctx, query, subdomain).Scan(&tenant.ID, &tenant.Name, &tenant.Subdomain, &tenant.License, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
	return tenant, err
}

func (r *tenantRepo) Update(ctx context.Context, tenant *models.Tenant) error {
	query := `
		UPDATE tenants
		SET name = $1, subdomain = $2, license_number = $3, status = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, tenant.Name, tenant.Subdomain, tenant.License, tenant.Status, tenant.ID)
	return err
}

func (r *tenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *tenantRepo) List(ctx context.Context, limit, offset int) ([]*models.Tenant, error) {
	query := `
		SELECT id, name, subdomain, license_number, status, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*models.Tenant
	for rows.Next() {
		tenant := &models.Tenant{}
		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Subdomain, &tenant.License, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}