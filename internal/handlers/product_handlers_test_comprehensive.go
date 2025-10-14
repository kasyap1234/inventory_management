//go:build handlers_tests
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Mock ProductService for testing
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) Create(ctx context.Context, tenantID uuid.UUID, product *models.Product) error {
	args := m.Called(ctx, tenantID, product)
	return args.Error(0)
}

func (m *MockProductService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) Update(ctx context.Context, tenantID uuid.UUID, product *models.Product) error {
	args := m.Called(ctx, tenantID, product)
	return args.Error(0)
}

func (m *MockProductService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockProductService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductService) GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*models.Product, error) {
	args := m.Called(ctx, tenantID, barcode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) UpdateStock(ctx context.Context, tenantID, productID uuid.UUID, change int) error {
	args := m.Called(ctx, tenantID, productID, change)
	return args.Error(0)
}

func (m *MockProductService) Search(ctx context.Context, tenantID uuid.UUID, query string, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, tenantID, query, categoryID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductService) CategoryAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]int, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockProductService) UploadProductImage(ctx context.Context, tenantID, productID uuid.UUID, filename string, reader interface{}, size int64, altText *string) error {
	args := m.Called(ctx, tenantID, productID, filename, reader, size, altText)
	return args.Error(0)
}

func (m *MockProductService) GetProductImages(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.ProductImage, error) {
	args := m.Called(ctx, tenantID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProductImage), args.Error(1)
}

func (m *MockProductService) GetProductImageURL(ctx context.Context, tenantID, imageID uuid.UUID, expiry time.Duration) (string, error) {
	args := m.Called(ctx, tenantID, imageID, expiry)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *MockProductService) DeleteProductImage(ctx context.Context, tenantID, imageID uuid.UUID) error {
	args := m.Called(ctx, tenantID, imageID)
	return args.Error(0)
}

func (m *MockProductService) BulkUpdateProducts(ctx context.Context, tenantID uuid.UUID, bulkUpdate *models.ProductBulkUpdate) (*models.BulkOperationResult, error) {
	args := m.Called(ctx, tenantID, bulkUpdate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BulkOperationResult), args.Error(1)
}

func (m *MockProductService) BulkCreateProducts(ctx context.Context, tenantID uuid.UUID, bulkCreate *models.ProductBulkCreate) (*models.BulkOperationResult, error) {
	args := m.Called(ctx, tenantID, bulkCreate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BulkOperationResult), args.Error(1)
}

func (m *MockProductService) BulkDeleteProducts(ctx context.Context, tenantID uuid.UUID, productIDs []uuid.UUID) (*models.BulkOperationResult, error) {
	args := m.Called(ctx, tenantID, productIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BulkOperationResult), args.Error(1)
}

// Mock RBACMiddleware for testing
type MockRBACMiddleware struct {
	mock.Mock
}

func (m *MockRBACMiddleware) RequirePermission(permission string) echo.MiddlewareFunc {
	args := m.Called(permission)
	return args.Get(0).(func(echo.HandlerFunc) echo.HandlerFunc)
}

// ProductHandlers test suite
type ProductHandlersComprehensiveTestSuite struct {
	suite.Suite
	mockProductService   *MockProductService
	mockRBACMiddleware   *MockRBACMiddleware
	handlers             *ProductHandlers
	echo                 *echo.Echo
	tenantID             uuid.UUID
	userID               uuid.UUID
	ctx                  context.Context
}

func (suite *ProductHandlersComprehensiveTestSuite) SetupTest() {
	suite.mockProductService = &MockProductService{}
	suite.mockRBACMiddleware = &MockRBACMiddleware{}
	suite.handlers = NewProductHandlers(suite.mockProductService, suite.mockRBACMiddleware)
	suite.echo = echo.New()
	suite.tenantID = uuid.New()
	suite.userID = uuid.New()
	suite.ctx = context.Background()

	// Setup mock RBAC middleware to just pass through
	suite.mockRBACMiddleware.On("RequirePermission", mock.AnythingOfType("string")).Return(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Set("tenant_id", suite.tenantID)
				c.Set("user_id", suite.userID)
				return next(c)
			}
		},
	)
}

func (suite *ProductHandlersComprehensiveTestSuite) TearDownTest() {
	suite.mockProductService.AssertExpectations(suite.T())
	suite.mockRBACMiddleware.AssertExpectations(suite.T())
}

func TestProductHandlersComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(ProductHandlersComprehensiveTestSuite))
}

func (suite *ProductHandlersComprehensiveTestSuite) TestCreateProduct() {
	suite.T().Run("Successful Product Creation", func(t *testing.T) {
		// Arrange
		productRequest := map[string]interface{}{
			"name":        "Test Product",
			"quantity":    100,
			"unit_price":  29.99,
			"description": "Test product description",
			"barcode":     "123456789012",
		}

		requestBody, _ := json.Marshal(productRequest)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)

		// Set tenant and user context
		c.Set("tenant_id", suite.tenantID)
		c.Set("user_id", suite.userID)

		suite.mockProductService.On("Create", context.Background(), suite.tenantID, mock.AnythingOfType("*models.Product")).Return(nil)

		// Act
		err := suite.handlers.CreateProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		suite.mockProductService.AssertExpectations(suite.T())
	})

	suite.T().Run("Invalid Request Body", func(t *testing.T) {
		// Arrange
		productRequest := map[string]interface{}{
			"name":       "", // Invalid empty name
			"quantity":   -10, // Invalid negative quantity
			"unit_price": -5.99, // Invalid negative price
		}

		requestBody, _ := json.Marshal(productRequest)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)

		// Act
		err := suite.handlers.CreateProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])
	})

	suite.T().Run("Service Error", func(t *testing.T) {
		// Arrange
		productRequest := map[string]interface{}{
			"name":       "Test Product",
			"quantity":   100,
			"unit_price": 29.99,
		}

		requestBody, _ := json.Marshal(productRequest)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)

		c.Set("tenant_id", suite.tenantID)
		c.Set("user_id", suite.userID)

		suite.mockProductService.On("Create", context.Background(), suite.tenantID, mock.AnythingOfType("*models.Product")).Return(errors.New("database error"))

		// Act
		err := suite.handlers.CreateProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})
}

func (suite *ProductHandlersComprehensiveTestSuite) TestGetProduct() {
	suite.T().Run("Successful Product Retrieval", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		expectedProduct := &models.Product{
			ID:        productID,
			TenantID:  suite.tenantID,
			Name:      "Test Product",
			Quantity:  100,
			UnitPrice: 29.99,
		}

		req := httptest.NewRequest(http.MethodGet, "/products/"+productID.String(), nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID.String())
		c.Set("tenant_id", suite.tenantID)

		suite.mockProductService.On("GetByID", context.Background(), suite.tenantID, productID).Return(expectedProduct, nil)

		// Act
		err := suite.handlers.GetProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		data := response["data"].(map[string]interface{})
		assert.Equal(t, productID.String(), data["id"])
		assert.Equal(t, "Test Product", data["name"])
		suite.mockProductService.AssertExpectations(suite.T())
	})

	suite.T().Run("Invalid UUID", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/products/invalid-uuid", nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("invalid-uuid")
		c.Set("tenant_id", suite.tenantID)

		// Act
		err := suite.handlers.GetProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	suite.T().Run("Product Not Found", func(t *testing.T) {
		// Arrange
		productID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/products/"+productID.String(), nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID.String())
		c.Set("tenant_id", suite.tenantID)

		suite.mockProductService.On("GetByID", context.Background(), suite.tenantID, productID).Return(nil, errors.New("product not found"))

		// Act
		err := suite.handlers.GetProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})
}

