package services

import (
	"context"
	"errors"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type WarehouseService interface {
	Create(ctx context.Context, tenantID uuid.UUID, warehouse *models.Warehouse) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Warehouse, error)
	Update(ctx context.Context, tenantID uuid.UUID, warehouse *models.Warehouse) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Warehouse, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Warehouse, error)
}

type warehouseService struct {
	warehouseRepo repositories.WarehouseRepository
}

func NewWarehouseService(warehouseRepo repositories.WarehouseRepository) WarehouseService {
	return &warehouseService{
		warehouseRepo: warehouseRepo,
	}
}

func (s *warehouseService) Create(ctx context.Context, tenantID uuid.UUID, warehouse *models.Warehouse) error {
	if warehouse.Name == "" {
		return errors.New("warehouse name is required")
	}

	if warehouse.Capacity == nil || *warehouse.Capacity <= 0 {
		return errors.New("warehouse capacity must be greater than 0")
	}

	// Check for duplicate name
	existing, err := s.warehouseRepo.GetByName(ctx, tenantID, warehouse.Name)
	if err == nil && existing != nil {
		return errors.New("warehouse with this name already exists")
	}

	warehouse.TenantID = tenantID
	warehouse.ID = uuid.New()

	return s.warehouseRepo.Create(ctx, warehouse)
}

func (s *warehouseService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Warehouse, error) {
	return s.warehouseRepo.GetByID(ctx, tenantID, id)
}

func (s *warehouseService) Update(ctx context.Context, tenantID uuid.UUID, warehouse *models.Warehouse) error {
	if warehouse.Name == "" {
		return errors.New("warehouse name is required")
	}

	if warehouse.Capacity != nil && *warehouse.Capacity <= 0 {
		return errors.New("warehouse capacity must be greater than 0")
	}

	warehouse.TenantID = tenantID
	return s.warehouseRepo.Update(ctx, warehouse)
}

func (s *warehouseService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.warehouseRepo.Delete(ctx, tenantID, id)
}

func (s *warehouseService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Warehouse, error) {
	return s.warehouseRepo.List(ctx, tenantID, limit, offset)
}

func (s *warehouseService) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Warehouse, error) {
	return s.warehouseRepo.GetByName(ctx, tenantID, name)
}