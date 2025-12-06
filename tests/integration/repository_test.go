package integration

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/testhelpers/containers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RepositoryTestSuite tests repositories with real PostgreSQL using testcontainers
type RepositoryTestSuite struct {
	suite.Suite
	container     *containers.PostgresContainer
	ctx           context.Context
	cancel        context.CancelFunc
	tenantRepo    repositories.TenantRepository
	productRepo   repositories.ProductRepository
	categoryRepo  repositories.CategoryRepository
	inventoryRepo repositories.InventoryRepository
	warehouseRepo repositories.WarehouseRepository
	tenantID      uuid.UUID
}

func TestRepositoryTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(RepositoryTestSuite))
}

func (s *RepositoryTestSuite) SetupSuite() {
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

	// Initialize repositories
	s.tenantRepo = repositories.NewTenantRepo(container.Pool)
	s.productRepo = repositories.NewProductRepo(container.Pool)
	s.categoryRepo = repositories.NewCategoryRepo(container.Pool)
	s.inventoryRepo = repositories.NewInventoryRepo(container.Pool)
	s.warehouseRepo = repositories.NewWarehouseRepository(container.Pool)

	// Create a test tenant
	s.tenantID = uuid.New()
}

func (s *RepositoryTestSuite) TearDownSuite() {
	if s.container != nil {
		s.container.Close(s.ctx)
	}
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RepositoryTestSuite) SetupTest() {
	// Clean tables before each test
	err := s.container.CleanTables(s.ctx)
	require.NoError(s.T(), err, "Failed to clean tables")

	// Create the test tenant
	_, err = s.container.Pool.Exec(s.ctx, `
		INSERT INTO tenants (id, name, subdomain, status, created_at)
		VALUES ($1, 'Test Tenant', $2, 'active', NOW())
	`, s.tenantID, "test-tenant-"+s.tenantID.String()[:8])
	require.NoError(s.T(), err)
}

// =====================
// Tenant Repository Tests
// =====================

func (s *RepositoryTestSuite) TestTenantCreateAndFetch() {
	tenantID := uuid.New()
	subdomain := "tenant-" + tenantID.String()[:8]

	tenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Tenant Repo",
		Subdomain: subdomain,
		License:   "LIC-123",
		Status:    "active",
	}

	err := s.tenantRepo.Create(s.ctx, tenant)
	require.NoError(s.T(), err)

	fetchedByID, err := s.tenantRepo.GetByID(s.ctx, tenantID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), tenant.Name, fetchedByID.Name)
	assert.Equal(s.T(), tenant.Subdomain, fetchedByID.Subdomain)
	assert.False(s.T(), fetchedByID.CreatedAt.IsZero())

	fetchedBySubdomain, err := s.tenantRepo.GetBySubdomain(s.ctx, subdomain)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), tenantID, fetchedBySubdomain.ID)
	assert.Equal(s.T(), subdomain, fetchedBySubdomain.Subdomain)
}

func (s *RepositoryTestSuite) TestTenantUpdateAndDelete() {
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Original Tenant",
		Subdomain: "tenant-" + tenantID.String()[:8],
		License:   "LIC-ORIGINAL",
		Status:    "active",
	}

	err := s.tenantRepo.Create(s.ctx, tenant)
	require.NoError(s.T(), err)

	tenant.Name = "Updated Tenant"
	tenant.Subdomain = "updated-" + tenantID.String()[:6]
	tenant.License = "LIC-UPDATED"
	tenant.Status = "inactive"

	err = s.tenantRepo.Update(s.ctx, tenant)
	require.NoError(s.T(), err)

	updated, err := s.tenantRepo.GetByID(s.ctx, tenantID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Updated Tenant", updated.Name)
	assert.Equal(s.T(), tenant.Subdomain, updated.Subdomain)
	assert.Equal(s.T(), "LIC-UPDATED", updated.License)
	assert.Equal(s.T(), "inactive", updated.Status)

	err = s.tenantRepo.Delete(s.ctx, tenantID)
	require.NoError(s.T(), err)

	_, err = s.tenantRepo.GetByID(s.ctx, tenantID)
	require.Error(s.T(), err)
}