func (suite *ProductHandlersComprehensiveTestSuite) TestUpdateProduct() {
	suite.T().Run("Successful Product Update", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		updateRequest := map[string]interface{}{
			"name":        "Updated Product",
			"quantity":    150,
			"unit_price":  39.99,
			"description": "Updated description",
		}

		requestBody, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest(http.MethodPut, "/products/"+productID.String(), bytes.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID.String())
		c.Set("tenant_id", suite.tenantID)
		c.Set("user_id", suite.userID)

		// Mock service expectations
		existingProduct := &models.Product{
			ID:        productID,
			TenantID:  suite.tenantID,
			Name:      "Original Product",
			Quantity:  100,
			UnitPrice: 29.99,
		}

		suite.mockProductService.On("GetByID", context.Background(), suite.tenantID, productID).Return(existingProduct, nil)
		suite.mockProductService.On("Update", context.Background(), suite.tenantID, mock.AnythingOfType("*models.Product")).Return(nil)

		// Act
		err := suite.handlers.UpdateProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})

	suite.T().Run("Invalid Update Data", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		updateRequest := map[string]interface{}{
			"name":       "", // Invalid empty name
			"unit_price": -10.99, // Invalid negative price
		}

		requestBody, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest(http.MethodPut, "/products/"+productID.String(), bytes.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID.String())
		c.Set("tenant_id", suite.tenantID)

		// Act
		err := suite.handlers.UpdateProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	suite.T().Run("Product Not Found for Update", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		updateRequest := map[string]interface{}{
			"name": "Updated Product",
		}

		requestBody, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest(http.MethodPut, "/products/"+productID.String(), bytes.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID.String())
		c.Set("tenant_id", suite.tenantID)
		c.Set("user_id", suite.userID)

		suite.mockProductService.On("GetByID", context.Background(), suite.tenantID, productID).Return(nil, errors.New("product not found"))

		// Act
		err := suite.handlers.UpdateProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})
}

func (suite *ProductHandlersComprehensiveTestSuite) TestDeleteProduct() {
	suite.T().Run("Successful Product Deletion", func(t *testing.T) {
		// Arrange
		productID := uuid.New()
		existingProduct := &models.Product{
			ID:        productID,
			TenantID:  suite.tenantID,
			Name:      "Product to Delete",
			Quantity:  100,
			UnitPrice: 29.99,
		}

		req := httptest.NewRequest(http.MethodDelete, "/products/"+productID.String(), nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID.String())
		c.Set("tenant_id", suite.tenantID)
		c.Set("user_id", suite.userID)

		suite.mockProductService.On("GetByID", context.Background(), suite.tenantID, productID).Return(existingProduct, nil)
		suite.mockProductService.On("Delete", context.Background(), suite.tenantID, productID).Return(nil)

		// Act
		err := suite.handlers.DeleteProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})

	suite.T().Run("Delete Non-Existent Product", func(t *testing.T) {
		// Arrange
		productID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/products/"+productID.String(), nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID.String())
		c.Set("tenant_id", suite.tenantID)
		c.Set("user_id", suite.userID)

		suite.mockProductService.On("GetByID", context.Background(), suite.tenantID, productID).Return(nil, errors.New("product not found"))

		// Act
		err := suite.handlers.DeleteProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})
}

