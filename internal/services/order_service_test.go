package services

import (
	"context"
	"testing"
	"time"

	"agromart2/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository is a mock implementation of OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, tenantID, orderID uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, tenantID, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, tenantID, orderID uuid.UUID) error {
	args := m.Called(ctx, tenantID, orderID)
	return args.Error(0)
}

func (m *MockOrderRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.OrderSearchFilter) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByTenantAndDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByDistributor(ctx context.Context, tenantID, distributorID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, distributorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByStatus(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, supplierID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByTypeAndStatus(ctx context.Context, tenantID uuid.UUID, orderType, status string, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, orderType, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

// MockInventoryRepository is now defined in mocks_test.go

// Test ValidateStatusTransition
func TestValidateStatusTransition(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockInventoryRepo := new(MockInventoryRepository)
	
	service := &orderService{
		orderRepo:     mockOrderRepo,
		inventoryRepo: mockInventoryRepo,
	}

	tests := []struct {
		name          string
		currentStatus string
		newStatus     string
		expectError   bool
	}{
		{"Valid: pending to approved", "pending", "approved", false},
		{"Valid: pending to cancelled", "pending", "cancelled", false},
		{"Valid: approved to processing", "approved", "processing", false},
		{"Valid: processing to shipped", "processing", "shipped", false},
		{"Valid: shipped to delivered", "shipped", "delivered", false},
		{"Invalid: pending to shipped", "pending", "shipped", true},
		{"Invalid: delivered to processing", "delivered", "processing", true},
		{"Invalid: cancelled to approved", "cancelled", "approved", true},
		{"Same status", "pending", "pending", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateStatusTransition(tt.currentStatus, tt.newStatus)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test CreateOrder with inventory validation
func TestCreateOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockInventoryRepo := new(MockInventoryRepository)
	
	service := &orderService{
		orderRepo:     mockOrderRepo,
		inventoryRepo: mockInventoryRepo,
	}

	ctx := context.Background()
	tenantID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()

	t.Run("Success: Create sales order with sufficient inventory", func(t *testing.T) {
		order := &models.Order{
			TenantID:    tenantID,
			OrderType:   "sales",
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    10,
			UnitPrice:   100.0,
		}

		inventory := &models.Inventory{
			ID:          uuid.New(),
			TenantID:    tenantID,
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    50,
		}

		mockInventoryRepo.On("GetByWarehouseAndProduct", ctx, tenantID, warehouseID, productID).Return(inventory, nil)
		mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*models.Order")).Return(nil)

		err := service.CreateOrder(ctx, tenantID, order)
		assert.NoError(t, err)
		assert.Equal(t, "pending", order.Status)
		assert.NotEqual(t, uuid.Nil, order.ID)

		mockInventoryRepo.AssertExpectations(t)
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("Failure: Insufficient inventory for sales order", func(t *testing.T) {
		order := &models.Order{
			TenantID:    tenantID,
			OrderType:   "sales",
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    100,
			UnitPrice:   100.0,
		}

		inventory := &models.Inventory{
			ID:          uuid.New(),
			TenantID:    tenantID,
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    50,
		}

		mockInventoryRepo.On("GetByWarehouseAndProduct", ctx, tenantID, warehouseID, productID).Return(inventory, nil)

		err := service.CreateOrder(ctx, tenantID, order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient inventory")

		mockInventoryRepo.AssertExpectations(t)
	})
}

// Test ProcessOrder with inventory reservation
func TestProcessOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockInventoryRepo := new(MockInventoryRepository)
	
	service := &orderService{
		orderRepo:     mockOrderRepo,
		inventoryRepo: mockInventoryRepo,
	}

	ctx := context.Background()
	tenantID := uuid.New()
	orderID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()

	t.Run("Success: Process order and reserve inventory", func(t *testing.T) {
		order := &models.Order{
			ID:          orderID,
			TenantID:    tenantID,
			OrderType:   "sales",
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    10,
			UnitPrice:   100.0,
			Status:      "approved",
		}

		inventory := &models.Inventory{
			ID:          uuid.New(),
			TenantID:    tenantID,
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    50,
		}

		mockOrderRepo.On("GetByID", ctx, tenantID, orderID).Return(order, nil)
		mockInventoryRepo.On("GetByWarehouseAndProduct", ctx, tenantID, warehouseID, productID).Return(inventory, nil)
		mockInventoryRepo.On("Update", ctx, mock.AnythingOfType("*models.Inventory")).Return(nil)
		mockOrderRepo.On("Update", ctx, mock.AnythingOfType("*models.Order")).Return(nil)

		err := service.ProcessOrder(ctx, tenantID, orderID)
		assert.NoError(t, err)

		mockOrderRepo.AssertExpectations(t)
		mockInventoryRepo.AssertExpectations(t)
	})

	t.Run("Failure: Invalid status transition", func(t *testing.T) {
		order := &models.Order{
			ID:          orderID,
			TenantID:    tenantID,
			OrderType:   "sales",
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    10,
			UnitPrice:   100.0,
			Status:      "delivered", // Cannot process a delivered order
		}

		mockOrderRepo.On("GetByID", ctx, tenantID, orderID).Return(order, nil)

		err := service.ProcessOrder(ctx, tenantID, orderID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")

		mockOrderRepo.AssertExpectations(t)
	})
}

// Test CancelOrder with inventory restoration
func TestCancelOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockInventoryRepo := new(MockInventoryRepository)
	
	service := &orderService{
		orderRepo:     mockOrderRepo,
		inventoryRepo: mockInventoryRepo,
	}

	ctx := context.Background()
	tenantID := uuid.New()
	orderID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()

	t.Run("Success: Cancel processing order and restore inventory", func(t *testing.T) {
		order := &models.Order{
			ID:          orderID,
			TenantID:    tenantID,
			OrderType:   "sales",
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    10,
			UnitPrice:   100.0,
			Status:      "processing",
		}

		inventory := &models.Inventory{
			ID:          uuid.New(),
			TenantID:    tenantID,
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    40, // 50 - 10 (reserved)
		}

		mockOrderRepo.On("GetByID", ctx, tenantID, orderID).Return(order, nil)
		mockInventoryRepo.On("GetByWarehouseAndProduct", ctx, tenantID, warehouseID, productID).Return(inventory, nil)
		mockInventoryRepo.On("Update", ctx, mock.AnythingOfType("*models.Inventory")).Return(nil)
		mockOrderRepo.On("Update", ctx, mock.AnythingOfType("*models.Order")).Return(nil)

		err := service.CancelOrder(ctx, tenantID, orderID)
		assert.NoError(t, err)

		mockOrderRepo.AssertExpectations(t)
		mockInventoryRepo.AssertExpectations(t)
	})

	t.Run("Failure: Cannot cancel delivered order", func(t *testing.T) {
		order := &models.Order{
			ID:          orderID,
			TenantID:    tenantID,
			OrderType:   "sales",
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    10,
			UnitPrice:   100.0,
			Status:      "delivered",
		}

		mockOrderRepo.On("GetByID", ctx, tenantID, orderID).Return(order, nil)

		err := service.CancelOrder(ctx, tenantID, orderID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")

		mockOrderRepo.AssertExpectations(t)
	})
}

// Test order lifecycle
func TestOrderLifecycle(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockInventoryRepo := new(MockInventoryRepository)
	
	service := &orderService{
		orderRepo:     mockOrderRepo,
		inventoryRepo: mockInventoryRepo,
	}

	ctx := context.Background()
	tenantID := uuid.New()
	orderID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()

	// Create order (pending)
	order := &models.Order{
		ID:          orderID,
		TenantID:    tenantID,
		OrderType:   "sales",
		ProductID:   productID,
		WarehouseID: warehouseID,
		Quantity:    10,
		UnitPrice:   100.0,
		Status:      "pending",
	}

	inventory := &models.Inventory{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ProductID:   productID,
		WarehouseID: warehouseID,
		Quantity:    50,
	}

	// Test: pending -> approved
	t.Run("Approve order", func(t *testing.T) {
		mockOrderRepo.On("GetByID", ctx, tenantID, orderID).Return(order, nil).Once()
		mockOrderRepo.On("Update", ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()

		err := service.ApproveOrder(ctx, tenantID, orderID)
		assert.NoError(t, err)
	})

	// Update order status for next test
	order.Status = "approved"

	// Test: approved -> processing
	t.Run("Process order", func(t *testing.T) {
		mockOrderRepo.On("GetByID", ctx, tenantID, orderID).Return(order, nil).Once()
		mockInventoryRepo.On("GetByWarehouseAndProduct", ctx, tenantID, warehouseID, productID).Return(inventory, nil).Once()
		mockInventoryRepo.On("Update", ctx, mock.AnythingOfType("*models.Inventory")).Return(nil).Once()
		mockOrderRepo.On("Update", ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()

		err := service.ProcessOrder(ctx, tenantID, orderID)
		assert.NoError(t, err)
	})

	// Update order status for next test
	order.Status = "processing"

	// Test: processing -> shipped
	t.Run("Ship order", func(t *testing.T) {
		mockOrderRepo.On("GetByID", ctx, tenantID, orderID).Return(order, nil).Once()
		mockOrderRepo.On("Update", ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()

		expectedDelivery := time.Now().Add(48 * time.Hour)
		err := service.ShipOrder(ctx, tenantID, orderID, &expectedDelivery)
		assert.NoError(t, err)
	})

	// Update order status for next test
	order.Status = "shipped"

	// Test: shipped -> delivered
	t.Run("Deliver order", func(t *testing.T) {
		mockOrderRepo.On("GetByID", ctx, tenantID, orderID).Return(order, nil).Once()
		mockOrderRepo.On("Update", ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()

		err := service.DeliverOrder(ctx, tenantID, orderID)
		assert.NoError(t, err)
	})

	mockOrderRepo.AssertExpectations(t)
	mockInventoryRepo.AssertExpectations(t)
}