func (s *RepositoryTestSuite) TestTenantSettingsRoundTrip() {
	tenantID := uuid.New()
	subdomain := "settings-" + tenantID.String()[:8]

	tenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Settings Tenant",
		Subdomain: subdomain,
		License:   "LIC-SET",
		Status:    "active",
	}

	err := s.tenantRepo.Create(s.ctx, tenant)
	require.NoError(s.T(), err)

	settings, err := s.tenantRepo.FindSettingsByTenantID(s.ctx, tenantID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), tenantID, settings.ID)

	settings.Name = "Settings Updated"
	settings.Subdomain = "settings-updated-" + tenantID.String()[:6]
	settings.License = "LIC-SET-NEW"

	err = s.tenantRepo.UpdateSettings(s.ctx, settings)
	require.NoError(s.T(), err)

	updated, err := s.tenantRepo.FindSettingsByTenantID(s.ctx, tenantID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Settings Updated", updated.Name)
	assert.Equal(s.T(), settings.Subdomain, updated.Subdomain)
	assert.Equal(s.T(), "LIC-SET-NEW", updated.License)
}

func (s *RepositoryTestSuite) TestTenantListPagination() {
	baseCount := 1 // SetupTest inserts a base tenant

	tenants := []*models.Tenant{
		{ID: uuid.New(), Name: "Tenant A", Subdomain: "tenant-a-" + uuid.NewString()[:6], Status: "active"},
		{ID: uuid.New(), Name: "Tenant B", Subdomain: "tenant-b-" + uuid.NewString()[:6], Status: "active"},
		{ID: uuid.New(), Name: "Tenant C", Subdomain: "tenant-c-" + uuid.NewString()[:6], Status: "active"},
	}

	for _, t := range tenants {
		err := s.tenantRepo.Create(s.ctx, t)
		require.NoError(s.T(), err)
	}

	firstPage, err := s.tenantRepo.List(s.ctx, 2, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), firstPage, 2)

	secondPage, err := s.tenantRepo.List(s.ctx, 2, 2)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(secondPage), 1)

	totalFetched := len(firstPage) + len(secondPage)
	assert.Equal(s.T(), baseCount+len(tenants), totalFetched)
}

// =====================
// Category Repository Tests
// =====================

func (s *RepositoryTestSuite) TestCategoryCreate() {
	category := &models.Category{
		ID:          uuid.New(),
		TenantID:    s.tenantID,
		Name:        "Test Category",
		Description: "Test Description",
	}

	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	// Verify it was created
	fetched, err := s.categoryRepo.GetByID(s.ctx, s.tenantID, category.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), category.Name, fetched.Name)
	assert.Equal(s.T(), category.Description, fetched.Description)
}

func (s *RepositoryTestSuite) TestCategoryList() {
	// Create multiple categories
	categories := []*models.Category{
		{ID: uuid.New(), TenantID: s.tenantID, Name: "Category A", Description: "Desc A"},
		{ID: uuid.New(), TenantID: s.tenantID, Name: "Category B", Description: "Desc B"},
		{ID: uuid.New(), TenantID: s.tenantID, Name: "Category C", Description: "Desc C"},
	}

	for _, cat := range categories {
		err := s.categoryRepo.Create(s.ctx, cat)
		require.NoError(s.T(), err)
	}

	// List categories
	list, err := s.categoryRepo.List(s.ctx, s.tenantID, 10, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), list, 3)
}

