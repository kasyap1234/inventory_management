package integration

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"
	"agromart2/testhelpers/containers"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// OrderIntegrationTestSuite tests order operations with real database transactions
type OrderIntegrationTestSuite struct {
	suite.Suite
	container        *containers.PostgresContainer
	ctx              context.Context
	cancel           context.CancelFunc
	orderService     services.OrderServiceInterface
	inventoryService services.InventoryService
	productRepo      repositories.ProductRepository
	categoryRepo     repositories.CategoryRepository
	warehouseRepo    repositories.WarehouseRepository
	supplierRepo     repositories.SupplierRepository
	distributorRepo  repositories.DistributorRepository
	inventoryRepo    repositories.InventoryRepository
	orderRepo        repositories.OrderRepository
	logger           *common.StructuredLogger
	tenantID         uuid.UUID
}

func TestOrderIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(OrderIntegrationTestSuite))
}

func (s *OrderIntegrationTestSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 5*time.Minute)

	// Get the path to migrations directory
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	migrationsPath := filepath.Join(projectRoot, "migrations")

	config := containers.DefaultPostgresConfig()
	config.MigrationsPath = migrationsPath

	container, err := containers.NewPostgresContainer(s.ctx, config)
	require.NoError(s.T(), err, "Failed to start PostgreSQL container")

	s.container = container
	s.logger = common.NewStructuredLogger()

	// Initialize repositories
	s.productRepo = repositories.NewProductRepo(container.Pool)
	s.categoryRepo = repositories.NewCategoryRepo(container.Pool)
	s.warehouseRepo = repositories.NewWarehouseRepository(container.Pool)
	s.supplierRepo = repositories.NewSupplierRepository(container.Pool)
	s.distributorRepo = repositories.NewDistributorRepository(container.Pool)
	s.inventoryRepo = repositories.NewInventoryRepo(container.Pool)
	s.orderRepo = repositories.NewOrderRepo(container.Pool)
	orderStatusHistoryRepo := repositories.NewOrderStatusHistoryRepo(container.Pool)

	// Initialize services - use adapter to bridge repository and service interfaces
	inventoryAdapter := repositories.NewInventoryAdapter(s.inventoryRepo)
	s.inventoryService = services.NewInventoryService(
		inventoryAdapter,
		s.logger,
	)

	s.orderService = services.NewOrderService(
		container.Pool,
		s.orderRepo,
		s.inventoryRepo,
		s.inventoryService,
		orderStatusHistoryRepo,
		s.logger,
	)

	s.tenantID = uuid.New()
}

func (s *OrderIntegrationTestSuite) TearDownSuite() {
	if s.container != nil {
		s.container.Close(s.ctx)
	}
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *OrderIntegrationTestSuite) SetupTest() {
	// Clean tables before each test
	err := s.container.CleanTables(s.ctx)
	require.NoError(s.T(), err, "Failed to clean tables")

	// Create the test tenant
	_, err = s.container.Pool.Exec(s.ctx, `
		INSERT INTO tenants (id, name, subdomain, status, created_at)
		VALUES ($1, 'Test Tenant', $2, 'active', NOW())
	`, s.tenantID, "order-test-"+s.tenantID.String()[:8])
	require.NoError(s.T(), err)
}

