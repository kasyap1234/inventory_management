//go:build legacy_tests
package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Mock repositories
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, tenantID, productID uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, tenantID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*models.Product, error) {
	args := m.Called(ctx, tenantID, barcode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, tenantID, productID uuid.UUID) error {
	args := m.Called(ctx, tenantID, productID)
	return args.Error(0)
}

func (m *MockProductRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, tenantID, query, categoryID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) ListWithCategory(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]*models.ProductWithCategory, error) {
	args := m.Called(ctx, tenantID, categoryID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProductWithCategory), args.Error(1)
}

func (m *MockProductRepository) CategoryAnalytics(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]int, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]int), args.Error(1)
}

func (m *MockProductRepository) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.ProductSearchFilter) ([]*models.Product, error) {
	args := m.Called(ctx, tenantID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) GetByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.Inventory, error) {
	args := m.Called(ctx, tenantID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Inventory), args.Error(1)
}

type ProductServiceTestSuite struct {
	suite.Suite
	mockProductRepo   *MockProductRepository
	mockInventoryRepo *MockInventoryRepository
	mockCategoryRepo  *MockCategoryRepository
	mockImageRepo     *MockProductImageRepository
	mockMinioService  *MockMinioService
	mockCacheService  *testhelpers.MockCacheService
	productService    ProductService
	tenantID          uuid.UUID
	ctx               context.Context
}

// Mock repositories that might be needed
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *models.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Category, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *models.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Category, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Category), args.Error(1)
}

type MockProductImageRepository struct {
	mock.Mock
}

func (m *MockProductImageRepository) Create(ctx context.Context, image *models.ProductImage) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *MockProductImageRepository) GetByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.ProductImage, error) {
	args := m.Called(ctx, tenantID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProductImage), args.Error(1)
}

func (m *MockProductImageRepository) Delete(ctx context.Context, tenantID, imageID uuid.UUID) error {
	args := m.Called(ctx, tenantID, imageID)
	return args.Error(0)
}

type MockMinioService struct {
	mock.Mock
}

func (m *MockMinioService) UploadImage(ctx context.Context, bucket string, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	args := m.Called(ctx, bucket, objectKey, reader, size, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockMinioService) GetImageURL(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, bucket, objectKey, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockMinioService) DeleteImage(ctx context.Context, bucket string, objectKey string) error {
	args := m.Called(ctx, bucket, objectKey)
	return args.Error(0)
}

func (suite *ProductServiceTestSuite) SetupTest() {
	suite.mockProductRepo = &MockProductRepository{}
	suite.mockInventoryRepo = &MockInventoryRepository{}
	suite.mockCategoryRepo = &MockCategoryRepository{}
	suite.mockImageRepo = &MockProductImageRepository{}
	suite.mockMinioService = &MockMinioService{}
	suite.mockCacheService = &testhelpers.MockCacheService{}
	suite.productService = NewProductService(
		suite.mockProductRepo,
		suite.mockInventoryRepo,
		suite.mockCategoryRepo,
		suite.mockImageRepo,
		suite.mockMinioService,
		suite.mockCacheService,
	)
	suite.tenantID = uuid.New()
	suite.ctx = context.Background()
}

func (suite *ProductServiceTestSuite) TearDownTest() {
	suite.mockProductRepo.AssertExpectations(suite.T())
	suite.mockInventoryRepo.AssertExpectations(suite.T())
	suite.mockCategoryRepo.AssertExpectations(suite.T())
	suite.mockImageRepo.AssertExpectations(suite.T())
	suite.mockMinioService.AssertExpectations(suite.T())
	suite.mockCacheService.AssertExpectations(suite.T())
}

func TestProductServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProductServiceTestSuite))
}

