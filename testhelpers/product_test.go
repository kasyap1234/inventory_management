package testhelpers

import (
	"context"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := SetupTestDB(t, "")
	defer testDB.Cleanup()

	// Setup test data
	tenantID := SetupTestTenant(t, testDB)
	categoryID := SetupTestCategory(t, testDB, tenantID)

	// Initialize repository
	repo := repositories.NewProductRepo(testDB.Pool)

	t.Run("Create", func(t *testing.T) {
		product := &models.Product{
			ID:             uuid.New(),
			TenantID:       tenantID,
			CategoryID:     &categoryID,
			Name:           "Test Fertilizer",
			BatchNumber:    stringPtr("BATCH001"),
			ExpiryDate:     timePtr(time.Now().Add(365 * 24 * time.Hour)),
			Quantity:       500,
			UnitPrice:      15.99,
			Barcode:        stringPtr("123456789012"),
			UnitOfMeasure:  stringPtr("kg"),
			Description:    stringPtr("High-quality fertilizer"),
		}

		err := repo.Create(context.Background(), product)
		require.NoError(t, err)

		// Verify creation
		created, err := repo.GetByID(context.Background(), tenantID, product.ID)
		require.NoError(t, err)
		assert.Equal(t, product.Name, created.Name)
		assert.Equal(t, product.Quantity, created.Quantity)
	})

	t.Run("GetByID", func(t *testing.T) {
		// Create product first
		productID := uuid.New()
		product := &models.Product{
			ID:             productID,
			TenantID:       tenantID,
			CategoryID:     &categoryID,
			Name:           "Pesticide X",
			Quantity:       100,
			UnitPrice:      25.50,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)

		// Test GetByID
		retrieved, err := repo.GetByID(context.Background(), tenantID, productID)
		require.NoError(t, err)
		assert.Equal(t, productID, retrieved.ID)
		assert.Equal(t, "Pesticide X", retrieved.Name)

		// Test non-existent ID
		_, err = repo.GetByID(context.Background(), tenantID, uuid.New())
		assert.Error(t, err)
	})

	t.Run("GetByBarcode", func(t *testing.T) {
		barcode := "987654321098"
		product := &models.Product{
			ID:             uuid.New(),
			TenantID:       tenantID,
			CategoryID:     &categoryID,
			Name:           "Seed Y",
			Barcode:        &barcode,
			Quantity:       1000,
			UnitPrice:      5.99,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)

		// Test GetByBarcode
		retrieved, err := repo.GetByBarcode(context.Background(), tenantID, barcode)
		require.NoError(t, err)
		assert.Equal(t, barcode, *retrieved.Barcode)

		// Test non-existent barcode
		_, err = repo.GetByBarcode(context.Background(), tenantID, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("List", func(t *testing.T) {
		// Ensure at least one product exists
		product := &models.Product{
			ID:             uuid.New(),
			TenantID:       tenantID,
			Name:           "List Test Item",
			Quantity:       50,
			UnitPrice:      12.99,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)

		// Test List
		products, err := repo.List(context.Background(), tenantID, 10, 0)
		require.NoError(t, err)
		assert.True(t, len(products) > 0)

		// Verify tenant isolation
		otherTenantID := SetupTestTenant(t, testDB)
		otherProducts, err := repo.List(context.Background(), otherTenantID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, otherProducts, 0)
	})

	t.Run("Update", func(t *testing.T) {
		product := &models.Product{
			ID:             uuid.New(),
			TenantID:       tenantID,
			Name:           "Update Test Item",
			Quantity:       200,
			UnitPrice:      8.99,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)

		// Update product
		product.Name = "Updated Item"
		product.Quantity = 250
		product.UnitPrice = 9.99

		err = repo.Update(context.Background(), product)
		require.NoError(t, err)

		// Verify update
		updated, err := repo.GetByID(context.Background(), tenantID, product.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Item", updated.Name)
		assert.Equal(t, 250, updated.Quantity)
		assert.Equal(t, 9.99, updated.UnitPrice)
	})

	t.Run("Delete", func(t *testing.T) {
		product := &models.Product{
			ID:             uuid.New(),
			TenantID:       tenantID,
			Name:           "Delete Test Item",
			Quantity:       30,
			UnitPrice:      5.49,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)

		// Delete product
		err = repo.Delete(context.Background(), tenantID, product.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(context.Background(), tenantID, product.ID)
		assert.Error(t, err)
	})

	t.Run("AdvancedSearch", func(t *testing.T) {
		// Create test products
		product1 := &models.Product{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Organic Pesticide",
			Quantity:  100,
			UnitPrice: 30.00,
		}
		err := repo.Create(context.Background(), product1)
		require.NoError(t, err)

		product2 := &models.Product{
			ID:             uuid.New(),
			TenantID:       tenantID,
			Name:           "Fertilizer Mix",
			Quantity:       50,
			UnitPrice:      20.00,
			Barcode:        stringPtr("111111111"),
		}
		err = repo.Create(context.Background(), product2)
		require.NoError(t, err)

		// Test search by name
		filter := &models.ProductSearchFilter{
			Query: "Pesticide",
			Limit: 10,
		}
		results, err := repo.AdvancedSearch(context.Background(), tenantID, filter)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Organic Pesticide", results[0].Name)

		// Test search by barcode
		filter = &models.ProductSearchFilter{
			Barcode: stringPtr("111111111"),
			Limit:   10,
		}
		results, err = repo.AdvancedSearch(context.Background(), tenantID, filter)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Fertilizer Mix", results[0].Name)

		// Test price range
		filter = &models.ProductSearchFilter{
			MinPrice: floatPtr(25.00),
			MaxPrice: floatPtr(35.00),
			Limit:    10,
		}
		results, err = repo.AdvancedSearch(context.Background(), tenantID, filter)
		require.NoError(t, err)
		assert.True(t, len(results) > 0)
		for _, p := range results {
			assert.True(t, p.UnitPrice >= 25.00 && p.UnitPrice <= 35.00)
		}

		// Test sorting
		filter = &models.ProductSearchFilter{
			SortBy:    "name",
			SortOrder: "asc",
			Limit:     10,
		}
		results, err = repo.AdvancedSearch(context.Background(), tenantID, filter)
		require.NoError(t, err)
		if len(results) > 1 {
			assert.True(t, results[0].Name <= results[1].Name)
		}
	})

	t.Run("ListWithCategory", func(t *testing.T) {
		// Create product with category
		product := &models.Product{
			ID:         uuid.New(),
			TenantID:   tenantID,
			CategoryID: &categoryID,
			Name:       "Categorized Product",
			Quantity:   75,
			UnitPrice:  14.99,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)

		// Test filtering by category
		products, err := repo.ListWithCategory(context.Background(), tenantID, &categoryID, 10, 0)
		require.NoError(t, err)
		assert.True(t, len(products) > 0)
		for _, p := range products {
			assert.Equal(t, categoryID, *p.CategoryID)
		}

		// Test listing all with category info
		allProducts, err := repo.ListWithCategory(context.Background(), tenantID, nil, 10, 0)
		require.NoError(t, err)
		assert.True(t, len(allProducts) >= len(products))
	})

	t.Run("Search", func(t *testing.T) {
		// Create product for search
		product := &models.Product{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Name:      "Searchable Item",
			Quantity:  120,
			UnitPrice: 7.50,
			Barcode:   stringPtr("555666777"),
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)

		// Test name search
		results, err := repo.Search(context.Background(), tenantID, "Searchable", nil, 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Searchable Item", results[0].Name)

		// Test barcode search
		results, err = repo.Search(context.Background(), tenantID, "555666777", nil, 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Searchable Item", results[0].Name)
	})

	t.Run("CategoryAnalytics", func(t *testing.T) {
		// Ensure products exist with categories
		product := &models.Product{
			ID:         uuid.New(),
			TenantID:   tenantID,
			CategoryID: &categoryID,
			Name:       "Analytics Product",
			Quantity:   60,
			UnitPrice:  11.99,
		}
		err := repo.Create(context.Background(), product)
		require.NoError(t, err)

		// Test analytics
		analytics, err := repo.CategoryAnalytics(context.Background(), tenantID)
		require.NoError(t, err)
		assert.True(t, len(analytics) > 0)
		// Should have at least "Uncategorized" and the test category
		assert.Contains(t, analytics, "Uncategorized")
	})
}

// Helper functions for pointers
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func floatPtr(f float64) *float64 {
	return &f
}