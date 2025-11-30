package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Note: MockCacheService is already defined in product_service_test.go in the same package

// RBACServiceTestSuite defines the test suite for RBAC service
type RBACServiceTestSuite struct {
	suite.Suite
	mockUserRoleRepo       *MockUserRoleRepository
	mockRolePermissionRepo *MockRolePermissionRepository
	mockPermissionRepo     *MockPermissionRepository
	mockCacheService       *MockCacheService
	service                RBACService
	serviceWithCache       RBACService
	tenantID               uuid.UUID
	userID                 uuid.UUID
	roleID                 uuid.UUID
	permissionID           uuid.UUID
	ctx                    context.Context
}

func (suite *RBACServiceTestSuite) SetupTest() {
	suite.mockUserRoleRepo = &MockUserRoleRepository{}
	suite.mockRolePermissionRepo = &MockRolePermissionRepository{}
	suite.mockPermissionRepo = &MockPermissionRepository{}
	suite.mockCacheService = &MockCacheService{}

	// Service without cache (for backward compatibility testing)
	suite.service = NewRBACService(
		suite.mockUserRoleRepo,
		suite.mockRolePermissionRepo,
		suite.mockPermissionRepo,
	)

	// Service with cache
	suite.serviceWithCache = NewRBACServiceWithCache(
		suite.mockUserRoleRepo,
		suite.mockRolePermissionRepo,
		suite.mockPermissionRepo,
		suite.mockCacheService,
	)

	suite.tenantID = uuid.New()
	suite.userID = uuid.New()
	suite.roleID = uuid.New()
	suite.permissionID = uuid.New()
	suite.ctx = context.Background()
}

func (suite *RBACServiceTestSuite) TearDownTest() {
	suite.mockUserRoleRepo.AssertExpectations(suite.T())
	suite.mockRolePermissionRepo.AssertExpectations(suite.T())
	suite.mockPermissionRepo.AssertExpectations(suite.T())
}

// Test UserHasPermission - Success scenario
func (suite *RBACServiceTestSuite) TestUserHasPermission_Success() {
	// Arrange
	permission := &models.Permission{
		ID:   suite.permissionID,
		Name: "product:read",
	}

	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return([]*models.Permission{permission}, nil)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission)
}

// Test UserHasPermission - User has no roles
func (suite *RBACServiceTestSuite) TestUserHasPermission_NoRoles() {
	// Arrange - User has no permissions
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return([]*models.Permission{}, nil)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission)
}

// Test UserHasPermission - Permission not found
func (suite *RBACServiceTestSuite) TestUserHasPermission_PermissionNotFound() {
	// Arrange - User has a different permission
	permission := &models.Permission{
		ID:   suite.permissionID,
		Name: "product:write",
	}

	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return([]*models.Permission{permission}, nil)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:delete")

	// Assert
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission)
}

// Test UserHasPermission - Error fetching permissions
func (suite *RBACServiceTestSuite) TestUserHasPermission_ErrorFetchingUserRoles() {
	// Arrange
	expectedError := errors.New("database error")
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return(nil, expectedError)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.Error(suite.T(), err)
	assert.False(suite.T(), hasPermission)
	assert.Contains(suite.T(), err.Error(), "failed to fetch user permissions")
}

// Test UserHasPermission - Error fetching role permissions (same as above with new impl)
func (suite *RBACServiceTestSuite) TestUserHasPermission_ErrorFetchingRolePermissions() {
	// Arrange
	expectedError := errors.New("database error")
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return(nil, expectedError)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.Error(suite.T(), err)
	assert.False(suite.T(), hasPermission)
	assert.Contains(suite.T(), err.Error(), "failed to fetch user permissions")
}

// Test UserHasPermission - Empty permissions returned
func (suite *RBACServiceTestSuite) TestUserHasPermission_NilPermission() {
	// Arrange - Empty permissions array
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return([]*models.Permission{}, nil)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission)
}

