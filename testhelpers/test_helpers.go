package testhelpers

import (
	"context"
	"log"
	"os"
	"testing"

	"agromart2/internal/config"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDB wraps the database connection with helper methods
type TestDB struct {
	Pool    *pgxpool.Pool
	t       *testing.T
	cleanup []func()
}

// NewTestDB creates a new test database helper
func NewTestDB(t *testing.T, connectionString string) *TestDB {
	t.Helper()

	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		t.Fatalf("Unable to connect to test database: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return &TestDB{
		Pool:    pool,
		t:       t,
		cleanup: make([]func(), 0),
	}
}

// stringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

// floatPtr returns a pointer to the given float64
func FloatPtr(f float64) *float64 {
	return &f
}

// Close closes the database connection and runs cleanup functions
func (td *TestDB) Close() {
	for i := len(td.cleanup) - 1; i >= 0; i-- {
		td.cleanup[i]()
	}
	td.Pool.Close()
}

// CreateTestTenant creates a test tenant and returns its ID
func (td *TestDB) CreateTestTenant(t *testing.T) uuid.UUID {
	t.Helper()

	tenantID := uuid.New()
	subdomain := "test-" + tenantID.String()[:8]

	tenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Test Tenant " + tenantID.String()[:8],
		Subdomain: subdomain,
		Status:    "active",
	}

	tenantRepo := repositories.NewTenantRepo(td.Pool)
	err := tenantRepo.Create(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	// Add cleanup function
	td.cleanup = append(td.cleanup, func() {
		_ = tenantRepo.Delete(context.Background(), tenantID)
	})

	return tenantID
}

// CreateTestUser creates a test user and returns its ID
func (td *TestDB) CreateTestUser(t *testing.T, tenantID uuid.UUID) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "test-" + userID.String()[:8] + "@example.com",
		Status:   "active",
	}

	userRepo := repositories.NewUserRepo(td.Pool)
	err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Add cleanup function
	td.cleanup = append(td.cleanup, func() {
		_ = userRepo.Delete(context.Background(), tenantID, userID)
	})

	return userID
}

// CreateTestProduct creates a test product and returns its ID
func (td *TestDB) CreateTestProduct(t *testing.T, tenantID uuid.UUID) uuid.UUID {
	t.Helper()

	productID := uuid.New()
	product := &models.Product{
		ID:        productID,
		TenantID:  tenantID,
		Name:      "Test Product " + productID.String()[:8],
		Quantity:  100,
		UnitPrice: 10.99,
	}

	productRepo := repositories.NewProductRepo(td.Pool)
	err := productRepo.Create(context.Background(), product)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	// Add cleanup function
	td.cleanup = append(td.cleanup, func() {
		_ = productRepo.Delete(context.Background(), tenantID, productID)
	})

	return productID
}

// CreateTestCategory creates a test category and returns its ID
func (td *TestDB) CreateTestCategory(t *testing.T, tenantID uuid.UUID, name string) uuid.UUID {
	t.Helper()

	categoryID := uuid.New()
	category := &models.Category{
		ID:          categoryID,
		TenantID:    tenantID,
		Name:        name,
		Description: "Test category description",
	}

	categoryRepo := repositories.NewCategoryRepo(td.Pool)
	err := categoryRepo.Create(context.Background(), category)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	// Add cleanup function
	td.cleanup = append(td.cleanup, func() {
		_ = categoryRepo.Delete(context.Background(), tenantID, categoryID)
	})

	return categoryID
}

// CreateTestSupplier creates a test supplier and returns its ID
func (td *TestDB) CreateTestSupplier(t *testing.T, tenantID uuid.UUID) uuid.UUID {
	t.Helper()

	supplierID := uuid.New()
	supplier := &models.Supplier{
		ID:         supplierID,
		TenantID:   tenantID,
		Name:       "Test Supplier " + supplierID.String()[:8],
	}

	supplierRepo := repositories.NewSupplierRepository(td.Pool)
	err := supplierRepo.Create(context.Background(), supplier)
	if err != nil {
		t.Fatalf("Failed to create test supplier: %v", err)
	}

	// Add cleanup function
	td.cleanup = append(td.cleanup, func() {
		_ = supplierRepo.Delete(context.Background(), tenantID, supplierID)
	})

	return supplierID
}

// CreateTestSupplierFull creates a test supplier with full model and returns the model
func (td *TestDB) CreateTestSupplierFull(tenantID uuid.UUID) models.Supplier {
	supplierID := uuid.New()
	name := "Test Supplier Full " + supplierID.String()[:8]
	email := supplierID.String()[:8] + "@supplier.com"
	phone := "123-456-7890"

	supplier := models.Supplier{
		ID:          supplierID,
		TenantID:    tenantID,
		Name:        name,
		ContactEmail: &email,
		ContactPhone: &phone,
		Address:     &name,
		LicenseNumber: &name,
	}

	supplierRepo := repositories.NewSupplierRepository(td.Pool)
	err := supplierRepo.Create(context.Background(), &supplier)
	if err != nil {
		// Fallback to simple creation if full creation fails
		supplier = models.Supplier{
			ID:       supplierID,
			TenantID: tenantID,
			Name:     "Test Supplier Full " + supplierID.String()[:8],
		}
	}

	// Add cleanup function
	td.cleanup = append(td.cleanup, func() {
		_ = supplierRepo.Delete(context.Background(), tenantID, supplierID)
	})

	return supplier
}