func (s *RepositoryTestSuite) TestCategoryUpdate() {
	category := &models.Category{
		ID:          uuid.New(),
		TenantID:    s.tenantID,
		Name:        "Original Name",
		Description: "Original Description",
	}

	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	// Update the category
	category.Name = "Updated Name"
	category.Description = "Updated Description"
	err = s.categoryRepo.Update(s.ctx, category)
	require.NoError(s.T(), err)

	// Verify update
	fetched, err := s.categoryRepo.GetByID(s.ctx, s.tenantID, category.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Updated Name", fetched.Name)
	assert.Equal(s.T(), "Updated Description", fetched.Description)
}

func (s *RepositoryTestSuite) TestCategoryDelete() {
	category := &models.Category{
		ID:          uuid.New(),
		TenantID:    s.tenantID,
		Name:        "To Be Deleted",
		Description: "Will be deleted",
	}

	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	// Delete
	err = s.categoryRepo.Delete(s.ctx, s.tenantID, category.ID)
	require.NoError(s.T(), err)

	// Verify deletion
	_, err = s.categoryRepo.GetByID(s.ctx, s.tenantID, category.ID)
	require.Error(s.T(), err)
}

// =====================
// Product Repository Tests
// =====================

func (s *RepositoryTestSuite) TestProductCreate() {
	// First create a category
	category := &models.Category{
		ID:          uuid.New(),
		TenantID:    s.tenantID,
		Name:        "Product Category",
		Description: "For products",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	barcode := "1234567890"
	unitMeasure := "kg"
	description := "Test product description"

	product := &models.Product{
		ID:            uuid.New(),
		TenantID:      s.tenantID,
		CategoryID:    &category.ID,
		Name:          "Test Product",
		Quantity:      100,
		UnitPrice:     10.99,
		Barcode:       &barcode,
		UnitOfMeasure: &unitMeasure,
		Description:   &description,
	}

	err = s.productRepo.Create(s.ctx, product)
	require.NoError(s.T(), err)

	// Verify
	fetched, err := s.productRepo.GetByID(s.ctx, s.tenantID, product.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Test Product", fetched.Name)
	assert.Equal(s.T(), 100, fetched.Quantity)
	assert.Equal(s.T(), 10.99, fetched.UnitPrice)
}

func (s *RepositoryTestSuite) TestProductGetByBarcode() {
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Barcode Category",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	barcode := "BARCODE123"
	product := &models.Product{
		ID:         uuid.New(),
		TenantID:   s.tenantID,
		CategoryID: &category.ID,
		Name:       "Barcode Product",
		Quantity:   50,
		UnitPrice:  5.00,
		Barcode:    &barcode,
	}

	err = s.productRepo.Create(s.ctx, product)
	require.NoError(s.T(), err)

	// Find by barcode
	fetched, err := s.productRepo.GetByBarcode(s.ctx, s.tenantID, barcode)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), product.ID, fetched.ID)
	assert.Equal(s.T(), "Barcode Product", fetched.Name)
}

func (s *RepositoryTestSuite) TestProductList() {
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "List Category",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	// Create products
	for i := 0; i < 5; i++ {
		product := &models.Product{
			ID:         uuid.New(),
			TenantID:   s.tenantID,
			CategoryID: &category.ID,
			Name:       "Product " + string(rune('A'+i)),
			Quantity:   i * 10,
			UnitPrice:  float64(i) * 1.00,
		}
		err := s.productRepo.Create(s.ctx, product)
		require.NoError(s.T(), err)
	}

	// List with pagination
	list, err := s.productRepo.List(s.ctx, s.tenantID, 3, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), list, 3)

	list2, err := s.productRepo.List(s.ctx, s.tenantID, 3, 3)
	require.NoError(s.T(), err)
	assert.Len(s.T(), list2, 2)
}

func (s *RepositoryTestSuite) TestProductUpdateQuantity() {
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Quantity Category",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	product := &models.Product{
		ID:         uuid.New(),
		TenantID:   s.tenantID,
		CategoryID: &category.ID,
		Name:       "Quantity Product",
		Quantity:   100,
		UnitPrice:  10.00,
	}
	err = s.productRepo.Create(s.ctx, product)
	require.NoError(s.T(), err)

	// Update quantity
	err = s.productRepo.UpdateQuantity(s.ctx, s.tenantID, product.ID, 150)
	require.NoError(s.T(), err)

	// Verify
	fetched, err := s.productRepo.GetByID(s.ctx, s.tenantID, product.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 150, fetched.Quantity)
}

