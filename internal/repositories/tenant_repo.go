package repositories

import (
	"context"
	"encoding/json"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Tenant, error)
	FindSettingsByTenantID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	UpdateSettings(ctx context.Context, tenant *models.Tenant) error
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
	query := `
		SELECT id, name, subdomain, license_number, status, created_at, updated_at, sso_config
		FROM tenants
		WHERE id = $1
	`
	return scanTenant(r.db.QueryRow(ctx, query, id))
}

func (r *tenantRepo) GetBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error) {
	query := `
		SELECT id, name, subdomain, license_number, status, created_at, updated_at, sso_config
		FROM tenants
		WHERE subdomain = $1
	`
	return scanTenant(r.db.QueryRow(ctx, query, subdomain))
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
	// sso_config intentionally omitted here to keep list lightweight
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
		var license *string
		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Subdomain, &license, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt); err != nil {
			return nil, err
		}
		if license != nil {
			tenant.License = *license
		}
		tenants = append(tenants, tenant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *tenantRepo) FindSettingsByTenantID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return r.GetByID(ctx, id)
}

func (r *tenantRepo) UpdateSettings(ctx context.Context, tenant *models.Tenant) error {
	return r.Update(ctx, tenant)
}

func scanTenant(row pgx.Row) (*models.Tenant, error) {
	tenant := &models.Tenant{}
	var ssoConfig json.RawMessage
	var license *string
	if err := row.Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Subdomain,
		&license,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&ssoConfig,
	); err != nil {
		return nil, err
	}

	if license != nil {
		tenant.License = *license
	}

	if len(ssoConfig) > 0 {
		var cfg models.SSOConfig
		if err := json.Unmarshal(ssoConfig, &cfg); err == nil {
			tenant.SSOConfig = &cfg
		}
	}

	return tenant, nil
}
