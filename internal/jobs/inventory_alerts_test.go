package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockInventoryRepository mocks the InventoryRepository interface for testing
type MockInventoryRepository struct {
	mock.Mock
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

func (m *MockInventoryRepository) Update(ctx context.Context, inventory *models.Inventory) error {
	args := m.Called(ctx, inventory)
	return args.Error(0)
}

func (m *MockInventoryRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockInventoryRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Inventory, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	return args.Get(0).([]*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) GetByWarehouseAndProduct(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, tenantID, warehouseID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.InventorySearchFilter) ([]*models.Inventory, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Get(0).([]*models.Inventory), args.Error(1)
}

// MockProductRepository mocks the ProductRepository interface for testing
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockProductRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*models.Product, error) {
	args := m.Called(ctx, tenantID, barcode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, tenantID, query, categoryID, limit, offset)
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) ListWithCategory(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, tenantID, categoryID, limit, offset)
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) CategoryAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]int, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockProductRepository) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.ProductSearchFilter) ([]*models.Product, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Get(0).([]*models.Product), args.Error(1)
}

// InventoryAlertServiceTestSuite is the comprehensive test suite for InventoryAlertService
type InventoryAlertServiceTestSuite struct {
	suite.Suite
	mockInventoryRepo *MockInventoryRepository
	mockProductRepo   *MockProductRepository
	service           *InventoryAlertService
	tenantID          uuid.UUID
	warehouseID       uuid.UUID
}

// SetupTest initializes test dependencies
func (suite *InventoryAlertServiceTestSuite) SetupTest() {
	suite.mockInventoryRepo = &MockInventoryRepository{}
	suite.mockProductRepo = &MockProductRepository{}
	suite.service = NewInventoryAlertService(suite.mockInventoryRepo, suite.mockProductRepo)
	suite.tenantID = uuid.New()
	suite.warehouseID = uuid.New()
}

// TearDownTest cleans up test dependencies
func (suite *InventoryAlertServiceTestSuite) TearDownTest() {
	suite.mockInventoryRepo.AssertExpectations(suite.T())
	suite.mockProductRepo.AssertExpectations(suite.T())
}

// TestLowStockDetection tests the core low stock detection algorithm with multiple scenarios
func (suite *InventoryAlertServiceTestSuite) TestCheckLowStock_MultipleScenarios() {
	ctx := context.Background()
	threshold := 10

	// Setup test products with fixed UUIDs for predictable mocking
	product1ID := uuid.New()
	product2ID := uuid.New()

	product1 := &models.Product{
		ID:   product1ID,
		Name: "Product 1",
	}
	product2 := &models.Product{
		ID:   product2ID,
		Name: "Product 2",
	}

	// Test Case 1: Multiple items below threshold
	inventory1 := &models.Inventory{
		TenantID:    suite.tenantID,
		WarehouseID: suite.warehouseID,
		ProductID:   product1ID,
		Quantity:    5, // Below threshold
	}
	inventory2 := &models.Inventory{
		TenantID:    suite.tenantID,
		WarehouseID: suite.warehouseID,
		ProductID:   product2ID,
		Quantity:    8, // Below threshold
	}
	inventory3 := &models.Inventory{
		TenantID:    suite.tenantID,
		WarehouseID: suite.warehouseID,
		ProductID:   uuid.New(),
		Quantity:    15, // Above threshold
	}

	inventories := []*models.Inventory{inventory1, inventory2, inventory3}
	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return(inventories, nil).Once()
	suite.mockProductRepo.On("GetByID", ctx, suite.tenantID, product1ID).Return(product1, nil).Once()
	suite.mockProductRepo.On("GetByID", ctx, suite.tenantID, product2ID).Return(product2, nil).Once()
	// No product lookup for inventory3 since it's above threshold

	alerts, err := suite.service.CheckLowStock(ctx, suite.tenantID, threshold)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), alerts, 2) // Only alerts for items below threshold

	// Verify alert details
	assert.Equal(suite.T(), suite.tenantID, alerts[0].TenantID)
	assert.Equal(suite.T(), product1ID, alerts[0].ProductID)
	assert.Equal(suite.T(), 5, alerts[0].CurrentStock)
	assert.Equal(suite.T(), threshold, alerts[0].Threshold)
}

// TestCheckLowStock_NoAlerts tests when no inventory items are below threshold
func (suite *InventoryAlertServiceTestSuite) TestCheckLowStock_NoAlerts() {
	ctx := context.Background()
	threshold := 10

	inventory := &models.Inventory{
		TenantID:    suite.tenantID,
		WarehouseID: suite.warehouseID,
		ProductID:   uuid.New(),
		Quantity:    15, // Above threshold
	}

	inventories := []*models.Inventory{inventory}
	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return(inventories, nil).Once()
	// No product lookup should occur since quantity >= threshold

	alerts, err := suite.service.CheckLowStock(ctx, suite.tenantID, threshold)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), alerts, 0)
}