func (s *RepositoryTestSuite) TestProductAdvancedSearch() {
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Advanced Category",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	hazardClass := "flammable"
	sdsURL := "http://example.com/sds"
	activeIngredients := "ABC-123"
	description := "Hazard gear description"

	targetProduct := &models.Product{
		ID:                uuid.New(),
		TenantID:          s.tenantID,
		CategoryID:        &category.ID,
		Name:              "Hazard Gear",
		Quantity:          10,
		UnitPrice:         50.0,
		Description:       &description,
		IsHazardous:       true,
		HazardClass:       &hazardClass,
		SDSUrl:            &sdsURL,
		ActiveIngredients: &activeIngredients,
	}

	otherProduct := &models.Product{
		ID:         uuid.New(),
		TenantID:   s.tenantID,
		CategoryID: &category.ID,
		Name:       "Regular Item",
		Quantity:   5,
		UnitPrice:  20.0,
	}

	require.NoError(s.T(), s.productRepo.Create(s.ctx, targetProduct))
	require.NoError(s.T(), s.productRepo.Create(s.ctx, otherProduct))

	filter := &models.ProductSearchFilter{
		Query: "hazard",
		Limit: 10,
	}

	results, err := s.productRepo.AdvancedSearch(s.ctx, s.tenantID, filter)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), results)

	var found *models.Product
	for _, p := range results {
		if p.ID == targetProduct.ID {
			found = p
			break
		}
	}

	require.NotNil(s.T(), found, "expected to find the hazardous product in results")
	assert.Equal(s.T(), targetProduct.IsHazardous, found.IsHazardous)
	assert.Equal(s.T(), hazardClass, *found.HazardClass)
	assert.Equal(s.T(), sdsURL, *found.SDSUrl)
	assert.Equal(s.T(), activeIngredients, *found.ActiveIngredients)
}

func (s *RepositoryTestSuite) TestProductBulkUpdatePrices() {
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Price Category",
	}
	require.NoError(s.T(), s.categoryRepo.Create(s.ctx, category))

	productA := &models.Product{
		ID:         uuid.New(),
		TenantID:   s.tenantID,
		CategoryID: &category.ID,
		Name:       "Price A",
		Quantity:   10,
		UnitPrice:  100.0,
	}
	productB := &models.Product{
		ID:         uuid.New(),
		TenantID:   s.tenantID,
		CategoryID: &category.ID,
		Name:       "Price B",
		Quantity:   20,
		UnitPrice:  200.0,
	}

	require.NoError(s.T(), s.productRepo.Create(s.ctx, productA))
	require.NoError(s.T(), s.productRepo.Create(s.ctx, productB))

	rows, err := s.productRepo.BulkUpdatePrices(s.ctx, s.tenantID, []uuid.UUID{productA.ID, productB.ID}, "percentage", 10.0)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2), rows)

	updatedA, err := s.productRepo.GetByID(s.ctx, s.tenantID, productA.ID)
	require.NoError(s.T(), err)
	updatedB, err := s.productRepo.GetByID(s.ctx, s.tenantID, productB.ID)
	require.NoError(s.T(), err)

	assert.InDelta(s.T(), 110.0, updatedA.UnitPrice, 0.001)
	assert.InDelta(s.T(), 220.0, updatedB.UnitPrice, 0.001)
}

// =====================
// Warehouse Repository Tests
// =====================

func (s *RepositoryTestSuite) TestWarehouseCreate() {
	warehouse := &models.Warehouse{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Test Warehouse",
		Address:  stringPtr("123 Test Street"),
		Capacity: intPtr(1000),
	}

	err := s.warehouseRepo.Create(s.ctx, warehouse)
	require.NoError(s.T(), err)

	// Verify
	fetched, err := s.warehouseRepo.GetByID(s.ctx, s.tenantID, warehouse.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Test Warehouse", fetched.Name)
	assert.Equal(s.T(), stringPtr("123 Test Street"), fetched.Address)
	assert.Equal(s.T(), intPtr(1000), fetched.Capacity)
}

// =====================
// Inventory Repository Tests
// =====================

func (s *RepositoryTestSuite) TestInventoryCreate() {
	// Create category, product, and warehouse first
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Inventory Category",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	product := &models.Product{
		ID:         uuid.New(),
		TenantID:   s.tenantID,
		CategoryID: &category.ID,
		Name:       "Inventory Product",
		Quantity:   0,
		UnitPrice:  10.00,
	}
	err = s.productRepo.Create(s.ctx, product)
	require.NoError(s.T(), err)

	warehouse := &models.Warehouse{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Inventory Warehouse",
		Capacity: intPtr(1000),
	}
	err = s.warehouseRepo.Create(s.ctx, warehouse)
	require.NoError(s.T(), err)

	// Create inventory
	inventory := &models.Inventory{
		ID:          uuid.New(),
		TenantID:    s.tenantID,
		WarehouseID: warehouse.ID,
		ProductID:   product.ID,
		Quantity:    50,
	}

	err = s.inventoryRepo.Create(s.ctx, inventory)
	require.NoError(s.T(), err)

	// Verify
	fetched, err := s.inventoryRepo.GetByWarehouseAndProduct(s.ctx, s.tenantID, warehouse.ID, product.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 50, fetched.Quantity)
}