func (suite *ProductServiceTestSuite) TestCreateProduct() {
	suite.T().Run("Successful Product Creation", func(t *testing.T) {
		// Arrange
		product := &models.Product{
			ID:          uuid.New(),
			TenantID:    suite.tenantID,
			Name:        "Test Product",
			Description: testhelpers.StringPtr("Test product description"),
			Quantity:    100,
			UnitPrice:   29.99,
			Barcode:     testhelpers.StringPtr("123456789012"),
			Status:      "active",
		}

		suite.mockProductRepo.On("Create", suite.ctx, product).Return(nil)

		// Act
		err := suite.productService.Create(suite.ctx, suite.tenantID, product)

		// Assert
		assert.NoError(t, err)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Product Creation with Invalid Data", func(t *testing.T) {
		// Arrange
		product := &models.Product{
			ID:          uuid.New(),
			TenantID:    suite.tenantID,
			Name:        "", // Invalid empty name
			Quantity:    -1, // Invalid negative quantity
			UnitPrice:   -10.99, // Invalid negative price
		}

		//Act
		err := suite.productService.Create(suite.ctx, suite.tenantID, product)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	suite.T().Run("Repository Error", func(t *testing.T) {
		// Arrange
		product := &models.Product{
			ID:          uuid.New(),
			TenantID:    suite.tenantID,
			Name:        "Test Product",
			Quantity:    10,
			UnitPrice:   5.99,
		}

		suite.mockProductRepo.On("Create", suite.ctx, product).Return(errors.New("database error"))

		// Act
		err := suite.productService.Create(suite.ctx, suite.tenantID, product)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create product")
		suite.mockProductRepo.AssertExpectations(suite.T())
	})
}

func (suite *ProductServiceTestSuite) TestUpdateProduct() {
	suite.T().Run("Successful Product Update", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		existingProduct := &models.Product{
			ID:          productID,
			TenantID:    suite.tenantID,
			Name:        "Original Product",
			Description: testhelpers.StringPtr("Original description"),
			Quantity:    100,
			UnitPrice:   29.99,
		}

		updateData := &models.ProductUpdateData{
			Name:        testhelpers.StringPtr("Updated Product"),
			Description: testhelpers.StringPtr("Updated description"),
			UnitPrice:   testhelpers.FloatPtr(39.99),
			Quantity:    testhelpers.IntPtr(150),
		}

		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, productID).Return(existingProduct, nil)
		suite.mockProductRepo.On("Update", suite.ctx, mock.AnythingOfType("*models.Product")).Return(nil)

		// Act
		err := suite.productService.Update(suite.ctx, suite.tenantID, productID, updateData)

		// Assert
		assert.NoError(t, err)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Product Not Found", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		updateData := &models.ProductUpdateData{
			Name: testhelpers.StringPtr("Updated Product"),
		}

		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, productID).Return(nil, errors.New("product not found"))

		// Act
		err := suite.productService.Update(suite.ctx, suite.tenantID, productID, updateData)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Invalid Update Data", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		existingProduct := &models.Product{
			ID:        productID,
			TenantID:  suite.tenantID,
			Name:      "Original Product",
			Quantity:  100,
			UnitPrice: 29.99,
		}

		updateData := &models.ProductUpdateData{
			Quantity:  testhelpers.IntPtr(-10), // Invalid negative quantity
			UnitPrice: testhelpers.FloatPtr(-5.99), // Invalid negative price
		}

		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, productID).Return(existingProduct, nil)

		// Act
		err := suite.productService.Update(suite.ctx, suite.tenantID, productID, updateData)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
		suite.mockProductRepo.AssertExpectations(suite.T())
	})
}

func (suite *ProductServiceTestSuite) TestDeleteProduct() {
	suite.T().Run("Successful Product Deletion", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		existingProduct := &models.Product{
			ID:          productID,
			TenantID:    suite.tenantID,
			Name:        "Product to Delete",
			Quantity:    100,
			UnitPrice:   29.99,
		}

		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, productID).Return(existingProduct, nil)
		suite.mockProductRepo.On("Delete", suite.ctx, suite.tenantID, productID).Return(nil)

		// Act
		err := suite.productService.Delete(suite.ctx, suite.tenantID, productID)

		// Assert
		assert.NoError(t, err)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Product Not Found for Deletion", func(t *testing.T) {
		// Arrange
		productID := uuid.New()

		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, productID).Return(nil, errors.New("product not found"))

		// Act
		err := suite.productService.Delete(suite.ctx, suite.tenantID, productID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Repository Deletion Error", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		existingProduct := &models.Product{
			ID:        productID,
			TenantID:  suite.tenantID,
			Name:      "Product to Delete",
			Quantity:  100,
			UnitPrice: 29.99,
		}

		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, productID).Return(existingProduct, nil)
		suite.mockProductRepo.On("Delete", suite.ctx, suite.tenantID, productID).Return(errors.New("database error"))

		// Act
		err := suite.productService.Delete(suite.ctx, suite.tenantID, productID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete product")
		suite.mockProductRepo.AssertExpectations(suite.T())
	})
}

