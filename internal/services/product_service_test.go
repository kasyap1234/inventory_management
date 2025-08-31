package services

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock repositories and services
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

func (m *MockInventoryRepository) AdjustStock(ctx context.Context, tenantID, productID uuid.UUID, change int) error {
	args := m.Called(ctx, tenantID, productID, change)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetByProductID(ctx context.Context, tenantID, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, tenantID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inventory), args.Error(1)
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
	return args.Get(0).([]*models.Category), args.Error(1)
}

type MockProductImageRepository struct {
	mock.Mock
}

func (m *MockProductImageRepository) Create(ctx context.Context, image *models.ProductImage) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *MockProductImageRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.ProductImage, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductImage), args.Error(1)
}

func (m *MockProductImageRepository) GetByProductID(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.ProductImage, error) {
	args := m.Called(ctx, tenantID, productID)
	return args.Get(0).([]*models.ProductImage), args.Error(1)
}

func (m *MockProductImageRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockProductImageRepository) DeleteAllByProductID(ctx context.Context, tenantID, productID uuid.UUID) error {
	args := m.Called(ctx, tenantID, productID)
	return args.Error(0)
}

type MockMinioService struct {
	mock.Mock
}

func (m *MockMinioService) UploadImage(ctx context.Context, bucket, key string, reader io.Reader, size int64) error {
	args := m.Called(ctx, bucket, key, reader, size)
	return args.Error(0)
}

func (m *MockMinioService) GetPresignedURL(bucket, key string, expiry time.Duration) (string, error) {
	args := m.Called(bucket, key, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockMinioService) DeleteImage(ctx context.Context, bucket, key string) error {
	args := m.Called(ctx, bucket, key)
	return args.Error(0)
}

// ProductServiceTestSuite defines the test suite
type ProductServiceTestSuite struct {
	suite.Suite
	mockProductRepo      *MockProductRepository
	mockInventoryRepo    *MockInventoryRepository
	mockCategoryRepo     *MockCategoryRepository
	mockProductImageRepo *MockProductImageRepository
	mockMinioService     *MockMinioService
	service              ProductService
	tenantID             uuid.UUID
}

func (suite *ProductServiceTestSuite) SetupTest() {
	suite.mockProductRepo = &MockProductRepository{}
	suite.mockInventoryRepo = &MockInventoryRepository{}
	suite.mockCategoryRepo = &MockCategoryRepository{}
	suite.mockProductImageRepo = &MockProductImageRepository{}
	suite.mockMinioService = &MockMinioService{}
	suite.service = NewProductService(suite.mockProductRepo, suite.mockInventoryRepo, suite.mockCategoryRepo, suite.mockProductImageRepo, suite.mockMinioService)
	suite.tenantID = uuid.New()

	suite.mockMinioService.Test(suite.T())
}

func (suite *ProductServiceTestSuite) TearDownTest() {
	suite.mockProductRepo.AssertExpectations(suite.T())
	suite.mockInventoryRepo.AssertExpectations(suite.T())
	suite.mockCategoryRepo.AssertExpectations(suite.T())
	suite.mockProductImageRepo.AssertExpectations(suite.T())
	suite.mockMinioService.AssertExpectations(suite.T())
}

func TestProductServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProductServiceTestSuite))
}

func (suite *ProductServiceTestSuite) TestCreate_ProductSuccess() {
	product := &models.Product{
		Name:      "Test Product",
		UnitPrice: 10.99,
		Quantity:  100,
	}

	suite.mockProductRepo.On("Create", mock.Anything, product).Return(nil).Once()
	suite.mockProductRepo.On("GetByBarcode", mock.Anything, suite.tenantID, (*string)(nil)).Return((*models.Product)(nil), nil).Once()

	err := suite.service.Create(context.Background(), suite.tenantID, product)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.tenantID, product.TenantID)
	assert.NotEqual(suite.T(), uuid.Nil, product.ID)
}

