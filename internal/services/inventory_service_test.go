package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProductRepoForInventory is a mock implementation of ProductRepository for inventory tests
type MockProductRepoForInventory struct {
	mock.Mock
}

func (m *MockProductRepoForInventory) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

// MockWarehouseRepoForInventory is a mock implementation of WarehouseRepository for inventory tests
type MockWarehouseRepoForInventory struct {
	mock.Mock
}

func (m *MockWarehouseRepoForInventory) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Warehouse, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Warehouse), args.Error(1)
}

// Inventory Service Tests using shared MockInventoryRepository from mocks_test.go

func TestInventoryService_CreateInventory(t *testing.T) {
	tenantID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()

	tests := []struct {
		name          string
		inventory     *models.Inventory
		mockSetup     func(*MockInventoryRepository, *MockProductRepoForInventory, *MockWarehouseRepoForInventory)
		expectedError bool
		errorContains string
	}{
		{
			name: "successful create",
			inventory: &models.Inventory{
				TenantID:    tenantID,
				ProductID:   productID,
				WarehouseID: warehouseID,
				Quantity:    100,
			},
			mockSetup: func(invRepo *MockInventoryRepository, prodRepo *MockProductRepoForInventory, whRepo *MockWarehouseRepoForInventory) {
				prodRepo.On("GetByID", mock.Anything, tenantID, productID).
					Return(&models.Product{ID: productID, TenantID: tenantID, Name: "Test Product"}, nil)
				whRepo.On("GetByID", mock.Anything, tenantID, warehouseID).
					Return(&models.Warehouse{ID: warehouseID, TenantID: tenantID, Name: "Test Warehouse"}, nil)
				invRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Inventory")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "product not found",
			inventory: &models.Inventory{
				TenantID:    tenantID,
				ProductID:   productID,
				WarehouseID: warehouseID,
				Quantity:    100,
			},
			mockSetup: func(invRepo *MockInventoryRepository, prodRepo *MockProductRepoForInventory, whRepo *MockWarehouseRepoForInventory) {
				prodRepo.On("GetByID", mock.Anything, tenantID, productID).
					Return(nil, errors.New("product not found"))
			},
			expectedError: true,
			errorContains: "product not found",
		},
		{
			name: "negative quantity rejected",
			inventory: &models.Inventory{
				TenantID:    tenantID,
				ProductID:   productID,
				WarehouseID: warehouseID,
				Quantity:    -10,
			},
			mockSetup: func(invRepo *MockInventoryRepository, prodRepo *MockProductRepoForInventory, whRepo *MockWarehouseRepoForInventory) {
				prodRepo.On("GetByID", mock.Anything, tenantID, productID).
					Return(&models.Product{ID: productID, TenantID: tenantID, Name: "Test Product"}, nil)
				whRepo.On("GetByID", mock.Anything, tenantID, warehouseID).
					Return(&models.Warehouse{ID: warehouseID, TenantID: tenantID, Name: "Test Warehouse"}, nil)
				invRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Inventory")).
					Return(errors.New("quantity cannot be negative"))
			},
			expectedError: true,
			errorContains: "negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvRepo := new(MockInventoryRepository)
			mockProdRepo := new(MockProductRepoForInventory)
			mockWhRepo := new(MockWarehouseRepoForInventory)
			tt.mockSetup(mockInvRepo, mockProdRepo, mockWhRepo)

			// Simulate the workflow: first check product, then warehouse, then create
			_, prodErr := mockProdRepo.GetByID(context.Background(), tt.inventory.TenantID, tt.inventory.ProductID)
			if prodErr != nil {
				if tt.expectedError {
					assert.Contains(t, prodErr.Error(), tt.errorContains)
				}
				mockProdRepo.AssertExpectations(t)
				return
			}

			_, whErr := mockWhRepo.GetByID(context.Background(), tt.inventory.TenantID, tt.inventory.WarehouseID)
			if whErr != nil {
				if tt.expectedError {
					assert.Contains(t, whErr.Error(), tt.errorContains)
				}
				mockWhRepo.AssertExpectations(t)
				return
			}

			err := mockInvRepo.Create(context.Background(), tt.inventory)

			if tt.expectedError {
				if err != nil {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			mockInvRepo.AssertExpectations(t)
			mockProdRepo.AssertExpectations(t)
			mockWhRepo.AssertExpectations(t)
		})
	}
}

func TestInventoryService_GetInventory(t *testing.T) {
	tenantID := uuid.New()
	inventoryID := uuid.New()

	tests := []struct {
		name          string
		mockSetup     func(*MockInventoryRepository)
		expectedError bool
	}{
		{
			name: "successful get",
			mockSetup: func(m *MockInventoryRepository) {
				m.On("GetByID", mock.Anything, tenantID, inventoryID).
					Return(&models.Inventory{
						ID:          inventoryID,
						TenantID:    tenantID,
						Quantity:    50,
						LastUpdated: time.Now(),
					}, nil)
			},
			expectedError: false,
		},
		{
			name: "inventory not found",
			mockSetup: func(m *MockInventoryRepository) {
				m.On("GetByID", mock.Anything, tenantID, inventoryID).
					Return(nil, errors.New("inventory not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			tt.mockSetup(mockRepo)

			inventory, err := mockRepo.GetByID(context.Background(), tenantID, inventoryID)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, inventory)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, inventory)
				assert.Equal(t, inventoryID, inventory.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInventoryService_UpdateInventory(t *testing.T) {
	tenantID := uuid.New()
	inventoryID := uuid.New()

	tests := []struct {
		name          string
		inventory     *models.Inventory
		mockSetup     func(*MockInventoryRepository)
		expectedError bool
	}{
		{
			name: "successful update",
			inventory: &models.Inventory{
				ID:       inventoryID,
				TenantID: tenantID,
				Quantity: 75,
			},
			mockSetup: func(m *MockInventoryRepository) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Inventory")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "update non-existent inventory",
			inventory: &models.Inventory{
				ID:       inventoryID,
				TenantID: tenantID,
				Quantity: 75,
			},
			mockSetup: func(m *MockInventoryRepository) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Inventory")).
					Return(errors.New("inventory not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			tt.mockSetup(mockRepo)

			err := mockRepo.Update(context.Background(), tt.inventory)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInventoryService_DeleteInventory(t *testing.T) {
	tenantID := uuid.New()
	inventoryID := uuid.New()

	tests := []struct {
		name          string
		mockSetup     func(*MockInventoryRepository)
		expectedError bool
	}{
		{
			name: "successful delete",
			mockSetup: func(m *MockInventoryRepository) {
				m.On("Delete", mock.Anything, tenantID, inventoryID).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "delete non-existent inventory",
			mockSetup: func(m *MockInventoryRepository) {
				m.On("Delete", mock.Anything, tenantID, inventoryID).
					Return(errors.New("inventory not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			tt.mockSetup(mockRepo)

			err := mockRepo.Delete(context.Background(), tenantID, inventoryID)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInventoryService_TransferInventory(t *testing.T) {
	tenantID := uuid.New()
	productID := uuid.New()
	fromWarehouseID := uuid.New()
	toWarehouseID := uuid.New()

	tests := []struct {
		name          string
		quantity      int
		mockSetup     func(*MockInventoryRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:     "successful transfer",
			quantity: 25,
			mockSetup: func(m *MockInventoryRepository) {
				m.On("Transfer", mock.Anything, tenantID, productID, fromWarehouseID, toWarehouseID, 25).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "insufficient stock",
			quantity: 1000,
			mockSetup: func(m *MockInventoryRepository) {
				m.On("Transfer", mock.Anything, tenantID, productID, fromWarehouseID, toWarehouseID, 1000).
					Return(errors.New("insufficient stock for transfer"))
			},
			expectedError: true,
			errorContains: "insufficient",
		},
		{
			name:     "zero quantity rejected",
			quantity: 0,
			mockSetup: func(m *MockInventoryRepository) {
				m.On("Transfer", mock.Anything, tenantID, productID, fromWarehouseID, toWarehouseID, 0).
					Return(errors.New("quantity must be positive"))
			},
			expectedError: true,
			errorContains: "positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockInventoryRepository)
			tt.mockSetup(mockRepo)

			err := mockRepo.Transfer(context.Background(), tenantID, productID, fromWarehouseID, toWarehouseID, tt.quantity)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
