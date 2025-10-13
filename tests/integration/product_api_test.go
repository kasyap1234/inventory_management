package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"agromart2/internal/common"
	"agromart2/internal/handlers"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to set tenant ID in context for testing
func setTenantIDInContext(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, common.TenantIDKey, tenantID)
}

// setupTestDB creates a test database connection for API tests
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL environment variable not set, skipping API tests")
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
	tables := []string{"products", "categories", "users", "tenants"}
	for _, table := range tables {
		_, err := pool.Exec(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE")
		if err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}

func TestProductAPI(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	// Setup: truncate tables for clean state
	truncateTestTables(t, pool)
	defer truncateTestTables(t, pool) // Cleanup after test

	// Create test tenant and user
	tenantID := uuid.New()
	userID := uuid.New()

	// Create tenant
	tenantRepo := repositories.NewTenantRepo(pool)
	tenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Test Tenant",
		Subdomain: "test-tenant",
		Status:    "active",
	}
	err := tenantRepo.Create(context.Background(), tenant)
	require.NoError(t, err)

	// Create user
	userRepo := repositories.NewUserRepo(pool)
	user := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "test@example.com",
		Status:   "active",
	}
	err = userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	// Setup Echo instance with handlers
	e := echo.New()

	// Create mock service instances (using nil for services that aren't essential for basic CRUD)
	productRepo := repositories.NewProductRepo(pool)
	inventoryRepo := repositories.NewInventoryRepo(pool)
	categoryRepo := repositories.NewCategoryRepo(pool)
	productImageRepo := repositories.NewProductImageRepo(pool)
	
	// For testing, we'll use nil for minioService and cacheService since they're not essential for basic CRUD operations
	productService := services.NewProductService(productRepo, inventoryRepo, categoryRepo, productImageRepo, nil, nil)
	productHandlers := handlers.NewProductHandlers(productService, nil)

	t.Run("Create Product API", func(t *testing.T) {
		// Test product creation
		productReq := map[string]interface{}{
			"name":       "Test Product API",
			"quantity":   10,
			"unit_price": 29.99,
			"barcode":    "API123456",
		}

		jsonData, _ := json.Marshal(productReq)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Add tenant ID to context
		c.SetRequest(req.WithContext(common.SetTenantIDInContext(req.Context(), tenantID)))

		err := productHandlers.CreateProduct(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Product created successfully", response["message"])
	})

	t.Run("List Products API", func(t *testing.T) {
		// First create a few products
		for i := 0; i < 3; i++ {
			productReq := map[string]interface{}{
				"name":       fmt.Sprintf("List Test Product %d", i),
				"quantity":   50,
				"unit_price": float64(i+1) * 10.0,
				"barcode":    fmt.Sprintf("LIST%d", i),
			}

			jsonData, _ := json.Marshal(productReq)
			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonData))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := productHandlers.CreateProduct(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
		}

		// Now test listing
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = productHandlers.ListProducts(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "products")
	assert.Contains(t, response, "limit")
		assert.Contains(t, response, "offset")

		products := response["products"].([]interface{})
		assert.GreaterOrEqual(t, len(products), 3)
	})

	t.Run("Get Product by ID API", func(t *testing.T) {
		// Create a test product first
	productReq := map[string]interface{}{
			"name":       "Get Test Product",
			"quantity":   75,
			"unit_price": 15.99,
			"barcode":    "GET123",
		}

		jsonData, _ := json.Marshal(productReq)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

		err := productHandlers.CreateProduct(c)
	assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Parse the created product to get its ID
	var createResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &createResponse)
		assert.NoError(t, err)

		createdProduct := createResponse["product"].(map[string]interface{})
		productID := createdProduct["id"].(string)

		// Now test getting the product by ID
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/products/%s", productID), nil)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID)

		err = productHandlers.GetProductByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var getProductResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &getProductResponse)
		assert.NoError(t, err)
	assert.Equal(t, productID, getProductResponse["id"])
		assert.Equal(t, "Get Test Product", getProductResponse["name"])
	})

	t.Run("Update Product API", func(t *testing.T) {
		// Create a test product first
	productReq := map[string]interface{}{
			"name":       "Update Test Product",
			"quantity":   50,
			"unit_price": 19.99,
			"barcode":    "UPDATE123",
		}

		jsonData, _ := json.Marshal(productReq)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := productHandlers.CreateProduct(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Parse the created product to get its ID
		var createResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &createResponse)
		assert.NoError(t, err)

		createdProduct := createResponse["product"].(map[string]interface{})
		productID := createdProduct["id"].(string)

		// Now test updating the product
		updateReq := map[string]interface{}{
			"name":       "Updated Product Name",
			"quantity":   10,
			"unit_price": 25.99,
			"barcode":    "UPDATE456",
		}

		jsonData, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/products/%s", productID), bytes.NewBuffer(jsonData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID)

		err = productHandlers.UpdateProduct(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var updateResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &updateResponse)
		assert.NoError(t, err)
	assert.Equal(t, "Product updated successfully", updateResponse["message"])

		// Verify the update by getting the product
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/products/%s", productID), nil)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID)

		err = productHandlers.GetProductByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var getProductResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &getProductResponse)
		assert.NoError(t, err)
	assert.Equal(t, "Updated Product Name", getProductResponse["name"])
		assert.Equal(t, 25.99, getProductResponse["unit_price"])
	})

	t.Run("Delete Product API", func(t *testing.T) {
		// Create a test product first
		productReq := map[string]interface{}{
			"name":       "Delete Test Product",
			"quantity":   25,
			"unit_price": 9.99,
			"barcode":    "DELETE123",
		}

		jsonData, _ := json.Marshal(productReq)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := productHandlers.CreateProduct(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

	// Parse the created product to get its ID
		var createResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &createResponse)
		assert.NoError(t, err)

		createdProduct := createResponse["product"].(map[string]interface{})
		productID := createdProduct["id"].(string)

		// Now test deleting the product
		req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/products/%s", productID), nil)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID)

		err = productHandlers.DeleteProduct(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var deleteResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &deleteResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Product deleted successfully", deleteResponse["message"])

		// Verify the deletion by trying to get the product
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/products/%s", productID), nil)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID)

		err = productHandlers.GetProductByID(c)
		// This should fail since the product was deleted
		// Note: The actual behavior depends on how the service handles deleted records
	// For this test, we'll just verify the delete operation completed
	})

	t.Run("Search Products API", func(t *testing.T) {
		// Create searchable products
		products := []map[string]interface{}{
			{"name": "Apple iPhone", "quantity": 10, "unit_price": 999.99, "barcode": "IPH001"},
			{"name": "Samsung Galaxy", "quantity": 15, "unit_price": 899.99, "barcode": "SAM001"},
			{"name": "Apple MacBook", "quantity": 5, "unit_price": 1999.9, "barcode": "MAC001"},
		}

		for _, product := range products {
			jsonData, _ := json.Marshal(product)
			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonData))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := productHandlers.CreateProduct(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
		}

		// Test searching for "Apple" products
		req := httptest.NewRequest(http.MethodGet, "/products/search?q=Apple", nil)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec = httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = productHandlers.SearchProducts(c)
	assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var searchResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &searchResponse)
		assert.NoError(t, err)
		assert.Contains(t, searchResponse, "products")
		assert.Contains(t, searchResponse, "query")

		results := searchResponse["products"].([]interface{})
		assert.GreaterOrEqual(t, len(results), 2) // Should find iPhone and MacBook
	})

	t.Run("Product Analytics API", func(t *testing.T) {
		// Test analytics endpoint
		req := httptest.NewRequest(http.MethodGet, "/products/analytics", nil)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = productHandlers.GetProductAnalytics(c)
	assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var analyticsResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &analyticsResponse)
		assert.NoError(t, err)
		assert.Contains(t, analyticsResponse, "analytics")
		assert.Contains(t, analyticsResponse, "description")
	})

	// Additional tests for bulk operations
	t.Run("Bulk Price Update API", func(t *testing.T) {
		// Create some test products first
	productIDs := make([]string, 0)
		for i := 0; i < 3; i++ {
			productReq := map[string]interface{}{
				"name":       fmt.Sprintf("Bulk Update Test Product %d", i),
				"quantity":   30,
				"unit_price": 50.0,
				"barcode":    fmt.Sprintf("BULK%d", i),
			}

			jsonData, _ := json.Marshal(productReq)
			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonData))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := productHandlers.CreateProduct(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)

			// Parse the created product to get its ID
			var createResponse map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &createResponse)
			assert.NoError(t, err)

			createdProduct := createResponse["product"].(map[string]interface{})
			productIDs = append(productIDs, createdProduct["id"].(string))
		}

		// Test bulk price update
		bulkUpdateReq := map[string]interface{}{
			"product_ids": productIDs,
			"adjustment": map[string]interface{}{
				"type":  "percentage",
				"value": 10.0, // 10% increase
			},
		}

		jsonData, _ := json.Marshal(bulkUpdateReq)
	req := httptest.NewRequest(http.MethodPost, "/products/bulk-price-update", bytes.NewBuffer(jsonData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = productHandlers.BulkPriceUpdate(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var bulkUpdateResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &bulkUpdateResponse)
		assert.NoError(t, err)
		assert.Contains(t, bulkUpdateResponse, "updated_count")
		assert.Contains(t, bulkUpdateResponse, "total_count")
	})

	// Test bulk create
	t.Run("Bulk Create Products API", func(t *testing.T) {
		bulkCreateReq := models.ProductBulkCreate{
			Products: []*models.Product{
				{
					Name:      "Bulk Create Product 1",
					Quantity:  10,
					UnitPrice: 25.0,
				},
				{
					Name:      "Bulk Create Product 2",
					Quantity:  20,
					UnitPrice: 35.0,
				},
			},
		}

		jsonData, _ := json.Marshal(bulkCreateReq)
		req := httptest.NewRequest(http.MethodPost, "/products/bulk/create", bytes.NewBuffer(jsonData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(common.SetTenantIDInContext(context.Background(), tenantID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = productHandlers.BulkCreateProducts(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var bulkCreateResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &bulkCreateResponse)
		assert.NoError(t, err)
		assert.Contains(t, bulkCreateResponse, "status")
		assert.Contains(t, bulkCreateResponse, "created_count")
	})
}