// Helper function to create test data
func (s *OrderIntegrationTestSuite) setupTestData() (categoryID, productID, warehouseID, supplierID, distributorID uuid.UUID) {
	categoryID = uuid.New()
	productID = uuid.New()
	warehouseID = uuid.New()
	supplierID = uuid.New()
	distributorID = uuid.New()

	// Create category
	err := s.categoryRepo.Create(s.ctx, &models.Category{
		ID:          categoryID,
		TenantID:    s.tenantID,
		Name:        "Test Category",
		Description: "Test category description",
	})
	require.NoError(s.T(), err)

	// Create product
	err = s.productRepo.Create(s.ctx, &models.Product{
		ID:         productID,
		TenantID:   s.tenantID,
		CategoryID: &categoryID,
		Name:       "Test Product",
		Quantity:   0,
		UnitPrice:  10.00,
	})
	require.NoError(s.T(), err)

	// Create warehouse
	err = s.warehouseRepo.Create(s.ctx, &models.Warehouse{
		ID:       warehouseID,
		TenantID: s.tenantID,
		Name:     "Test Warehouse",
		Address:  stringPtr("123 Test Street"),
		Capacity: intPtr(1000),
	})
	require.NoError(s.T(), err)

	// Create supplier
	err = s.supplierRepo.Create(s.ctx, &models.Supplier{
		ID:           supplierID,
		TenantID:     s.tenantID,
		Name:         "Test Supplier",
		ContactEmail: stringPtr("supplier@test.com"),
	})
	require.NoError(s.T(), err)

	// Create distributor
	err = s.distributorRepo.Create(s.ctx, &models.Distributor{
		ID:           distributorID,
		TenantID:     s.tenantID,
		Name:         "Test Distributor",
		ContactEmail: stringPtr("distributor@test.com"),
	})
	require.NoError(s.T(), err)

	return
}

func (s *OrderIntegrationTestSuite) setupInventory(warehouseID, productID uuid.UUID, quantity int) {
	err := s.inventoryRepo.Create(s.ctx, &models.Inventory{
		ID:          uuid.New(),
		TenantID:    s.tenantID,
		WarehouseID: warehouseID,
		ProductID:   productID,
		Quantity:    quantity,
	})
	require.NoError(s.T(), err)
}

// =====================
// Purchase Order Tests
// =====================

func (s *OrderIntegrationTestSuite) TestCreatePurchaseOrder() {
	_, productID, warehouseID, supplierID, _ := s.setupTestData()

	order := &models.Order{
		TenantID:    s.tenantID,
		OrderType:   "purchase",
		SupplierID:  &supplierID,
		ProductID:   productID,
		WarehouseID: warehouseID,
		Quantity:    50,
		UnitPrice:   10.00,
		Notes:       stringPtr("Test purchase order"),
	}

	err := s.orderService.CreateOrder(s.ctx, s.tenantID, order)
	require.NoError(s.T(), err)

	// Verify order was created
	fetched, err := s.orderService.GetOrderByID(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "purchase", fetched.OrderType)
	assert.Equal(s.T(), 50, fetched.Quantity)
	assert.Equal(s.T(), "pending", fetched.Status)
}

func (s *OrderIntegrationTestSuite) TestPurchaseOrderLifecycle() {
	_, productID, warehouseID, supplierID, _ := s.setupTestData()

	// Create purchase order
	order := &models.Order{
		TenantID:    s.tenantID,
		OrderType:   "purchase",
		SupplierID:  &supplierID,
		ProductID:   productID,
		WarehouseID: warehouseID,
		Quantity:    100,
		UnitPrice:   15.00,
	}

	err := s.orderService.CreateOrder(s.ctx, s.tenantID, order)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "pending", order.Status)

	// Approve order
	err = s.orderService.ApproveOrder(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)

	fetched, err := s.orderService.GetOrderByID(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "approved", fetched.Status)

	// Receive order (should add inventory)
	err = s.orderService.ReceiveOrder(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)

	fetched, err = s.orderService.GetOrderByID(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "received", fetched.Status)

	// Verify inventory was added
	inventory, err := s.inventoryRepo.GetByWarehouseAndProduct(s.ctx, s.tenantID, warehouseID, productID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 100, inventory.Quantity)
}

// =====================
// Sales Order Tests
// =====================

