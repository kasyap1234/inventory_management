package services_test

	import (
	"context"
	"testing"
	"time"

	testhelpers "agromart2/testhelpers"
	"agromart2/internal/caching"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	. "agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	)

func TestOrderService(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	mockCache := &testhelpers.MockCacheService{}
	cacheService := caching.NewCacheService(mockCache)

	orderRepo := repositories.NewOrderRepository(db)
	inventoryRepo := repositories.NewInventoryRepository(db)
	productRepo := repositories.NewProductRepository(db)
	inventoryService := NewInventoryService(inventoryRepo, productRepo, cacheService)
	service := NewOrderService(orderRepo, inventoryRepo, inventoryService)

	ctx := context.Background()

	testDB := testhelpers.NewTestDB(t, "")
	tenantID := testDB.CreateTestTenant(t)

	t.Run("CreateOrder_PurchaseOrder_Success", func(t *testing.T) {
		product := testDB.CreateTestProductFull(tenantID)
		supplier := testDB.CreateTestSupplierFull(tenantID)
		warehouse := testDB.CreateTestWarehouseFull(tenantID)

		order := &models.Order{
			TenantID:    tenantID,
			ProductID:   product.ID,
			SupplierID:  &supplier.ID,
			WarehouseID: warehouse.ID,
			Quantity:    100,
			UnitPrice:   10.5,
			OrderType:   "purchase",
			Status:      "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, order.ID)

		// Verify order was created
		savedOrder, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, tenantID, savedOrder.TenantID)
		assert.Equal(t, product.ID, savedOrder.ProductID)
		assert.Equal(t, 100, savedOrder.Quantity)
		assert.Equal(t, "pending", savedOrder.Status)
	})

	t.Run("CreateOrder_SalesOrder_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		distributor := testhelpers.CreateTestDistributor(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		// Create inventory first
		testhelpers.CreateTestInventory(t, db, tenantID, warehouse.ID, product.ID, 200)

		order := &models.Order{
			TenantID:      tenantID,
			ProductID:     product.ID,
			DistributorID: &distributor.ID,
			WarehouseID:   warehouse.ID,
			Quantity:      50,
			UnitPrice:     15.0,
			OrderType:     "sales",
			Status:        "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, order.ID)
	})

	t.Run("CreateOrder_SalesOrder_InsufficientInventory", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		distributor := testhelpers.CreateTestDistributor(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		// Create inventory with insufficient quantity
		testhelpers.CreateTestInventory(t, db, tenantID, warehouse.ID, product.ID, 20)

		order := &models.Order{
			TenantID:      tenantID,
			ProductID:     product.ID,
			DistributorID: &distributor.ID,
			WarehouseID:   warehouse.ID,
			Quantity:      50, // More than available
			UnitPrice:     15.0,
			OrderType:     "sales",
			Status:        "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient inventory")
	})

	t.Run("CreateOrder_InvalidData", func(t *testing.T) {
		order := &models.Order{
			TenantID: tenantID,
			Quantity: -10, // Invalid quantity
			UnitPrice: -5.0, // Invalid price
		}

		err := service.CreateOrder(ctx, tenantID, order)
		assert.Error(t, err)
	})

	t.Run("GetOrderByID_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		order := &models.Order{
			TenantID:    tenantID,
			ProductID:   product.ID,
			SupplierID:  &supplier.ID,
			WarehouseID: warehouse.ID,
			Quantity:    25,
			UnitPrice:   12.5,
			OrderType:   "purchase",
			Status:      "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		retrievedOrder, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedOrder)
		assert.Equal(t, order.ID, retrievedOrder.ID)
		assert.Equal(t, 25, retrievedOrder.Quantity)
	})

	t.Run("GetOrderByID_WrongTenant", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)
		wrongTenantID := testhelpers.CreateTestTenant(t, db)

		order := &models.Order{
			TenantID:    tenantID,
			ProductID:   product.ID,
			SupplierID:  &supplier.ID,
			WarehouseID: warehouse.ID,
			Quantity:    30,
			UnitPrice:   8.75,
			OrderType:   "purchase",
			Status:      "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		retrievedOrder, err := service.GetOrderByID(ctx, wrongTenantID, order.ID)
		assert.NoError(t, err)
		assert.Nil(t, retrievedOrder)
	})

	t.Run("ListOrders_Success", func(t *testing.T) {
		// Create multiple orders
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		for i := 0; i < 3; i++ {
			order := &models.Order{
				TenantID:    tenantID,
				ProductID:   product.ID,
				SupplierID:  &supplier.ID,
				WarehouseID: warehouse.ID,
				Quantity:    10 + i,
				UnitPrice:   5.5 + float64(i),
				OrderType:   "purchase",
				Status:      "pending",
			}
			err := service.CreateOrder(ctx, tenantID, order)
			require.NoError(t, err)
		}

		orders, err := service.ListOrders(ctx, tenantID, 10, 0)
		assert.NoError(t, err)
		assert.True(t, len(orders) >= 3) // At least the orders we created
	})

	t.Run("UpdateOrder_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		order := &models.Order{
			TenantID:    tenantID,
			ProductID:   product.ID,
			SupplierID:  &supplier.ID,
			WarehouseID: warehouse.ID,
			Quantity:    20,
			UnitPrice:   9.99,
			OrderType:   "purchase",
			Status:      "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		// Update order
		order.Quantity = 25
		order.UnitPrice = 11.99
		notesValue := "Updated test order"
		order.Notes = &notesValue

		err = service.UpdateOrder(ctx, tenantID, order)
		assert.NoError(t, err)

		// Verify update
		updatedOrder, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, 25, updatedOrder.Quantity)
		assert.Equal(t, 11.99, updatedOrder.UnitPrice)
		assert.Equal(t, "Updated test order", updatedOrder.Notes)
	})

	t.Run("DeleteOrder_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		order := &models.Order{
			TenantID:    tenantID,
			ProductID:   product.ID,
			SupplierID:  &supplier.ID,
			WarehouseID: warehouse.ID,
			Quantity:    15,
			UnitPrice:   7.50,
			OrderType:   "purchase",
			Status:      "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		// Delete order
		err = service.DeleteOrder(ctx, tenantID, order.ID)
		assert.NoError(t, err)

		// Verify deletion
		deletedOrder, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.Nil(t, deletedOrder)
	})

	t.Run("ValidateStatusTransition_ValidTransitions", func(t *testing.T) {
		// Test valid transitions
		validTransitions := []struct {
			from, to string
		}{
			{"pending", "approved"},
			{"pending", "cancelled"},
			{"approved", "processing"},
			{"approved", "cancelled"},
			{"processing", "shipped"},
			{"processing", "cancelled"},
			{"shipped", "delivered"},
			{"shipped", "cancelled"},
		}

		for _, transition := range validTransitions {
			err := service.ValidateStatusTransition(transition.from, transition.to)
			assert.NoError(t, err, "Transition from %s to %s should be valid", transition.from, transition.to)
		}
	})

	t.Run("ValidateStatusTransition_InvalidTransitions", func(t *testing.T) {
		invalidTransitions := []struct {
			from, to string
		}{
			{"delivered", "cancelled"}, // Terminal state
			{"cancelled", "pending"},  // Terminal state
			{"shipped", "approved"},   // Backward transition
			{"processing", "approved"}, // Backward transition
		}

		for _, transition := range invalidTransitions {
			err := service.ValidateStatusTransition(transition.from, transition.to)
			assert.Error(t, err, "Transition from %s to %s should be invalid", transition.from, transition.to)
		}
	})

	t.Run("GetOrderAnalytics_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		// Create orders over time
		baseTime := time.Now().AddDate(0, 0, -7) // 7 days ago
		for i := 0; i < 5; i++ {
			orderDate := baseTime.AddDate(0, 0, i)
			order := &models.Order{
				TenantID:    tenantID,
				ProductID:   product.ID,
				SupplierID:  &supplier.ID,
				WarehouseID: warehouse.ID,
				Quantity:    10 + i,
				UnitPrice:   10.0 + float64(i),
				OrderType:   "purchase",
				Status:      "delivered",
				OrderDate:   orderDate,
				CreatedAt:   orderDate,
				UpdatedAt:   orderDate,
			}

			// Create order directly in repo to control dates
			err := repo.Create(ctx, order)
			require.NoError(t, err)
		}

		startDate := baseTime
		endDate := time.Now()

		analytics, err := service.GetOrderAnalytics(ctx, tenantID, startDate, endDate)
		assert.NoError(t, err)
		assert.NotNil(t, analytics)

		// Verify structure
		assert.Contains(t, analytics, "total_orders")
		assert.Contains(t, analytics, "total_value")
		assert.Contains(t, analytics, "status_breakdown")
		assert.Contains(t, analytics, "period")

		totalOrders := analytics["total_orders"].(int)
		assert.GreaterOrEqual(t, totalOrders, 5)

		totalValue := analytics["total_value"].(float64)
		assert.Greater(t, totalValue, 0.0)
	})

	t.Run("SearchOrders_BasicSearch", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		order := &models.Order{
			TenantID:    tenantID,
			ProductID:   product.ID,
			SupplierID:  &supplier.ID,
			WarehouseID: warehouse.ID,
			Quantity:    40,
			UnitPrice:   6.25,
			OrderType:   "purchase",
			Status:      "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		// Create search filter
		filter := &models.OrderSearchFilter{
			Query:  product.Name,
			Limit:  10,
			Offset: 0,
		}

		results, err := service.SearchOrders(ctx, tenantID, filter)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("ApproveOrder_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		order := &models.Order{
			TenantID:    tenantID,
			ProductID:   product.ID,
			SupplierID:  &supplier.ID,
			WarehouseID: warehouse.ID,
			Quantity:    30,
			UnitPrice:   13.75,
			OrderType:   "purchase",
			Status:      "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		// Approve order
		err = service.ApproveOrder(ctx, tenantID, order.ID)
		assert.NoError(t, err)

		// Verify status
		approvedOrder, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, "approved", approvedOrder.Status)
	})

	t.Run("ProcessOrder_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)
		distributor := testhelpers.CreateTestDistributor(t, db, tenantID)

		// Create sufficient inventory
		testhelpers.CreateTestInventory(t, db, tenantID, warehouse.ID, product.ID, 100)

		order := &models.Order{
			TenantID:      tenantID,
			ProductID:     product.ID,
			DistributorID: &distributor.ID,
			WarehouseID:   warehouse.ID,
			Quantity:      25,
			UnitPrice:     17.50,
			OrderType:     "sales",
			Status:        "approved",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		// Process order (should reduce inventory)
		err = service.ProcessOrder(ctx, tenantID, order.ID)
		assert.NoError(t, err)

		// Verify order status
		orderAfter, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, "processing", orderAfter.Status)

		// Verify inventory was reduced
		inventory, err := inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, warehouse.ID, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, 75, inventory.Quantity) // 100 - 25
	})

	t.Run("ReceiveOrder_PurchaseOrder_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		order := &models.Order{
			TenantID:    tenantID,
			ProductID:   product.ID,
			SupplierID:  &supplier.ID,
			WarehouseID: warehouse.ID,
			Quantity:    35,
			UnitPrice:   9.99,
			OrderType:   "purchase",
			Status:      "processing",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		// Receive order (should add to inventory)
		err = service.ReceiveOrder(ctx, tenantID, order.ID)
		assert.NoError(t, err)

		// Verify order status
		orderAfter, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, "delivered", orderAfter.Status)

		// Verify inventory was increased
		inventory, err := inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, warehouse.ID, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, 35, inventory.Quantity)
	})

	t.Run("ShipOrder_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		distributor := testhelpers.CreateTestDistributor(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		order := &models.Order{
			TenantID:      tenantID,
			ProductID:     product.ID,
			DistributorID: &distributor.ID,
			WarehouseID:   warehouse.ID,
			Quantity:      18,
			UnitPrice:     22.00,
			OrderType:     "sales",
			Status:        "processing",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		expectedDelivery := time.Now().AddDate(0, 0, 3)

		// Ship order
		err = service.ShipOrder(ctx, tenantID, order.ID, &expectedDelivery)
		assert.NoError(t, err)

		// Verify order status
		orderAfter, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, "shipped", orderAfter.Status)
		assert.True(t, orderAfter.ExpectedDelivery.Equal(expectedDelivery))
	})

	t.Run("DeliverOrder_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		distributor := testhelpers.CreateTestDistributor(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		order := &models.Order{
			TenantID:      tenantID,
			ProductID:     product.ID,
			DistributorID: &distributor.ID,
			WarehouseID:   warehouse.ID,
			Quantity:      12,
			UnitPrice:     19.99,
			OrderType:     "sales",
			Status:        "shipped",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		// Deliver order
		err = service.DeliverOrder(ctx, tenantID, order.ID)
		assert.NoError(t, err)

		// Verify order status
		orderAfter, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, "delivered", orderAfter.Status)
	})

	t.Run("CancelOrder_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		distributor := testhelpers.CreateTestDistributor(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		// Create inventory
		testhelpers.CreateTestInventory(t, db, tenantID, warehouse.ID, product.ID, 50)

		order := &models.Order{
			TenantID:      tenantID,
			ProductID:     product.ID,
			DistributorID: &distributor.ID,
			WarehouseID:   warehouse.ID,
			Quantity:      20,
			UnitPrice:     14.50,
			OrderType:     "sales",
			Status:        "approved",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		// Process order first (reduces inventory)
		err = service.ProcessOrder(ctx, tenantID, order.ID)
		require.NoError(t, err)

		// Cancel order (should restore inventory)
		err = service.CancelOrder(ctx, tenantID, order.ID)
		assert.NoError(t, err)

		// Verify order status
		orderAfter, err := service.GetOrderByID(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.Equal(t, "cancelled", orderAfter.Status)

		// Verify inventory was restored
		inventory, err := inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, warehouse.ID, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, 50, inventory.Quantity) // Should be back to original
	})

	t.Run("GetOrderHistory_Success", func(t *testing.T) {
		product := testhelpers.CreateTestProduct(t, db, tenantID)
		supplier := testhelpers.CreateTestSupplier(t, db, tenantID)
		warehouse := testhelpers.CreateTestWarehouse(t, db, tenantID)

		order := &models.Order{
			TenantID:    tenantID,
			ProductID:   product.ID,
			SupplierID:  &supplier.ID,
			WarehouseID: warehouse.ID,
			Quantity:    45,
			UnitPrice:   8.80,
			OrderType:   "purchase",
			Status:      "pending",
		}

		err := service.CreateOrder(ctx, tenantID, order)
		require.NoError(t, err)

		history, err := service.GetOrderHistory(ctx, tenantID, order.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, history)
		assert.Equal(t, order.ID, history[0].ID)
	})
}
