package services

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockInventoryRepository is a shared mock implementation of InventoryRepository
type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) GetByWarehouseAndProduct(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, tenantID, warehouseID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) Update(ctx context.Context, inventory *models.Inventory) error {
	args := m.Called(ctx, inventory)
	return args.Error(0)
}

func (m *MockInventoryRepository) Create(ctx context.Context, inventory *models.Inventory) error {
	args := m.Called(ctx, inventory)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockInventoryRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error) {
	args := m.Called(ctx, tenantID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) GetByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.Inventory, error) {
	args := m.Called(ctx, tenantID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Inventory), args.Error(1)
}

// GetByProductForUpdate retrieves inventory with SELECT FOR UPDATE lock
func (m *MockInventoryRepository) GetByProductForUpdate(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.Inventory, error) {
	args := m.Called(ctx, tenantID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) Transfer(ctx context.Context, tenantID, productID, fromWarehouseID, toWarehouseID uuid.UUID, quantity int) error {
	args := m.Called(ctx, tenantID, productID, fromWarehouseID, toWarehouseID, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepository) AdjustStock(ctx context.Context, tenantID, productID uuid.UUID, change int) error {
	args := m.Called(ctx, tenantID, productID, change)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetByProductID(ctx context.Context, tenantID, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, tenantID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inventory), args.Error(1)
}