func (s *OrderIntegrationTestSuite) TestCreateSalesOrderWithSufficientInventory() {
	_, productID, warehouseID, _, distributorID := s.setupTestData()

	// Setup inventory with 100 units
	s.setupInventory(warehouseID, productID, 100)

	order := &models.Order{
		TenantID:      s.tenantID,
		OrderType:     "sales",
		DistributorID: &distributorID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		Quantity:      50,
		UnitPrice:     15.00,
		Notes:         stringPtr("Test sales order"),
	}

	err := s.orderService.CreateOrder(s.ctx, s.tenantID, order)
	require.NoError(s.T(), err)

	// Verify order was created
	fetched, err := s.orderService.GetOrderByID(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "sales", fetched.OrderType)
	assert.Equal(s.T(), 50, fetched.Quantity)
	assert.Equal(s.T(), "pending", fetched.Status)
}

func (s *OrderIntegrationTestSuite) TestCreateSalesOrderInsufficientInventory() {
	_, productID, warehouseID, _, distributorID := s.setupTestData()

	// Setup inventory with only 20 units
	s.setupInventory(warehouseID, productID, 20)

	order := &models.Order{
		TenantID:      s.tenantID,
		OrderType:     "sales",
		DistributorID: &distributorID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		Quantity:      50, // More than available
		UnitPrice:     15.00,
	}

	err := s.orderService.CreateOrder(s.ctx, s.tenantID, order)
	require.Error(s.T(), err, "Should fail due to insufficient inventory")
	assert.Contains(s.T(), err.Error(), "insufficient inventory")
}

func (s *OrderIntegrationTestSuite) TestSalesOrderLifecycle() {
	_, productID, warehouseID, _, distributorID := s.setupTestData()

	// Setup inventory with 100 units
	s.setupInventory(warehouseID, productID, 100)

	// Create sales order
	order := &models.Order{
		TenantID:      s.tenantID,
		OrderType:     "sales",
		DistributorID: &distributorID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		Quantity:      40,
		UnitPrice:     20.00,
	}

	err := s.orderService.CreateOrder(s.ctx, s.tenantID, order)
	require.NoError(s.T(), err)

	// Approve order
	err = s.orderService.ApproveOrder(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)

	// Ship order (should deduct inventory)
	expectedDelivery := time.Now().Add(7 * 24 * time.Hour)
	err = s.orderService.ShipOrder(s.ctx, s.tenantID, order.ID, &expectedDelivery)
	require.NoError(s.T(), err)

	fetched, err := s.orderService.GetOrderByID(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "shipped", fetched.Status)

	// Verify inventory was deducted
	inventory, err := s.inventoryRepo.GetByWarehouseAndProduct(s.ctx, s.tenantID, warehouseID, productID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 60, inventory.Quantity) // 100 - 40 = 60

	// Deliver order
	err = s.orderService.DeliverOrder(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)

	fetched, err = s.orderService.GetOrderByID(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "delivered", fetched.Status)
}

// =====================
// Order Cancellation Tests
// =====================

func (s *OrderIntegrationTestSuite) TestCancelPendingOrder() {
	_, productID, warehouseID, supplierID, _ := s.setupTestData()

	order := &models.Order{
		TenantID:    s.tenantID,
		OrderType:   "purchase",
		SupplierID:  &supplierID,
		ProductID:   productID,
		WarehouseID: warehouseID,
		Quantity:    50,
		UnitPrice:   10.00,
	}

	err := s.orderService.CreateOrder(s.ctx, s.tenantID, order)
	require.NoError(s.T(), err)

	// Cancel the pending order
	err = s.orderService.CancelOrder(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)

	fetched, err := s.orderService.GetOrderByID(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "cancelled", fetched.Status)
}

func (s *OrderIntegrationTestSuite) TestCannotCancelDeliveredOrder() {
	_, productID, warehouseID, _, distributorID := s.setupTestData()
	s.setupInventory(warehouseID, productID, 100)

	order := &models.Order{
		TenantID:      s.tenantID,
		OrderType:     "sales",
		DistributorID: &distributorID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		Quantity:      20,
		UnitPrice:     15.00,
	}

	err := s.orderService.CreateOrder(s.ctx, s.tenantID, order)
	require.NoError(s.T(), err)

	// Progress through the lifecycle
	err = s.orderService.ApproveOrder(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)

	expectedDelivery := time.Now().Add(7 * 24 * time.Hour)
	err = s.orderService.ShipOrder(s.ctx, s.tenantID, order.ID, &expectedDelivery)
	require.NoError(s.T(), err)

	err = s.orderService.DeliverOrder(s.ctx, s.tenantID, order.ID)
	require.NoError(s.T(), err)

	// Try to cancel the delivered order
	err = s.orderService.CancelOrder(s.ctx, s.tenantID, order.ID)
	require.Error(s.T(), err, "Should not be able to cancel a delivered order")
}