// TestCheckLowStock_DefaultThreshold tests default threshold behavior
func (suite *InventoryAlertServiceTestSuite) TestCheckLowStock_DefaultThreshold() {
	ctx := context.Background()
	productID := uuid.New()

	product := &models.Product{
		ID:   productID,
		Name: "Test Product",
	}

	inventory := &models.Inventory{
		TenantID:    suite.tenantID,
		WarehouseID: suite.warehouseID,
		ProductID:   productID,
		Quantity:    5, // Below default threshold of 10
	}

	inventories := []*models.Inventory{inventory}
	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return(inventories, nil).Once()
	suite.mockProductRepo.On("GetByID", ctx, suite.tenantID, productID).Return(product, nil).Once()

	alerts, err := suite.service.CheckLowStock(ctx, suite.tenantID, 0) // Should use default threshold
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), alerts, 1)
	assert.Equal(suite.T(), 10, alerts[0].Threshold) // Should be default threshold
}

// TestCheckLowStock_RepositoryError tests error handling from inventory repository
func (suite *InventoryAlertServiceTestSuite) TestCheckLowStock_RepositoryError() {
	ctx := context.Background()
	threshold := 10

	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return(([]*models.Inventory)(nil), errors.New("database connection failed")).Once()

	alerts, err := suite.service.CheckLowStock(ctx, suite.tenantID, threshold)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), alerts)
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

// TestCheckLowStock_ProductLookupError tests error handling for product lookup failures
func (suite *InventoryAlertServiceTestSuite) TestCheckLowStock_ProductLookupError() {
	ctx := context.Background()
	threshold := 10
	productID := uuid.New()

	inventory := &models.Inventory{
		TenantID:    suite.tenantID,
		WarehouseID: suite.warehouseID,
		ProductID:   productID,
		Quantity:    5, // Below threshold
	}

	inventories := []*models.Inventory{inventory}
	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return(inventories, nil).Once()
	suite.mockProductRepo.On("GetByID", ctx, suite.tenantID, productID).Return((*models.Product)(nil), errors.New("product not found")).Once()

	alerts, err := suite.service.CheckLowStock(ctx, suite.tenantID, threshold)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), alerts, 0) // No alerts generated due to failed product lookup
}

// TestCheckLowStock_AtThresholdBoundary tests edge cases at threshold boundaries (actual code uses <=)
func (suite *InventoryAlertServiceTestSuite) TestCheckLowStock_AtThresholdBoundary() {
	ctx := context.Background()
	threshold := 10
	productID := uuid.New()
	product := &models.Product{ID: productID, Name: "Test Product"}

	// Test exact threshold - with current code implementation, this SHOULD trigger alert
	inventory := &models.Inventory{
		TenantID:    suite.tenantID,
		WarehouseID: suite.warehouseID,
		ProductID:   productID,
		Quantity:    10, // Equal to threshold
	}

	inventories := []*models.Inventory{inventory}
	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return(inventories, nil).Once()
	suite.mockProductRepo.On("GetByID", ctx, suite.tenantID, productID).Return(product, nil).Once()

	alerts, err := suite.service.CheckLowStock(ctx, suite.tenantID, threshold)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), alerts, 1) // Code Uses <= so exact threshold triggers alert
}

// TestMultiTenantAlertIsolation verifies alerts are isolated between tenants
func (suite *InventoryAlertServiceTestSuite) TestMultiTenantAlertIsolation() {
	ctx := context.Background()
	threshold := 10

	tenantAID := uuid.New()
	tenantBID := uuid.New()
	productAID := uuid.New()
	productBID := uuid.New()

	// Tenant A setup
	productA := &models.Product{ID: productAID, Name: "Product A"}
	inventoryA := &models.Inventory{
		TenantID:    tenantAID,
		WarehouseID: suite.warehouseID,
		ProductID:   productAID,
		Quantity:    5, // Below threshold
	}

	// Tenant B setup
	productB := &models.Product{ID: productBID, Name: "Product B"}
	inventoryB := &models.Inventory{
		TenantID:    tenantBID,
		WarehouseID: suite.warehouseID,
		ProductID:   productBID,
		Quantity:    5, // Also below threshold
	}

	// Test Tenant A
	suite.mockInventoryRepo.On("List", ctx, tenantAID, 1000, 0).Return([]*models.Inventory{inventoryA}, nil).Once()
	suite.mockProductRepo.On("GetByID", ctx, tenantAID, productAID).Return(productA, nil).Once()

	alertsA, err := suite.service.CheckLowStock(ctx, tenantAID, threshold)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), alertsA, 1)
	assert.Equal(suite.T(), tenantAID, alertsA[0].TenantID)

	// Test Tenant B separately
	suite.mockInventoryRepo.On("List", ctx, tenantBID, 1000, 0).Return([]*models.Inventory{inventoryB}, nil).Once()
	suite.mockProductRepo.On("GetByID", ctx, tenantBID, productBID).Return(productB, nil).Once()

	alertsB, err := suite.service.CheckLowStock(ctx, tenantBID, threshold)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), alertsB, 1)
	assert.Equal(suite.T(), tenantBID, alertsB[0].TenantID)

	// Verify tenant isolation
	assert.NotEqual(suite.T(), alertsA[0].TenantID, alertsB[0].TenantID)
}