func (suite *ProductHandlersComprehensiveTestSuite) TestListProducts() {
	suite.T().Run("Successful Product Listing", func(t *testing.T) {
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

		req := httptest.NewRequest(http.MethodGet, "/products?limit=10&offset=0", nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.Set("tenant_id", suite.tenantID)
		c.QueryParams().Set("limit", "10")
		c.QueryParams().Set("offset", "0")

		suite.mockProductService.On("List", context.Background(), suite.tenantID, 10, 0).Return(expectedProducts, nil)

		// Act
		err := suite.handlers.ListProducts(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
		suite.mockProductService.AssertExpectations(suite.T())
	})

	suite.T().Run("Invalid Query Parameters", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/products?limit=invalid&offset=-1", nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.Set("tenant_id", suite.tenantID)

		// Act
		err := suite.handlers.ListProducts(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	suite.T().Run("Service Error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.Set("tenant_id", suite.tenantID)

		suite.mockProductService.On("List", context.Background(), suite.tenantID, 100, 0).Return(nil, errors.New("database error"))

		// Act
		err := suite.handlers.ListProducts(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})
}

func (suite *ProductHandlersComprehensiveTestSuite) TestSearchProducts() {
	suite.T().Run("Successful Product Search", func(t *testing.T) {
		// Arrange
		categoryID := uuid.New()
		expectedProducts := []*models.Product{
			{
				ID:        uuid.New(),
				TenantID:  suite.tenantID,
				Name:      "Test Product",
				Quantity:  100,
				UnitPrice: 29.99,
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/products/search?query=test&category_id="+categoryID.String(), nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.Set("tenant_id", suite.tenantID)

		suite.mockProductService.On("Search", context.Background(), suite.tenantID, "test", &categoryID, 100, 0).Return(expectedProducts, nil)

		// Act
		err := suite.handlers.SearchProducts(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})

	suite.T().Run("Search Without Category", func(t *testing.T) {
		// Arrange
		expectedProducts := []*models.Product{
			{
				ID:        uuid.New(),
				TenantID:  suite.tenantID,
				Name:      "Test Product",
				Quantity:  100,
				UnitPrice: 29.99,
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/products/search?query=test", nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.Set("tenant_id", suite.tenantID)

		suite.mockProductService.On("Search", context.Background(), suite.tenantID, "test", (*uuid.UUID)(nil), 100, 0).Return(expectedProducts, nil)

		// Act
		err := suite.handlers.SearchProducts(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})
}

func (suite *ProductHandlersComprehensiveTestSuite) TestGetByBarcode() {
	suite.T().Run("Successful Barcode Lookup", func(t *testing.T) {
		// Arrange
		barcode := "123456789012"
		expectedProduct := &models.Product{
			ID:        uuid.New(),
			TenantID:  suite.tenantID,
			Name:      "Product by Barcode",
			Quantity:  100,
			UnitPrice: 29.99,
		}

		req := httptest.NewRequest(http.MethodGet, "/products/barcode/"+barcode, nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("barcode")
		c.SetParamValues(barcode)
		c.Set("tenant_id", suite.tenantID)

		suite.mockProductService.On("GetByBarcode", context.Background(), suite.tenantID, barcode).Return(expectedProduct, nil)

		// Act
		err := suite.handlers.GetByBarcode(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})

	suite.T().Run("Invalid Barcode Format", func(t *testing.T) {
		// Arrange
		barcode := strings.Repeat("1", 150) // Too long

		req := httptest.NewRequest(http.MethodGet, "/products/barcode/"+barcode, nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("barcode")
		c.SetParamValues(barcode)
		c.Set("tenant_id", suite.tenantID)

		// Act
		err := suite.handlers.GetByBarcode(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	suite.T().Run("Barcode Not Found", func(t *testing.T) {
		// Arrange
		barcode := "NONEXISTENT"

		req := httptest.NewRequest(http.MethodGet, "/products/barcode/"+barcode, nil)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetParamNames("barcode")
		c.SetParamValues(barcode)
		c.Set("tenant_id", suite.tenantID)

		suite.mockProductService.On("GetByBarcode", context.Background(), suite.tenantID, barcode).Return(nil, errors.New("product not found"))

		// Act
		err := suite.handlers.GetByBarcode(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		suite.mockProductService.AssertExpectations(suite.T())
	})
}

func (suite *ProductHandlersComprehensiveTestSuite) TestValidateUUID() {
	suite.T().Run("Valid UUID", func(t *testing.T) {
		// Arrange
		validUUID := "550e8400-e29b-41d4-a716-446655440000"
		expectedUUID := uuid.MustParse(validUUID)

		// Act
		result, err := suite.handlers.validateUUID(validUUID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedUUID, result)
	})

	suite.T().Run("Valid UUID with Whitespace", func(t *testing.T) {
		// Arrange
		validUUID := "  550e8400-e29b-41d4-a716-446655440000  "
		expectedUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

		// Act
		result, err := suite.handlers.validateUUID(validUUID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedUUID, result)
	})

	suite.T().Run("Empty UUID", func(t *testing.T) {
		// Act
		result, err := suite.handlers.validateUUID("")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, result)
	})

	suite.T().Run("Invalid UUID Format", func(t *testing.T) {
		// Test cases for invalid UUIDs
		testCases := []struct {
			name  string
			input string
		}{
			{"Too short", "550e8400-e29b-41d4-a716-44665544"},
			{"Too long", "550e8400-e29b-41d4-a716-4466554400000"},
			{"Missing hyphens", "550e8400e29b41d4a716446655440000"},
			{"Wrong hyphen positions", "550e8400e-29b-41d4-a716-446655440000"},
			{"Invalid characters", "550e8400-e29b-41d4-g716-446655440000"},
			{"Case insensitive valid", "550E8400-E29B-41D4-A716-446655440000"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := suite.handlers.validateUUID(tc.input)
				
				if tc.name == "Case insensitive valid" {
					// This one should pass
					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, result)
				} else {
					// These should fail
					assert.Error(t, err)
					assert.Equal(t, uuid.Nil, result)
				}
			})
		}
	})
}