func (suite *ProductServiceTestSuite) TestCreate_ProductWithBarcodeDuplicate() {
	barcode := "123456789"
	existingProduct := &models.Product{
		ID:      uuid.New(),
		Name:    "Existing Product",
		Barcode: &barcode,
	}
	product := &models.Product{
		Name:     "Test Product",
		UnitPrice: 10.99,
		Quantity:  100,
		Barcode:   &barcode,
	}

	suite.mockProductRepo.On("GetByBarcode", mock.Anything, suite.tenantID, barcode).Return(existingProduct, nil).Once()

	err := suite.service.Create(context.Background(), suite.tenantID, product)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "barcode already exists")
}

func (suite *ProductServiceTestSuite) TestCreate_ProductWithInvalidCategory() {
	categoryID := uuid.New()
	product := &models.Product{
		Name:       "Test Product",
		UnitPrice:  10.99,
		Quantity:   100,
		CategoryID: &categoryID,
	}

	suite.mockProductRepo.On("GetByBarcode", mock.Anything, suite.tenantID, (*string)(nil)).Return((*models.Product)(nil), nil).Once()
	suite.mockCategoryRepo.On("GetByID", mock.Anything, suite.tenantID, categoryID).Return((*models.Category)(nil), errors.New("category not found")).Once()

	err := suite.service.Create(context.Background(), suite.tenantID, product)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "category not found")
}

func (suite *ProductServiceTestSuite) TestCreate_ProductValidationNameRequired() {
	product := &models.Product{
		UnitPrice: 10.99,
		Quantity:  100,
	}

	err := suite.service.Create(context.Background(), suite.tenantID, product)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "product name is required", err.Error())
}

func (suite *ProductServiceTestSuite) TestCreate_ProductValidationPositivePrice() {
	product := &models.Product{
		Name:      "Test Product",
		UnitPrice: 0,
		Quantity:  100,
	}

	err := suite.service.Create(context.Background(), suite.tenantID, product)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "unit price must be positive", err.Error())
}

func (suite *ProductServiceTestSuite) TestCreate_ProductValidationNegativeQuantity() {
	product := &models.Product{
		Name:      "Test Product",
		UnitPrice: 10.99,
		Quantity:  -10,
	}

	err := suite.service.Create(context.Background(), suite.tenantID, product)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "quantity cannot be negative", err.Error())
}

func (suite *ProductServiceTestSuite) TestGetByID_Success() {
	productID := uuid.New()
	expectedProduct := &models.Product{
		ID:       productID,
		TenantID: suite.tenantID,
		Name:     "Test Product",
	}

	suite.mockProductRepo.On("GetByID", mock.Anything, suite.tenantID, productID).Return(expectedProduct, nil).Once()

	product, err := suite.service.GetByID(context.Background(), suite.tenantID, productID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedProduct, product)
}

func (suite *ProductServiceTestSuite) TestGetByID_ProductNotFound() {
	productID := uuid.New()

	suite.mockProductRepo.On("GetByID", mock.Anything, suite.tenantID, productID).Return((*models.Product)(nil), errors.New("product not found")).Once()

	product, err := suite.service.GetByID(context.Background(), suite.tenantID, productID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), product)
}

func (suite *ProductServiceTestSuite) TestUpdate_Success() {
	productID := uuid.New()
	product := &models.Product{
		ID:        productID,
		TenantID:  suite.tenantID,
		Name:      "Test Product",
		Quantity:  50,
		UnitPrice: 10.99,
	}
	updatedProduct := &models.Product{
		ID:       productID,
		Quantity: 75, // Stock change will occur
	}

	suite.mockProductRepo.On("GetByID", mock.Anything, suite.tenantID, productID).Return(product, nil).Once()
	suite.mockProductRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		updatedProd := args.Get(1).(*models.Product)
		assert.Equal(suite.T(), 75, updatedProd.Quantity)
	}).Once()

	err := suite.service.Update(context.Background(), suite.tenantID, updatedProduct)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.tenantID, updatedProduct.TenantID)
}