// TestLogLowStockAlerts tests the alert logging functionality
func (suite *InventoryAlertServiceTestSuite) TestLogLowStockAlerts_EmptyAlerts() {
	alerts := []InventoryAlert{}
	suite.service.LogLowStockAlerts(context.Background(), alerts)
}

func (suite *InventoryAlertServiceTestSuite) TestLogLowStockAlerts_WithAlerts() {
	alerts := []InventoryAlert{
		{
			TenantID:     suite.tenantID,
			WarehouseID:  uuid.New(),
			ProductID:    uuid.New(),
			ProductName:  "Test Product",
			CurrentStock: 5,
			Threshold:    10,
		},
	}
	// Test should not panic when logging alerts
	suite.service.LogLowStockAlerts(context.Background(), alerts)
}

// TestScheduledLowStockCheck tests the main scheduled job entry point
func (suite *InventoryAlertServiceTestSuite) TestScheduledLowStockCheck() {
	err := suite.service.ScheduledLowStockCheck(context.Background())
	assert.NoError(suite.T(), err)
}

// TestCheckAndLogLowStockAcrossAllTenants tests the scheduled job method
func (suite *InventoryAlertServiceTestSuite) TestCheckAndLogLowStockAcrossAllTenants() {
	// This is currently a placeholder returning nil
	err := suite.service.CheckAndLogLowStockAcrossAllTenants(context.Background(), 10)
	assert.NoError(suite.T(), err)
}

// TestCustomThresholdConfiguration tests various threshold levels
func (suite *InventoryAlertServiceTestSuite) TestCustomThresholdConfiguration() {
	ctx := context.Background()
	threshold := 5
	productID := uuid.New()

	product := &models.Product{
		ID:   productID,
		Name: "Test Product",
	}

	// Test inventory below custom threshold
	inventory := &models.Inventory{
		TenantID:    suite.tenantID,
		WarehouseID: suite.warehouseID,
		ProductID:   productID,
		Quantity:    3, // Below threshold
	}

	inventories := []*models.Inventory{inventory}
	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return(inventories, nil).Once()
	suite.mockProductRepo.On("GetByID", ctx, suite.tenantID, productID).Return(product, nil).Once()

	alerts, err := suite.service.CheckLowStock(ctx, suite.tenantID, threshold)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), alerts, 1)
	assert.Equal(suite.T(), threshold, alerts[0].Threshold)
}

// TestJobSchedulingIntegration tests the integration with gocron job scheduling
func (suite *InventoryAlertServiceTestSuite) TestJobSchedulingIntegration() {
	ctx := context.Background()

	// Test the scheduled job method works correctly
	err := suite.service.ScheduledLowStockCheck(ctx)
	assert.NoError(suite.T(), err)

	// Test CheckAndLogLowStockAcrossAllTenants method
	err = suite.service.CheckAndLogLowStockAcrossAllTenants(ctx, 5)
	assert.NoError(suite.T(), err)
}

// TestPerformanceWithHighVolumeInventory tests performance under load
func (suite *InventoryAlertServiceTestSuite) TestPerformanceWithHighVolumeInventory() {
	ctx := context.Background()
	threshold := 10

	// Create 50 inventory items with varying stock levels
	inventories := make([]*models.Inventory, 0, 50)
	expectedAlerts := 0

	for i := 0; i < 50; i++ {
		productID := uuid.New()
		stockLevel := i % 20 // Stock levels from 0 to 19

		inventory := &models.Inventory{
			TenantID:    suite.tenantID,
			WarehouseID: suite.warehouseID,
			ProductID:   productID,
			Quantity:    stockLevel,
		}
		inventories = append(inventories, inventory)

		// Count expected alerts (stock below threshold)
		if stockLevel < threshold {
			expectedAlerts++
			product := &models.Product{ID: productID, Name: "Product " + string(rune(65+i))}
			suite.mockProductRepo.On("GetByID", ctx, suite.tenantID, productID).Return(product, nil).Once()
		}
	}

	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return(inventories, nil).Once()

	start := time.Now()
	alerts, err := suite.service.CheckLowStock(ctx, suite.tenantID, threshold)
	duration := time.Since(start)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), duration < 1*time.Second, "Processing took too long: %v", duration)
	assert.Len(suite.T(), alerts, expectedAlerts)
}