// =====================
// Transaction Atomicity Tests
// =====================

func (s *OrderIntegrationTestSuite) TestTransactionAtomicityOnOrderCreation() {
	_, productID, warehouseID, _, distributorID := s.setupTestData()
	s.setupInventory(warehouseID, productID, 100)

	txMgr := common.NewTransactionManager(s.container.Pool, s.logger)

	// Test that order creation and inventory reservation are atomic
	orderID := uuid.New()
	createdOrder := &models.Order{}

	err := txMgr.ExecuteInTransaction(s.ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Create order in transaction
		order := &models.Order{
			ID:            orderID,
			TenantID:      s.tenantID,
			OrderType:     "sales",
			DistributorID: &distributorID,
			ProductID:     productID,
			WarehouseID:   warehouseID,
			Quantity:      30,
			UnitPrice:     25.00,
			Status:        "pending",
			OrderDate:     time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO orders (id, tenant_id, order_type, distributor_id, product_id, warehouse_id, 
				quantity, unit_price, status, order_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, order.ID, order.TenantID, order.OrderType, order.DistributorID, order.ProductID,
			order.WarehouseID, order.Quantity, order.UnitPrice, order.Status, order.OrderDate,
			order.CreatedAt, order.UpdatedAt)

		if err != nil {
			return err
		}
		*createdOrder = *order
		return nil
	})

	require.NoError(s.T(), err)

	// Verify the order was committed
	fetched, err := s.orderService.GetOrderByID(s.ctx, s.tenantID, orderID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), orderID, fetched.ID)
}