func (s *RepositoryTestSuite) TestInventoryUpsert() {
	// Create dependencies
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Upsert Category",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	product := &models.Product{
		ID:         uuid.New(),
		TenantID:   s.tenantID,
		CategoryID: &category.ID,
		Name:       "Upsert Product",
		Quantity:   0,
		UnitPrice:  10.00,
	}
	err = s.productRepo.Create(s.ctx, product)
	require.NoError(s.T(), err)

	warehouse := &models.Warehouse{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Upsert Warehouse",
		Capacity: intPtr(1000),
	}
	err = s.warehouseRepo.Create(s.ctx, warehouse)
	require.NoError(s.T(), err)

	// First insert
	inventory := &models.Inventory{
		ID:          uuid.New(),
		TenantID:    s.tenantID,
		WarehouseID: warehouse.ID,
		ProductID:   product.ID,
		Quantity:    50,
	}
	err = s.inventoryRepo.Create(s.ctx, inventory)
	require.NoError(s.T(), err)

	// Second insert (should upsert and add quantity)
	inventory2 := &models.Inventory{
		ID:          uuid.New(), // Different ID
		TenantID:    s.tenantID,
		WarehouseID: warehouse.ID,
		ProductID:   product.ID,
		Quantity:    30,
	}
	err = s.inventoryRepo.Create(s.ctx, inventory2)
	require.NoError(s.T(), err)

	// Verify the quantity was added (50 + 30 = 80)
	fetched, err := s.inventoryRepo.GetByWarehouseAndProduct(s.ctx, s.tenantID, warehouse.ID, product.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 80, fetched.Quantity)
}

func (s *RepositoryTestSuite) TestInventoryTransfer() {
	// Create dependencies
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Transfer Category",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	product := &models.Product{
		ID:         uuid.New(),
		TenantID:   s.tenantID,
		CategoryID: &category.ID,
		Name:       "Transfer Product",
		Quantity:   0,
		UnitPrice:  10.00,
	}
	err = s.productRepo.Create(s.ctx, product)
	require.NoError(s.T(), err)

	warehouse1 := &models.Warehouse{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Source Warehouse",
		Capacity: intPtr(1000),
	}
	err = s.warehouseRepo.Create(s.ctx, warehouse1)
	require.NoError(s.T(), err)

	warehouse2 := &models.Warehouse{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Destination Warehouse",
		Capacity: intPtr(1000),
	}
	err = s.warehouseRepo.Create(s.ctx, warehouse2)
	require.NoError(s.T(), err)

	// Create initial inventory in source warehouse
	inventory := &models.Inventory{
		ID:          uuid.New(),
		TenantID:    s.tenantID,
		WarehouseID: warehouse1.ID,
		ProductID:   product.ID,
		Quantity:    100,
	}
	err = s.inventoryRepo.Create(s.ctx, inventory)
	require.NoError(s.T(), err)

	// Transfer 40 units from warehouse1 to warehouse2
	err = s.inventoryRepo.Transfer(s.ctx, s.tenantID, product.ID, warehouse1.ID, warehouse2.ID, 40)
	require.NoError(s.T(), err)

	// Verify source has 60
	source, err := s.inventoryRepo.GetByWarehouseAndProduct(s.ctx, s.tenantID, warehouse1.ID, product.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 60, source.Quantity)

	// Verify destination has 40
	dest, err := s.inventoryRepo.GetByWarehouseAndProduct(s.ctx, s.tenantID, warehouse2.ID, product.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 40, dest.Quantity)
}