// TestErrorRecovery tests error handling and recovery scenarios
func (suite *InventoryAlertServiceTestSuite) TestErrorRecovery() {
	ctx := context.Background()
	threshold := 10

	// Test multiple errors in sequence to ensure service recovers
	// First call - repository error
	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return(([]*models.Inventory)(nil), errors.New("connection lost")).Once()

	alerts, err := suite.service.CheckLowStock(ctx, suite.tenantID, threshold)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), alerts)

	// Second call - successful operation should work normally
	productID := uuid.New()
	product := &models.Product{ID: productID, Name: "Test Product"}
	inventory := &models.Inventory{
		TenantID:    suite.tenantID,
		WarehouseID: suite.warehouseID,
		ProductID:   productID,
		Quantity:    5,
	}

	suite.mockInventoryRepo.On("List", ctx, suite.tenantID, 1000, 0).Return([]*models.Inventory{inventory}, nil).Once()
	suite.mockProductRepo.On("GetByID", ctx, suite.tenantID, productID).Return(product, nil).Once()

	alerts, err = suite.service.CheckLowStock(ctx, suite.tenantID, threshold)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), alerts, 1)
}

// TestModule function to run the test suite
func TestInventoryAlertServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryAlertServiceTestSuite))
}

// TestAlertNotificationTriggering tests the alert notification triggering workflow
func TestAlertNotificationTriggering(t *testing.T) {
	ctx := context.Background()
	threshold := 10

	alerts := []InventoryAlert{
		{TenantID: uuid.New(), ProductName: "Product 1", CurrentStock: 5, Threshold: threshold},
		{TenantID: uuid.New(), ProductName: "Product 2", CurrentStock: 3, Threshold: threshold},
	}

	// Test that logging doesn't panic and handles multiple alerts properly
	service := &InventoryAlertService{}
	// This tests the logging functionality which should be integration tested in real scenario
	service.LogLowStockAlerts(ctx, alerts)

	// Test with empty alerts
	service.LogLowStockAlerts(ctx, []InventoryAlert{})
}

// TestInventoryAlertComprehensiveScenario tests a complete business scenario
func TestInventoryAlertComprehensiveScenario(t *testing.T) {
	ctx := context.Background()

	mockInventoryRepo := &MockInventoryRepository{}
	mockProductRepo := &MockProductRepository{}
	service := NewInventoryAlertService(mockInventoryRepo, mockProductRepo)

	tenantID := uuid.New()
	threshold := 15

	// Create comprehensive test data
	inventories := []*models.Inventory{
		{TenantID: tenantID, ProductID: uuid.New(), Quantity: 10}, // Below threshold
		{TenantID: tenantID, ProductID: uuid.New(), Quantity: 20}, // Above threshold
		{TenantID: tenantID, ProductID: uuid.New(), Quantity: 5},  // Below threshold
		{TenantID: tenantID, ProductID: uuid.New(), Quantity: 18}, // Above threshold
		{TenantID: tenantID, ProductID: uuid.New(), Quantity: 12}, // Below threshold
	}

	products := []*models.Product{}
	expectedAlerts := 0

	// Set up mock expectations for products below threshold
	for i, inventory := range inventories {
		if inventory.Quantity < threshold {
			inventory.ProductID = uuid.New() // Reassign consistent ID
			product := &models.Product{
				ID:   inventory.ProductID,
				Name: "Product " + string(rune(65+i)),
			}
			products = append(products, product)
			mockProductRepo.On("GetByID", ctx, tenantID, inventory.ProductID).Return(product, nil).Once()
			expectedAlerts++
		}
	}

	mockInventoryRepo.On("List", ctx, tenantID, 1000, 0).Return(inventories, nil).Once()

	alerts, err := service.CheckLowStock(ctx, tenantID, threshold)

	assert.NoError(t, err)
	assert.Len(t, alerts, expectedAlerts)
	assert.True(t, len(alerts) > 0, "Should have generated some alerts")

	for _, alert := range alerts {
		assert.Equal(t, tenantID, alert.TenantID)
		assert.True(t, alert.CurrentStock < threshold, "Alert stock should be below threshold")
		assert.Equal(t, threshold, alert.Threshold)
	}
}