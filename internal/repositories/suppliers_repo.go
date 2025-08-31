package repositories

import (
	"context"
	"agromart2/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *models.Supplier) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Supplier, error)
	Update(ctx context.Context, supplier *models.Supplier) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Supplier, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Supplier, error)
}

type supplierRepo struct {
	db *pgxpool.Pool
}

func NewSupplierRepository(db *pgxpool.Pool) SupplierRepository {
	return &supplierRepo{db: db}
}

func (r *supplierRepo) Create(ctx context.Context, supplier *models.Supplier) error {
	query := `
		INSERT INTO suppliers (id, tenant_id, name, contact_email, contact_phone, address, license_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, supplier.ID, supplier.TenantID, supplier.Name, supplier.ContactEmail, supplier.ContactPhone, supplier.Address, supplier.LicenseNumber)
	return err
}

func (r *supplierRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Supplier, error) {
	supplier := &models.Supplier{}
	query := `
		SELECT id, tenant_id, name, contact_email, contact_phone, address, license_number, created_at, updated_at
		FROM suppliers
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&supplier.ID, &supplier.TenantID, &supplier.Name, &supplier.ContactEmail, &supplier.ContactPhone, &supplier.Address, &supplier.LicenseNumber, &supplier.CreatedAt, &supplier.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return supplier, nil
}

func (r *supplierRepo) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Supplier, error) {
	supplier := &models.Supplier{}
	query := `
		SELECT id, tenant_id, name, contact_email, contact_phone, address, license_number, created_at, updated_at
		FROM suppliers
		WHERE tenant_id = $1 AND name = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, name).Scan(&supplier.ID, &supplier.TenantID, &supplier.Name, &supplier.ContactEmail, &supplier.ContactPhone, &supplier.Address, &supplier.LicenseNumber, &supplier.CreatedAt, &supplier.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return supplier, nil
}

func (r *supplierRepo) Update(ctx context.Context, supplier *models.Supplier) error {
	query := `
		UPDATE suppliers
		SET name = $1, contact_email = $2, contact_phone = $3, address = $4, license_number = $5, updated_at = NOW()
		WHERE tenant_id = $6 AND id = $7
	`
	_, err := r.db.Exec(ctx, query, supplier.Name, supplier.ContactEmail, supplier.ContactPhone, supplier.Address, supplier.LicenseNumber, supplier.TenantID, supplier.ID)
	return err
}

func (r *supplierRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM suppliers WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *supplierRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Supplier, error) {
	query := `
		SELECT id, tenant_id, name, contact_email, contact_phone, address, license_number, created_at, updated_at
		FROM suppliers
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []*models.Supplier
	for rows.Next() {
		supplier := &models.Supplier{}
		if err := rows.Scan(&supplier.ID, &supplier.TenantID, &supplier.Name, &supplier.ContactEmail, &supplier.ContactPhone, &supplier.Address, &supplier.LicenseNumber, &supplier.CreatedAt, &supplier.UpdatedAt); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, supplier)
	}
	return suppliers, nil
}