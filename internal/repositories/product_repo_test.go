package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection for repository tests
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL environment variable not set, skipping repository tests")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Unable to connect to test database: %v", err)
	}

	// Ping the database to ensure a good connection
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return pool
}

// truncateTestTables cleans up test data
func truncateTestTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	tables := []string{"products", "categories"}
	for _, table := range tables {
		_, err := pool.Exec(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE")
		if err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}

func TestProductRepository(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	// Setup: truncate tables for clean state
	truncateTestTables(t, pool)
	defer truncateTestTables(t, pool) // Cleanup after test

	repo := NewProductRepo(pool)

	ctx := context.Background()
	tenantID := uuid.New()

	// Setup test data
	categoryID := uuid.New()
	category := &models.Category{
		ID:          categoryID,
		TenantID:    tenantID,
		Name:        "Test Category",
		Description: "Test category description",
	}
	categoryRepo := NewCategoryRepo(pool)
	err := categoryRepo.Create(ctx, category)
	require.NoError(t, err)
	defer categoryRepo.Delete(ctx, tenantID, categoryID)

	t.Run("Create", func(t *testing.T) {
		batchNum := "BATCH001"
		expiry := time.Now().Add(30 * 24 * time.Hour)
		barcode := "123456789"
		uom := "pcs"
		desc := "Test product description"

		product := &models.Product{
			ID:             uuid.New(),
			TenantID:       tenantID,
			CategoryID:     &categoryID,
			Name:           "Test Product",
			BatchNumber:    &batchNum,
			ExpiryDate:     &expiry,
			Quantity:       100,
			UnitPrice:      29.99,
			Barcode:        &barcode,
			UnitOfMeasure:  &uom,
			Description:    &desc,
		}

		err := repo.Create(ctx, product)
		assert.NoError(t, err)

		// Verify creation
		retrieved, err := repo.GetByID(ctx, tenantID, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, product.Name, retrieved.Name)
		assert.Equal(t, *product.Barcode, *retrieved.Barcode)
		assert.Equal(t, product.Quantity, retrieved.Quantity)
		assert.Equal(t, product.UnitPrice, retrieved.UnitPrice)
	})

	t.Run("GetByID", func(t *testing.T) {
		product := &models.Product{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Product for GetByID",
			Quantity:  50,
			UnitPrice: 19.99,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, tenantID, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, product.ID, retrieved.ID)
		assert.Equal(t, product.Name, retrieved.Name)

		// Test with wrong tenant
		wrongTenantID := uuid.New()
		_, err = repo.GetByID(ctx, wrongTenantID, product.ID)
		assert.Error(t, err)

		// Test with non-existent ID
		_, err = repo.GetByID(ctx, tenantID, uuid.New())
		assert.Error(t, err)
	})

	t.Run("GetByBarcode", func(t *testing.T) {
		barcode := "BARCODE123"
		product := &models.Product{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Product by Barcode",
			Barcode:   &barcode,
			Quantity:  25,
			UnitPrice: 9.99,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)

		retrieved, err := repo.GetByBarcode(ctx, tenantID, barcode)
		assert.NoError(t, err)
		assert.Equal(t, product.ID, retrieved.ID)
		assert.Equal(t, barcode, *retrieved.Barcode)

		// Test with wrong tenant
		wrongTenantID := uuid.New()
		_, err = repo.GetByBarcode(ctx, wrongTenantID, barcode)
		assert.Error(t, err)

		// Test with non-existent barcode
		_, err = repo.GetByBarcode(ctx, tenantID, "NONEXISTENT")
		assert.Error(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		product := &models.Product{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Original Name",
			Quantity:  10,
			UnitPrice: 5.99,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)

		// Update product
		product.Name = "Updated Name"
		product.Quantity = 20
		product.UnitPrice = 15.99
		err = repo.Update(ctx, product)
		assert.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetByID(ctx, tenantID, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
		assert.Equal(t, 20, retrieved.Quantity)
		assert.Equal(t, 15.99, retrieved.UnitPrice)
	})

	t.Run("Delete", func(t *testing.T) {
		product := &models.Product{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Product to Delete",
			Quantity:  5,
			UnitPrice: 1.99,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)

		// Verify exists
		_, err = repo.GetByID(ctx, tenantID, product.ID)
		assert.NoError(t, err)

		// Delete
		err = repo.Delete(ctx, tenantID, product.ID)
		assert.NoError(t, err)

		// Verify deleted
		_, err = repo.GetByID(ctx, tenantID, product.ID)
		assert.Error(t, err)
	})

	t.Run("List", func(t *testing.T) {
		// Create multiple products
		for i := 0; i < 5; i++ {
			product := &models.Product{
				ID:        uuid.New(),
				TenantID:  tenantID,
				Name:      fmt.Sprintf("List Product %d", i),
				Quantity:  i + 1,
				UnitPrice: float64(i+1) * 2.99,
			}
			err := repo.Create(ctx, product)
			require.NoError(t, err)
		}

		products, err := repo.List(ctx, tenantID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(products), 5)

		// Test pagination
		limited, err := repo.List(ctx, tenantID, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, limited, 2)
	})

	t.Run("Search", func(t *testing.T) {
		// Create searchable products
		barcode1 := "APPLE001"
		barcode2 := "ORANGE001"
		barcode3 := "APPLE002"

		product1 := &models.Product{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Apple Juice",
			Barcode:   &barcode1,
			Quantity:  10,
			UnitPrice: 3.99,
		}
		product2 := &models.Product{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Orange Juice",
			Barcode:   &barcode2,
			Quantity:  15,
			CategoryID: &categoryID,
			UnitPrice:  4.49,
		}
		product3 := &models.Product{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Green Apple",
			Barcode:   &barcode3,
			Quantity:  8,
			UnitPrice: 2.99,
		}

		err := repo.Create(ctx, product1)
		require.NoError(t, err)
		err = repo.Create(ctx, product2)
		require.NoError(t, err)
		err = repo.Create(ctx, product3)
		require.NoError(t, err)

		// Search by name
		results, err := repo.Search(ctx, tenantID, "Apple", nil, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)

		// Search by barcode
		results, err = repo.Search(ctx, tenantID, "APPLE001", nil, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Apple Juice", results[0].Name)

		// Search with category filter
		results, err = repo.Search(ctx, tenantID, "Juice", &categoryID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})

	t.Run("ListWithCategory", func(t *testing.T) {
		products, err := repo.ListWithCategory(ctx, tenantID, &categoryID, 10, 0)
		assert.NoError(t, err)
		assert.NotNil(t, products)

		// Test with nil category (should list all)
		allProducts, err := repo.ListWithCategory(ctx, tenantID, nil, 10, 0)
		assert.NoError(t, err)
		assert.NotNil(t, allProducts)
	})

	t.Run("CategoryAnalytics", func(t *testing.T) {
		analytics, err := repo.CategoryAnalytics(ctx, tenantID)
		assert.NoError(t, err)
		assert.NotNil(t, analytics)

		// Analytics should contain data for existing categories
		total := 0
		for _, count := range analytics {
			total += count
		}
		assert.Greater(t, total, 0)
	})

	t.Run("AdvancedSearch", func(t *testing.T) {
		// Create test data with various attributes
		urgentBarcode := "URGENT001"
		urgentExpiry := time.Now().Add(24 * time.Hour)

		urgentProduct := &models.Product{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Name:       "Urgent Product",
			Quantity:   1,
			UnitPrice:  100.00,
			Barcode:    &urgentBarcode,
			ExpiryDate: &urgentExpiry,
		}
		err := repo.Create(ctx, urgentProduct)
		require.NoError(t, err)

		filter := &models.ProductSearchFilter{
			Query:    "",
			Limit:    10,
			SortBy:   "created_at",
			SortOrder: "desc",
		}

		results, err := repo.AdvancedSearch(ctx, tenantID, filter)
		assert.NoError(t, err)
		assert.NotNil(t, results)

		// Test filters
		minQty := 5
		filter.MinQuantity = &minQty
		results, err = repo.AdvancedSearch(ctx, tenantID, filter)
		assert.NoError(t, err)

		minPrice := 50.0
		filter.MinPrice = &minPrice
		results, err = repo.AdvancedSearch(ctx, tenantID, filter)
		assert.NoError(t, err)

		barcodeFilter := "URGENT001"
		filter.Barcode = &barcodeFilter
		results, err = repo.AdvancedSearch(ctx, tenantID, filter)
		assert.NoError(t, err)
		// Should find the urgent product if filters match
	})

	// Cleanup is handled by defer statement above
}