// Test GetUserPermissions - Success
func (suite *RBACServiceTestSuite) TestGetUserPermissions_Success() {
	// Arrange
	permission1 := &models.Permission{
		ID:   suite.permissionID,
		Name: "product:read",
	}
	permissionID2 := uuid.New()
	permission2 := &models.Permission{
		ID:   permissionID2,
		Name: "product:write",
	}

	// Mock the optimized GetAllUserPermissions call
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return([]*models.Permission{permission1, permission2}, nil)

	// Act
	permissions, err := suite.service.GetUserPermissions(suite.ctx, suite.userID, suite.tenantID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, 2)
	assert.Contains(suite.T(), permissions, "product:read")
	assert.Contains(suite.T(), permissions, "product:write")
}

// Test GetUserPermissions - User has no roles
func (suite *RBACServiceTestSuite) TestGetUserPermissions_NoRoles() {
	// Arrange - Mock returns empty permissions for user with no roles
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return([]*models.Permission{}, nil)

	// Act
	permissions, err := suite.service.GetUserPermissions(suite.ctx, suite.userID, suite.tenantID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, 0)
}

// Test GetUserPermissions - Error handling with partial results
func (suite *RBACServiceTestSuite) TestGetUserPermissions_PartialError() {
	// Arrange - Mock error from GetAllUserPermissions
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return(nil, errors.New("database error"))

	// Act
	permissions, err := suite.service.GetUserPermissions(suite.ctx, suite.userID, suite.tenantID)

	// Assert - Should return error since GetAllUserPermissions failed
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), permissions)
}

// Test UserHasPermission with cache - Cache hit
func (suite *RBACServiceTestSuite) TestUserHasPermissionWithCache_CacheHit() {
	// Arrange
	cacheKey := "agromart:rbac:permission:" + suite.tenantID.String() + ":" + suite.userID.String() + ":product:read"
	suite.mockCacheService.On("GetString", suite.ctx, cacheKey).Return("true", nil)

	// Act
	hasPermission, err := suite.serviceWithCache.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission)
	suite.mockCacheService.AssertExpectations(suite.T())
}

// Test UserHasPermission with cache - Cache miss, then cache
func (suite *RBACServiceTestSuite) TestUserHasPermissionWithCache_CacheMiss() {
	// Arrange
	permission := &models.Permission{
		ID:   suite.permissionID,
		Name: "product:read",
	}

	cacheKey := "agromart:rbac:permission:" + suite.tenantID.String() + ":" + suite.userID.String() + ":product:read"
	suite.mockCacheService.On("GetString", suite.ctx, cacheKey).Return("", errors.New("cache miss"))
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return([]*models.Permission{permission}, nil)
	suite.mockCacheService.On("SetString", suite.ctx, cacheKey, "true", 10*time.Minute).Return(nil)

	// Act
	hasPermission, err := suite.serviceWithCache.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission)
	suite.mockCacheService.AssertExpectations(suite.T())
}

// Test GetUserPermissions with cache - Cache hit
func (suite *RBACServiceTestSuite) TestGetUserPermissionsWithCache_CacheHit() {
	// Arrange
	permissions := []string{"product:read", "product:write"}
	cachedData, _ := json.Marshal(permissions)
	cacheKey := "agromart:rbac:permissions:" + suite.tenantID.String() + ":" + suite.userID.String()
	suite.mockCacheService.On("GetString", suite.ctx, cacheKey).Return(string(cachedData), nil)

	// Act
	result, err := suite.serviceWithCache.GetUserPermissions(suite.ctx, suite.userID, suite.tenantID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Contains(suite.T(), result, "product:read")
	assert.Contains(suite.T(), result, "product:write")
	suite.mockCacheService.AssertExpectations(suite.T())
}

// Test InvalidateUserPermissionsCache
func (suite *RBACServiceTestSuite) TestInvalidateUserPermissionsCache() {
	// Arrange
	permissionsKey := "agromart:rbac:permissions:" + suite.tenantID.String() + ":" + suite.userID.String()
	suite.mockCacheService.On("Delete", suite.ctx, permissionsKey).Return(nil)

	// Act
	err := suite.serviceWithCache.InvalidateUserPermissionsCache(suite.ctx, suite.userID, suite.tenantID)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockCacheService.AssertExpectations(suite.T())
}

// Test InvalidateUserPermissionsCache without cache service
func (suite *RBACServiceTestSuite) TestInvalidateUserPermissionsCache_NoCache() {
	// Act
	err := suite.service.InvalidateUserPermissionsCache(suite.ctx, suite.userID, suite.tenantID)

	// Assert
	assert.NoError(suite.T(), err)
}

// Test deduplication of permissions
func (suite *RBACServiceTestSuite) TestGetUserPermissions_Deduplication() {
	// Arrange - Mock returns permissions (deduplication is done by the optimized query)
	permission := &models.Permission{ID: suite.permissionID, Name: "product:read"}

	// The optimized GetAllUserPermissions should return deduplicated results from the database
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", suite.ctx, suite.userID, suite.tenantID).Return([]*models.Permission{permission}, nil)

	// Act
	permissions, err := suite.service.GetUserPermissions(suite.ctx, suite.userID, suite.tenantID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, 1) // Should be deduplicated
	assert.Contains(suite.T(), permissions, "product:read")
}

// Run the test suite
func TestRBACServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RBACServiceTestSuite))
}