func (suite *ProductServiceTestSuite) TestSearchProducts() {
	suite.T().Run("Successful Product Search", func(t *testing.T) {
		// Arrange
		filter := &models.ProductSearchFilter{
			Query:     "Test Product",
			Limit:     10,
			Offset:    0,
			SortBy:    "name",
			SortOrder: "asc",
		}

		mockProducts := []*models.Product{
			{
				ID:        uuid.New(),
				TenantID:  suite.tenantID,
				Name:      "Test Product 1",
				Quantity:  100,
				UnitPrice: 29.99,
			},
			{
				ID:        uuid.New(),
				TenantID:  suite.tenantID,
				Name:      "Test Product 2",
				Quantity:  50,
				UnitPrice: 19.99,
			},
		}

		suite.mockProductRepo.On("AdvancedSearch", suite.ctx, suite.tenantID, filter).Return(mockProducts, nil)

		// Act
		results, err := suite.productService.Search(suite.ctx, suite.tenantID, filter)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "Test Product 1", results[0].Name)
		assert.Equal(t, "Test Product 2", results[1].Name)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Search with Filters", func(t *testing.T) {
		// Arrange
		categoryID := uuid.New()
		minQuantity := 10
		minPrice := 5.0
		maxPrice := 50.0
		
		filter := &models.ProductSearchFilter{
			Query:        "Product",
			CategoryID:   &categoryID,
			MinQuantity:  &minQuantity,
			MaxQuantity:  nil,
			MinPrice:     &minPrice,
			MaxPrice:     &maxPrice,
			Limit:        10,
			Offset:       0,
		}

		mockProducts := []*models.Product{}
		suite.mockProductRepo.On("AdvancedSearch", suite.ctx, suite.tenantID, filter).Return(mockProducts, nil)

		// Act
		results, err := suite.productService.Search(suite.ctx, suite.tenantID, filter)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 0)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Search with Empty Query", func(t *testing.T) {
		// Arrange
		filter := &models.ProductSearchFilter{
			Query:     "",
			Limit:     10,
			Offset:    0,
		}

		mockProducts := []*models.Product{
			{
				ID:        uuid.New(),
				TenantID:  suite.tenantID,
				Name:      "Product 1",
				Quantity:  25,
				UnitPrice: 10.99,
			},
		}
		suite.mockProductRepo.On("AdvancedSearch", suite.ctx, suite.tenantID, filter).Return(mockProducts, nil)

		// Act
		results, err := suite.productService.Search(suite.ctx, suite.tenantID, filter)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Repository Error", func(t *testing.T) {
		// Arrange
		filter := &models.ProductSearchFilter{
			Query:  "Test Product",
			Limit:  10,
			Offset: 0,
		}

		suite.mockProductRepo.On("AdvancedSearch", suite.ctx, suite.tenantID, filter).Return(nil, errors.New("database error"))

		// Act
		results, err := suite.productService.Search(suite.ctx, suite.tenantID, filter)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to search products")
		assert.Nil(t, results)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})
}

