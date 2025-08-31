package repositories

import (
	"context"
	"agromart2/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DistributorRepository interface {
	Create(ctx context.Context, distributor *models.Distributor) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Distributor, error)
	Update(ctx context.Context, distributor *models.Distributor) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Distributor, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Distributor, error)
}

type distributorRepo struct {
	db *pgxpool.Pool
}

func NewDistributorRepository(db *pgxpool.Pool) DistributorRepository {
	return &distributorRepo{db: db}
}

func (r *distributorRepo) Create(ctx context.Context, distributor *models.Distributor) error {
	query := `
		INSERT INTO distributors (id, tenant_id, name, contact_email, contact_phone, address, license_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, distributor.ID, distributor.TenantID, distributor.Name, distributor.ContactEmail, distributor.ContactPhone, distributor.Address, distributor.LicenseNumber)
	return err
}

func (r *distributorRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Distributor, error) {
	distributor := &models.Distributor{}
	query := `
		SELECT id, tenant_id, name, contact_email, contact_phone, address, license_number, created_at, updated_at
		FROM distributors
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&distributor.ID, &distributor.TenantID, &distributor.Name, &distributor.ContactEmail, &distributor.ContactPhone, &distributor.Address, &distributor.LicenseNumber, &distributor.CreatedAt, &distributor.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return distributor, nil
}

func (r *distributorRepo) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Distributor, error) {
	distributor := &models.Distributor{}
	query := `
		SELECT id, tenant_id, name, contact_email, contact_phone, address, license_number, created_at, updated_at
		FROM distributors
		WHERE tenant_id = $1 AND name = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, name).Scan(&distributor.ID, &distributor.TenantID, &distributor.Name, &distributor.ContactEmail, &distributor.ContactPhone, &distributor.Address, &distributor.LicenseNumber, &distributor.CreatedAt, &distributor.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return distributor, nil
}

func (r *distributorRepo) Update(ctx context.Context, distributor *models.Distributor) error {
	query := `
		UPDATE distributors
		SET name = $1, contact_email = $2, contact_phone = $3, address = $4, license_number = $5, updated_at = NOW()
		WHERE tenant_id = $6 AND id = $7
	`
	_, err := r.db.Exec(ctx, query, distributor.Name, distributor.ContactEmail, distributor.ContactPhone, distributor.Address, distributor.LicenseNumber, distributor.TenantID, distributor.ID)
	return err
}

func (r *distributorRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM distributors WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *distributorRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Distributor, error) {
	query := `
		SELECT id, tenant_id, name, contact_email, contact_phone, address, license_number, created_at, updated_at
		FROM distributors
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var distributors []*models.Distributor
	for rows.Next() {
		distributor := &models.Distributor{}
		if err := rows.Scan(&distributor.ID, &distributor.TenantID, &distributor.Name, &distributor.ContactEmail, &distributor.ContactPhone, &distributor.Address, &distributor.LicenseNumber, &distributor.CreatedAt, &distributor.UpdatedAt); err != nil {
			return nil, err
		}
		distributors = append(distributors, distributor)
	}
	return distributors, nil
}