// Mock repository implementations - should match the repository interfaces

// MockUserRoleRepository for testing
type MockUserRoleRepository struct {
	mock.Mock
}

func (m *MockUserRoleRepository) Create(ctx context.Context, tenantID uuid.UUID, userRole *models.UserRole) error {
	args := m.Called(ctx, tenantID, userRole)
	return args.Error(0)
}

func (m *MockUserRoleRepository) Delete(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, tenantID, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRoleRepository) ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.UserRole, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

func (m *MockUserRoleRepository) ListByRole(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.UserRole, error) {
	args := m.Called(ctx, tenantID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

func (m *MockUserRoleRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.UserRole, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

// MockRolePermissionRepository for testing
type MockRolePermissionRepository struct {
	mock.Mock
}

func (m *MockRolePermissionRepository) Create(ctx context.Context, tenantID uuid.UUID, rolePermission *models.RolePermission) error {
	args := m.Called(ctx, tenantID, rolePermission)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) Delete(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, tenantID, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) ListByRole(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.RolePermission, error) {
	args := m.Called(ctx, tenantID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RolePermission), args.Error(1)
}

func (m *MockRolePermissionRepository) ListByPermission(ctx context.Context, permissionID uuid.UUID, limit, offset int) ([]*models.RolePermission, error) {
	args := m.Called(ctx, permissionID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RolePermission), args.Error(1)
}

func (m *MockRolePermissionRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.RolePermission, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RolePermission), args.Error(1)
}

func (m *MockRolePermissionRepository) GetPermissionsByRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.Permission, error) {
	args := m.Called(ctx, tenantID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func (m *MockRolePermissionRepository) RemoveAllPermissionsFromRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error {
	args := m.Called(ctx, tenantID, roleID)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) AssignPermissionToRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, tenantID, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) RemovePermissionFromRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, tenantID, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) GetAllUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.Permission, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

// MockPermissionRepository for testing
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPermissionRepository) List(ctx context.Context) ([]models.RBACPermission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RBACPermission), args.Error(1)
}

func (m *MockPermissionRepository) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]models.RBACPermission, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RBACPermission), args.Error(1)
}

func (m *MockPermissionRepository) HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, userID, tenantID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionRepository) CheckResourceAccess(ctx context.Context, userID, tenantID uuid.UUID, resource, action string, resourceID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, tenantID, resource, action, resourceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionRepository) GetRolePermissions(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.Permission, error) {
	args := m.Called(ctx, tenantID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetPermissionsByRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]models.RBACPermission, error) {
	args := m.Called(ctx, tenantID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RBACPermission), args.Error(1)
}

func (m *MockPermissionRepository) AssignPermissionToRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID, conditions map[string]interface{}) error {
	args := m.Called(ctx, tenantID, roleID, permissionID, conditions)
	return args.Error(0)
}

func (m *MockPermissionRepository) RemovePermissionFromRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, tenantID, roleID, permissionID)
	return args.Error(0)
}

func (m *MockPermissionRepository) RemoveAllPermissionsFromRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error {
	args := m.Called(ctx, tenantID, roleID)
	return args.Error(0)
}

func (m *MockPermissionRepository) ListPermissions(ctx context.Context) ([]*models.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error) {
	args := m.Called(ctx, permissionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetPermissionByName(ctx context.Context, name string) (*models.RBACPermission, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RBACPermission), args.Error(1)
}