func (s *RepositoryTestSuite) TestInventoryTransferInsufficientQuantity() {
	// Create dependencies
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Insufficient Category",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	product := &models.Product{
		ID:         uuid.New(),
		TenantID:   s.tenantID,
		CategoryID: &category.ID,
		Name:       "Insufficient Product",
		Quantity:   0,
		UnitPrice:  10.00,
	}
	err = s.productRepo.Create(s.ctx, product)
	require.NoError(s.T(), err)

	warehouse1 := &models.Warehouse{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Source Insufficient",
		Capacity: intPtr(1000),
	}
	err = s.warehouseRepo.Create(s.ctx, warehouse1)
	require.NoError(s.T(), err)

	warehouse2 := &models.Warehouse{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Destination Insufficient",
		Capacity: intPtr(1000),
	}
	err = s.warehouseRepo.Create(s.ctx, warehouse2)
	require.NoError(s.T(), err)

	// Create initial inventory with only 20 units
	inventory := &models.Inventory{
		ID:          uuid.New(),
		TenantID:    s.tenantID,
		WarehouseID: warehouse1.ID,
		ProductID:   product.ID,
		Quantity:    20,
	}
	err = s.inventoryRepo.Create(s.ctx, inventory)
	require.NoError(s.T(), err)

	// Try to transfer 50 units (more than available)
	err = s.inventoryRepo.Transfer(s.ctx, s.tenantID, product.ID, warehouse1.ID, warehouse2.ID, 50)
	require.Error(s.T(), err, "Should fail due to insufficient quantity")
}

func (s *RepositoryTestSuite) TestInventoryList() {
	// Create dependencies
	category := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "List Inventory Category",
	}
	err := s.categoryRepo.Create(s.ctx, category)
	require.NoError(s.T(), err)

	warehouse := &models.Warehouse{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "List Inventory Warehouse",
		Capacity: intPtr(5000),
	}
	err = s.warehouseRepo.Create(s.ctx, warehouse)
	require.NoError(s.T(), err)

	// Create multiple products and inventory entries
	for i := 0; i < 5; i++ {
		product := &models.Product{
			ID:         uuid.New(),
			TenantID:   s.tenantID,
			CategoryID: &category.ID,
			Name:       "List Product " + string(rune('A'+i)),
			Quantity:   0,
			UnitPrice:  float64(i+1) * 5.00,
		}
		err := s.productRepo.Create(s.ctx, product)
		require.NoError(s.T(), err)

		inventory := &models.Inventory{
			ID:          uuid.New(),
			TenantID:    s.tenantID,
			WarehouseID: warehouse.ID,
			ProductID:   product.ID,
			Quantity:    (i + 1) * 20,
		}
		err = s.inventoryRepo.Create(s.ctx, inventory)
		require.NoError(s.T(), err)
	}

	// List inventory
	list, err := s.inventoryRepo.List(s.ctx, s.tenantID, 10, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), list, 5)
}

// =====================
// Multi-Tenant Isolation Tests
// =====================

func (s *RepositoryTestSuite) TestMultiTenantIsolation() {
	// Create a second tenant
	tenant2ID := uuid.New()
	_, err := s.container.Pool.Exec(s.ctx, `
		INSERT INTO tenants (id, name, subdomain, status, created_at)
		VALUES ($1, 'Tenant 2', $2, 'active', NOW())
	`, tenant2ID, "tenant2-"+tenant2ID.String()[:8])
	require.NoError(s.T(), err)

	// Create categories for both tenants
	cat1 := &models.Category{
		ID:       uuid.New(),
		TenantID: s.tenantID,
		Name:     "Tenant 1 Category",
	}
	err = s.categoryRepo.Create(s.ctx, cat1)
	require.NoError(s.T(), err)

	cat2 := &models.Category{
		ID:       uuid.New(),
		TenantID: tenant2ID,
		Name:     "Tenant 2 Category",
	}
	err = s.categoryRepo.Create(s.ctx, cat2)
	require.NoError(s.T(), err)

	// List categories for tenant 1 - should only see tenant 1's category
	list1, err := s.categoryRepo.List(s.ctx, s.tenantID, 10, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), list1, 1)
	assert.Equal(s.T(), "Tenant 1 Category", list1[0].Name)

	// List categories for tenant 2 - should only see tenant 2's category
	list2, err := s.categoryRepo.List(s.ctx, tenant2ID, 10, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), list2, 1)
	assert.Equal(s.T(), "Tenant 2 Category", list2[0].Name)

	// Tenant 1 should not be able to get tenant 2's category by ID
	_, err = s.categoryRepo.GetByID(s.ctx, s.tenantID, cat2.ID)
	require.Error(s.T(), err, "Should not be able to access other tenant's data")
}