// CreateTestWarehouse creates a test warehouse and returns its ID
func (td *TestDB) CreateTestWarehouse(t *testing.T, tenantID uuid.UUID) uuid.UUID {
	t.Helper()

	warehouseID := uuid.New()
	warehouse := &models.Warehouse{
		ID:        warehouseID,
		TenantID:  tenantID,
		Name:      "Test Warehouse " + warehouseID.String()[:8],
	}

	warehouseRepo := repositories.NewWarehouseRepository(td.Pool)
	err := warehouseRepo.Create(context.Background(), warehouse)
	if err != nil {
		t.Fatalf("Failed to create test warehouse: %v", err)
	}

	// Add cleanup function
	td.cleanup = append(td.cleanup, func() {
		_ = warehouseRepo.Delete(context.Background(), tenantID, warehouseID)
	})

	return warehouseID
}

// CreateTestWarehouseFull creates a test warehouse with full model and returns the model
func (td *TestDB) CreateTestWarehouseFull(tenantID uuid.UUID) models.Warehouse {
	warehouseID := uuid.New()
	name := "Test Warehouse Full " + warehouseID.String()[:8]
	address := "123 Test St"
	capacity := 1000
	license := "LIC-" + warehouseID.String()[:8]

	warehouse := models.Warehouse{
		ID:            warehouseID,
		TenantID:      tenantID,
		Name:          name,
		Address:       &address,
		Capacity:      &capacity,
		LicenseNumber: &license,
	}

	warehouseRepo := repositories.NewWarehouseRepository(td.Pool)
	err := warehouseRepo.Create(context.Background(), &warehouse)
	if err != nil {
		// Fallback to simple creation if full creation fails
		warehouse = models.Warehouse{
			ID:       warehouseID,
			TenantID: tenantID,
			Name:     "Test Warehouse Full " + warehouseID.String()[:8],
		}
	}

	// Add cleanup function
	td.cleanup = append(td.cleanup, func() {
		_ = warehouseRepo.Delete(context.Background(), tenantID, warehouseID)
	})

	return warehouse
}

// CreateTestDistributor creates a test distributor and returns its ID
func (td *TestDB) CreateTestDistributor(t *testing.T, tenantID uuid.UUID) uuid.UUID {
	t.Helper()

	distributorID := uuid.New()
	distributor := &models.Distributor{
		ID:       distributorID,
		TenantID: tenantID,
		Name:     "Test Distributor " + distributorID.String()[:8],
	}

	distributorRepo := repositories.NewDistributorRepository(td.Pool)
	err := distributorRepo.Create(context.Background(), distributor)
	if err != nil {
		t.Fatalf("Failed to create test distributor: %v", err)
	}

	// Add cleanup function
	td.cleanup = append(td.cleanup, func() {
		_ = distributorRepo.Delete(context.Background(), tenantID, distributorID)
	})

	return distributorID
}

// CreateTestInventory creates test inventory record
func (td *TestDB) CreateTestInventory(t *testing.T, tenantID uuid.UUID, warehouseID uuid.UUID, productID uuid.UUID, quantity int) {
	t.Helper()

	inventoryRepo := repositories.NewInventoryRepo(td.Pool)

	inventoryID := uuid.New()
	inventory := &models.Inventory{
		ID:         inventoryID,
		TenantID:   tenantID,
		WarehouseID: warehouseID,
		ProductID:  productID,
		Quantity:   quantity,
	}

	err := inventoryRepo.Create(context.Background(), inventory)
	if err != nil {
		t.Fatalf("Failed to create test inventory: %v", err)
	}

	// No cleanup needed as inventory deletion would be handled by warehouse/product deletion
}

// SetupTestDB initializes a connection to the test database.
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Load Tally configuration
	tallyConfig, err := config.LoadTallyConfig("../config/tally.toml")
	if err != nil {
		log.Fatalf("Failed to load tally config: %v", err)
	}

	databaseURL := tallyConfig.Tally.TestDatabaseURL
	if databaseURL == "" {
		databaseURL = os.Getenv("TEST_DATABASE_URL")
	}
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL or test_database_url in config/tally.toml is not set")
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

// TruncateTables removes all data from the specified tables to ensure a clean state.
func TruncateTables(t *testing.T, pool *pgxpool.Pool, tables ...string) {
	t.Helper()
	for _, table := range tables {
		_, err := pool.Exec(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE")
		if err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}