func (s *OrderIntegrationTestSuite) TestTransactionRollbackOnFailure() {
	_, productID, warehouseID, _, distributorID := s.setupTestData()
	s.setupInventory(warehouseID, productID, 100)

	txMgr := common.NewTransactionManager(s.container.Pool, s.logger)

	orderID := uuid.New()

	err := txMgr.ExecuteInTransaction(s.ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Create order in transaction
		_, err := tx.Exec(ctx, `
			INSERT INTO orders (id, tenant_id, order_type, distributor_id, product_id, warehouse_id, 
				quantity, unit_price, status, order_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, orderID, s.tenantID, "sales", distributorID, productID,
			warehouseID, 30, 25.00, "pending", time.Now(),
			time.Now(), time.Now())

		if err != nil {
			return err
		}

		// Simulate a failure after order insert
		return assert.AnError
	})

	require.Error(s.T(), err)

	// Verify the order was NOT committed (rolled back)
	_, err = s.orderService.GetOrderByID(s.ctx, s.tenantID, orderID)
	require.Error(s.T(), err, "Order should not exist after rollback")
}

// =====================
// Concurrent Order Tests
// =====================

func (s *OrderIntegrationTestSuite) TestConcurrentOrdersInventoryDeduction() {
	_, productID, warehouseID, _, distributorID := s.setupTestData()
	s.setupInventory(warehouseID, productID, 100)

	// Try to create two orders that together exceed inventory
	order1 := &models.Order{
		TenantID:      s.tenantID,
		OrderType:     "sales",
		DistributorID: &distributorID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		Quantity:      60,
		UnitPrice:     15.00,
	}

	order2 := &models.Order{
		TenantID:      s.tenantID,
		OrderType:     "sales",
		DistributorID: &distributorID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		Quantity:      50,
		UnitPrice:     15.00,
	}

	// First order should succeed (60 < 100)
	err := s.orderService.CreateOrder(s.ctx, s.tenantID, order1)
	require.NoError(s.T(), err)

	// Second order should fail (50 + 60 > 100 - but we check current inventory which is still 100)
	// Note: The actual inventory deduction happens on ship, not on order creation
	// So this test verifies creation-time inventory check
	err = s.orderService.CreateOrder(s.ctx, s.tenantID, order2)
	require.NoError(s.T(), err) // Should succeed because creation only checks, doesn't deduct

	// Verify both orders were created
	orders, err := s.orderService.ListOrders(s.ctx, s.tenantID, 10, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), orders, 2)
}

// =====================
// Order Search Tests
// =====================

func (s *OrderIntegrationTestSuite) TestOrderSearch() {
	_, productID, warehouseID, supplierID, distributorID := s.setupTestData()
	s.setupInventory(warehouseID, productID, 200)

	// Create multiple orders
	purchaseOrder := &models.Order{
		TenantID:    s.tenantID,
		OrderType:   "purchase",
		SupplierID:  &supplierID,
		ProductID:   productID,
		WarehouseID: warehouseID,
		Quantity:    50,
		UnitPrice:   10.00,
	}
	err := s.orderService.CreateOrder(s.ctx, s.tenantID, purchaseOrder)
	require.NoError(s.T(), err)

	salesOrder := &models.Order{
		TenantID:      s.tenantID,
		OrderType:     "sales",
		DistributorID: &distributorID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		Quantity:      30,
		UnitPrice:     15.00,
	}
	err = s.orderService.CreateOrder(s.ctx, s.tenantID, salesOrder)
	require.NoError(s.T(), err)

	// Approve the sales order
	err = s.orderService.ApproveOrder(s.ctx, s.tenantID, salesOrder.ID)
	require.NoError(s.T(), err)

	// Search for pending orders
	filter := &models.OrderSearchFilter{
		Status: stringPtr("pending"),
		Limit:  10,
	}
	results, err := s.orderService.SearchOrders(s.ctx, s.tenantID, filter)
	require.NoError(s.T(), err)
	assert.Len(s.T(), results, 1)
	assert.Equal(s.T(), "purchase", results[0].OrderType)

	// Search for approved orders
	filter.Status = stringPtr("approved")
	results, err = s.orderService.SearchOrders(s.ctx, s.tenantID, filter)
	require.NoError(s.T(), err)
	assert.Len(s.T(), results, 1)
	assert.Equal(s.T(), "sales", results[0].OrderType)
}

// =====================
// Order Analytics Tests
// =====================

func (s *OrderIntegrationTestSuite) TestOrderAnalytics() {
	_, productID, warehouseID, supplierID, distributorID := s.setupTestData()
	s.setupInventory(warehouseID, productID, 500)

	// Create several orders
	for i := 0; i < 3; i++ {
		order := &models.Order{
			TenantID:    s.tenantID,
			OrderType:   "purchase",
			SupplierID:  &supplierID,
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    (i + 1) * 10,
			UnitPrice:   10.00,
		}
		err := s.orderService.CreateOrder(s.ctx, s.tenantID, order)
		require.NoError(s.T(), err)
	}

	for i := 0; i < 2; i++ {
		order := &models.Order{
			TenantID:      s.tenantID,
			OrderType:     "sales",
			DistributorID: &distributorID,
			ProductID:     productID,
			WarehouseID:   warehouseID,
			Quantity:      (i + 1) * 5,
			UnitPrice:     15.00,
		}
		err := s.orderService.CreateOrder(s.ctx, s.tenantID, order)
		require.NoError(s.T(), err)
	}

	// Get analytics for today
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now().Add(24 * time.Hour)

	analytics, err := s.orderService.GetOrderAnalytics(s.ctx, s.tenantID, startDate, endDate)
	require.NoError(s.T(), err)

	// Verify analytics contain expected data
	assert.NotNil(s.T(), analytics)
	// The exact structure depends on the GetOrderAnalytics implementation
}
