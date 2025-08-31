package repositories

import (
	"context"
	"agromart2/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *models.Warehouse) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Warehouse, error)
	Update(ctx context.Context, warehouse *models.Warehouse) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Warehouse, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Warehouse, error)
}

type warehouseRepo struct {
	db *pgxpool.Pool
}

func NewWarehouseRepository(db *pgxpool.Pool) WarehouseRepository {
	return &warehouseRepo{db: db}
}

func (r *warehouseRepo) Create(ctx context.Context, warehouse *models.Warehouse) error {
	query := `
		INSERT INTO warehouses (id, tenant_id, name, address, capacity, license_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, warehouse.ID, warehouse.TenantID, warehouse.Name, warehouse.Address, warehouse.Capacity, warehouse.LicenseNumber)
	return err
}

func (r *warehouseRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Warehouse, error) {
	warehouse := &models.Warehouse{}
	query := `
		SELECT id, tenant_id, name, address, capacity, license_number, created_at, updated_at
		FROM warehouses
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&warehouse.ID, &warehouse.TenantID, &warehouse.Name, &warehouse.Address, &warehouse.Capacity, &warehouse.LicenseNumber, &warehouse.CreatedAt, &warehouse.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return warehouse, nil
}

func (r *warehouseRepo) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Warehouse, error) {
	warehouse := &models.Warehouse{}
	query := `
		SELECT id, tenant_id, name, address, capacity, license_number, created_at, updated_at
		FROM warehouses
		WHERE tenant_id = $1 AND name = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, name).Scan(&warehouse.ID, &warehouse.TenantID, &warehouse.Name, &warehouse.Address, &warehouse.Capacity, &warehouse.LicenseNumber, &warehouse.CreatedAt, &warehouse.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return warehouse, nil
}

func (r *warehouseRepo) Update(ctx context.Context, warehouse *models.Warehouse) error {
	query := `
		UPDATE warehouses
		SET name = $1, address = $2, capacity = $3, license_number = $4, updated_at = NOW()
		WHERE tenant_id = $5 AND id = $6
	`
	_, err := r.db.Exec(ctx, query, warehouse.Name, warehouse.Address, warehouse.Capacity, warehouse.LicenseNumber, warehouse.TenantID, warehouse.ID)
	return err
}

func (r *warehouseRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM warehouses WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *warehouseRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Warehouse, error) {
	query := `
		SELECT id, tenant_id, name, address, capacity, license_number, created_at, updated_at
		FROM warehouses
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warehouses []*models.Warehouse
	for rows.Next() {
		warehouse := &models.Warehouse{}
		if err := rows.Scan(&warehouse.ID, &warehouse.TenantID, &warehouse.Name, &warehouse.Address, &warehouse.Capacity, &warehouse.LicenseNumber, &warehouse.CreatedAt, &warehouse.UpdatedAt); err != nil {
			return nil, err
		}
		warehouses = append(warehouses, warehouse)
	}
	return warehouses, nil
}