func (suite *ProductServiceTestSuite) TestBulkUpdateProducts() {
	suite.T().Run("Successful Bulk Product Update", func(t *testing.T) {
		// Arrange
		product1ID := uuid.New()
		product2ID := uuid.New()
		product1 := &models.Product{
			ID:        product1ID,
			TenantID:  suite.tenantID,
			Name:      "Product 1",
			Quantity:  100,
			UnitPrice: 29.99,
		}
		product2 := &models.Product{
			ID:        product2ID,
			TenantID:  suite.tenantID,
			Name:      "Product 2",
			Quantity:  50,
			UnitPrice: 19.99,
		}

		updates := []models.ProductBulkUpdate{
			{
				ProductIDs:       []uuid.UUID{product1ID},
				Description:      testhelpers.StringPtr("Bulk Updated Product 1"),
				UnitPriceChange: testhelpers.FloatPtr(19.99),
				UnitPriceMode:    "absolute",
			},
			{
				ProductIDs:       []uuid.UUID{product2ID},
				Description:      testhelpers.StringPtr("Bulk Updated Product 2"),
				UnitPriceChange: testhelpers.FloatPtr(29.99),
				UnitPriceMode:    "absolute",
			},
		}

		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, product1ID).Return(product1, nil)
		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, product2ID).Return(product2, nil)
		suite.mockProductRepo.On("Update", suite.ctx, mock.AnythingOfType("*models.Product")).Return(nil).Twice()

		// Act
		results, err := suite.productService.BulkUpdate(suite.ctx, suite.tenantID, updates)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Bulk Update with Invalid Data", func(t *testing.T) {
		// Arrange
		updates := []models.ProductBulkUpdate{
			{
				ProductIDs:       []uuid.UUID{uuid.New()},
				UnitPriceChange: testhelpers.FloatPtr(-10.99), // Invalid negative price change
				UnitPriceMode:    "absolute",
			},
		}

		// Act
		results, err := suite.productService.BulkUpdate(suite.ctx, suite.tenantID, updates)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
		assert.Nil(t, results)
	})

	suite.T().Run("Product Not Found in Bulk Update", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		updates := []models.ProductBulkUpdate{
			{
				ProductIDs: []uuid.UUID{productID},
				UnitPriceChange: testhelpers.FloatPtr(19.99),
				UnitPriceMode: "relative",
			},
		}

		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, productID).Return(nil, errors.New("product not found"))

		// Act
		results, err := suite.productService.BulkUpdate(suite.ctx, suite.tenantID, updates)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, results)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})
}