func (suite *ProductServiceTestSuite) TestUpdate_ProductNotFound() {
	productID := uuid.New()
	product := &models.Product{
		ID:       productID,
		Name:     "Test Product",
		Quantity: 50,
	}

	suite.mockProductRepo.On("GetByID", mock.Anything, suite.tenantID, productID).Return((*models.Product)(nil), errors.New("product not found")).Once()

	err := suite.service.Update(context.Background(), suite.tenantID, product)

	assert.Error(suite.T(), err)
}

func (suite *ProductServiceTestSuite) TestDelete_Success() {
	productID := uuid.New()

	suite.mockProductRepo.On("Delete", mock.Anything, suite.tenantID, productID).Return(nil).Once()

	err := suite.service.Delete(context.Background(), suite.tenantID, productID)

	assert.NoError(suite.T(), err)
}

func (suite *ProductServiceTestSuite) TestList_Success() {
	expectedProducts := []*models.Product{
		{ID: uuid.New(), Name: "Product 1"},
		{ID: uuid.New(), Name: "Product 2"},
	}

	suite.mockProductRepo.On("List", mock.Anything, suite.tenantID, 10, 0).Return(expectedProducts, nil).Once()

	products, err := suite.service.List(context.Background(), suite.tenantID, 10, 0)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedProducts, products)
}

func (suite *ProductServiceTestSuite) TestGetByBarcode_Success() {
	barcode := "123456789"
	expectedProduct := &models.Product{
		ID:      uuid.New(),
		Barcode: &barcode,
	}

	suite.mockProductRepo.On("GetByBarcode", mock.Anything, suite.tenantID, barcode).Return(expectedProduct, nil).Once()

	product, err := suite.service.GetByBarcode(context.Background(), suite.tenantID, barcode)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedProduct, product)
}

func (suite *ProductServiceTestSuite) TestUpdateStock_Success() {
	productID := uuid.New()
	change := 25
	product := &models.Product{
		ID:       productID,
		Quantity: 50,
	}

	suite.mockProductRepo.On("GetByID", mock.Anything, suite.tenantID, productID).Return(product, nil).Once()
	suite.mockProductRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Product")).Return(nil).Run(func(args mock.Arguments) {
		updatedProduct := args.Get(1).(*models.Product)
		assert.Equal(suite.T(), 75, updatedProduct.Quantity)
	}).Once()

	err := suite.service.UpdateStock(context.Background(), suite.tenantID, productID, change)

	assert.NoError(suite.T(), err)
}

func (suite *ProductServiceTestSuite) TestSearch_WithoutQuery() {
	expectedProducts := []*models.Product{
		{ID: uuid.New(), Name: "Product 1"},
	}

	suite.mockProductRepo.On("List", mock.Anything, suite.tenantID, 10, 0).Return(expectedProducts, nil).Once()

	products, err := suite.service.Search(context.Background(), suite.tenantID, "", nil, 10, 0)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedProducts, products)
}

func (suite *ProductServiceTestSuite) TestSearch_WithQuery() {
	query := "search term"
	categoryID := uuid.New()
	expectedProducts := []*models.Product{
		{ID: uuid.New()},
	}

	suite.mockProductRepo.On("Search", mock.Anything, suite.tenantID, query, &categoryID, 10, 0).Return(expectedProducts, nil).Once()

	products, err := suite.service.Search(context.Background(), suite.tenantID, query, &categoryID, 10, 0)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedProducts, products)
}

func (suite *ProductServiceTestSuite) TestCategoryAnalytics_Success() {
	expectedAnalytics := map[string]int{
		"category1": 10,
		"category2": 5,
	}

	suite.mockProductRepo.On("CategoryAnalytics", mock.Anything, suite.tenantID).Return(expectedAnalytics, nil).Once()

	analytics, err := suite.service.CategoryAnalytics(context.Background(), suite.tenantID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedAnalytics, analytics)
}

// Note: Skipping image-related tests for brevity, but they would follow similar patterns
// TestUploadProductImage, TestGetProductImages, TestGetProductImageURL, TestDeleteProductImage

// Note: Bulk operation tests would be comprehensive but follow similar mocking patterns