func (suite *ProductServiceTestSuite) TestValidateProduct() {
	suite.T().Run("Valid Product", func(t *testing.T) {
		// Arrange
		product := &models.Product{
			ID:          uuid.New(),
			TenantID:    suite.tenantID,
			Name:        "Valid Product",
			Description: testhelpers.StringPtr("Valid description"),
			Quantity:    100,
			UnitPrice:   25.50,
			Barcode:     testhelpers.StringPtr("123456789012"),
		}

		// Act & Assert - Validate the product
		err := suite.productService.Validate(product)
		assert.NoError(t, err)
	})

	suite.T().Run("Invalid Product - Empty Name", func(t *testing.T) {
		// Arrange
		product := &models.Product{
			ID:        uuid.New(),
			TenantID:  suite.tenantID,
			Name:      "", // Invalid empty name
			Quantity:  100,
			UnitPrice: 25.50,
		}

		// Act & Assert
		err := suite.productService.Validate(product)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	suite.T().Run("Invalid Product - Negative Price", func(t *testing.T) {
		// Arrange
		product := &models.Product{
			ID:        uuid.New(),
			TenantID:  suite.tenantID,
			Name:      "Negative Price Product",
			Quantity:  100,
			UnitPrice: -10.00, // Invalid negative price
		}

		// Act & Assert
		err := suite.productService.Validate(product)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unit price must be non-negative")
	})

	suite.T().Run("Invalid Product - Negative Quantity", func(t *testing.T) {
		// Arrange
		product := &models.Product{
			ID:        uuid.New(),
			TenantID:  suite.tenantID,
			Name:      "Negative Quantity Product",
			Quantity:  -5, // Invalid negative quantity
			UnitPrice: 25.50,
		}

		// Act & Assert
		err := suite.productService.Validate(product)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity must be non-negative")
	})

	suite.T().Run("Valid Product with Optional Fields", func(t *testing.T) {
		// Arrange
		batchNum := "BATCH123"
		expiry := time.Now().AddDate(1, 0, 0)
		product := &models.Product{
			ID:          uuid.New(),
			TenantID:    suite.tenantID,
			Name:        "Product with Options",
			Description: testhelpers.StringPtr("Description"),
			Quantity:    50,
			UnitPrice:   15.99,
			BatchNumber: &batchNum,
			ExpiryDate:  &expiry,
			Barcode:     testhelpers.StringPtr("123456789012"),
			Status:      "active",
		}

		// Act & Assert
		err := suite.productService.Validate(product)
		assert.NoError(t, err)
	})
}

func (suite *ProductServiceTestSuite) TestProductAnalytics() {
	suite.T().Run("Category Analytics", func(t *testing.T) {
		// Arrange
		categoryID := uuid.New()
		expectedAnalytics := map[uuid.UUID]int{
			categoryID: 10,
		}

		suite.mockProductRepo.On("CategoryAnalytics", suite.ctx, suite.tenantID).Return(expectedAnalytics, nil)

		// Act
		analytics, err := suite.productService.GetCategoryAnalytics(suite.ctx, suite.tenantID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedAnalytics, analytics)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Repository Error in Analytics", func(t *testing.T) {
		// Arrange
		suite.mockProductRepo.On("CategoryAnalytics", suite.ctx, suite.tenantID).Return(nil, errors.New("database error"))

		// Act
		analytics, err := suite.productService.GetCategoryAnalytics(suite.ctx, suite.tenantID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get category analytics")
		assert.Nil(t, analytics)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})
}

// Additional test methods that might be in the service

func (suite *ProductServiceTestSuite) TestGetProductByID() {
	suite.T().Run("Successful GetByID", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		expectedProduct := &models.Product{
			ID:        productID,
			TenantID:  suite.tenantID,
			Name:      "Test Product",
			Quantity:  100,
			UnitPrice: 29.99,
		}

		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, productID).Return(expectedProduct, nil)

		// Act
		product, err := suite.productService.GetByID(suite.ctx, suite.tenantID, productID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedProduct.ID, product.ID)
		assert.Equal(t, expectedProduct.Name, product.Name)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Product Not Found", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		suite.mockProductRepo.On("GetByID", suite.ctx, suite.tenantID, productID).Return(nil, errors.New("product not found"))

		// Act
		product, err := suite.productService.GetByID(suite.ctx, suite.tenantID, productID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, product)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})
}

func (suite *ProductServiceTestSuite) TestListProducts() {
	suite.T().Run("Successful List", func(t *testing.T) {
		// Arrange
		expectedProducts := []*models.Product{
			{
				ID:        uuid.New(),
				TenantID:  suite.tenantID,
				Name:      "Product 1",
				Quantity:  100,
				UnitPrice: 29.99,
			},
			{
				ID:        uuid.New(),
				TenantID:  suite.tenantID,
				Name:      "Product 2",
				Quantity:  50,
				UnitPrice: 19.99,
			},
		}

		suite.mockProductRepo.On("List", suite.ctx, suite.tenantID, 10, 0).Return(expectedProducts, nil)

		// Act
		products, err := suite.productService.List(suite.ctx, suite.tenantID, 10, 0)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, products, 2)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Repository Error in List", func(t *testing.T) {
		// Arrange
		suite.mockProductRepo.On("List", suite.ctx, suite.tenantID, 10, 0).Return(nil, errors.New("database error"))

		// Act
		products, err := suite.productService.List(suite.ctx, suite.tenantID, 10, 0)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, products)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})
}

func (suite *ProductServiceTestSuite) TestGetByBarcode() {
	suite.T().Run("Successful GetByBarcode", func(t *testing.T) {
		// Arrange
		barcode := "123456789012"
		expectedProduct := &models.Product{
			ID:        uuid.New(),
			TenantID:  suite.tenantID,
			Name:      "Product by Barcode",
			Barcode:   &barcode,
			Quantity:  100,
			UnitPrice: 29.99,
		}

		suite.mockProductRepo.On("GetByBarcode", suite.ctx, suite.tenantID, barcode).Return(expectedProduct, nil)

		// Act
		product, err := suite.productService.GetByBarcode(suite.ctx, suite.tenantID, barcode)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, barcode, *product.Barcode)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})

	suite.T().Run("Product Not Found by Barcode", func(t *testing.T) {
		// Arrange
		barcode := "NONEXISTENT"
		suite.mockProductRepo.On("GetByBarcode", suite.ctx, suite.tenantID, barcode).Return(nil, errors.New("product not found"))

		// Act
		product, err := suite.productService.GetByBarcode(suite.ctx, suite.tenantID, barcode)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, product)
		suite.mockProductRepo.AssertExpectations(suite.T())
	})
}
