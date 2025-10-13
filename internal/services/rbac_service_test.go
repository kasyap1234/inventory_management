package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/testhelpers"

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
	mockCacheService       *testhelpers.MockCacheService
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
	suite.mockCacheService = &testhelpers.MockCacheService{}

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
	userRole := &models.UserRole{
		ID:     uuid.New(),
		UserID: suite.userID,
		RoleID: suite.roleID,
	}
	rolePermission := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       suite.roleID,
		PermissionID: suite.permissionID,
	}
	permission := &models.Permission{
		ID:   suite.permissionID,
		Name: "product:read",
	}

	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{userRole}, nil)
	suite.mockRolePermissionRepo.On("ListByRole", suite.ctx, suite.tenantID, suite.roleID).Return([]*models.RolePermission{rolePermission}, nil)
	suite.mockPermissionRepo.On("GetByID", suite.ctx, suite.permissionID).Return(permission, nil)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission)
}

// Test UserHasPermission - User has no roles
func (suite *RBACServiceTestSuite) TestUserHasPermission_NoRoles() {
	// Arrange
	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{}, nil)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission)
}

// Test UserHasPermission - Permission not found
func (suite *RBACServiceTestSuite) TestUserHasPermission_PermissionNotFound() {
	// Arrange
	userRole := &models.UserRole{
		ID:     uuid.New(),
		UserID: suite.userID,
		RoleID: suite.roleID,
	}
	rolePermission := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       suite.roleID,
		PermissionID: suite.permissionID,
	}
	permission := &models.Permission{
		ID:   suite.permissionID,
		Name: "product:write",
	}

	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{userRole}, nil)
	suite.mockRolePermissionRepo.On("ListByRole", suite.ctx, suite.tenantID, suite.roleID).Return([]*models.RolePermission{rolePermission}, nil)
	suite.mockPermissionRepo.On("GetByID", suite.ctx, suite.permissionID).Return(permission, nil)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:delete")

	// Assert
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission)
}

// Test UserHasPermission - Error fetching user roles
func (suite *RBACServiceTestSuite) TestUserHasPermission_ErrorFetchingUserRoles() {
	// Arrange
	expectedError := errors.New("database error")
	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return(nil, expectedError)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.Error(suite.T(), err)
	assert.False(suite.T(), hasPermission)
	assert.Contains(suite.T(), err.Error(), "failed to fetch user roles")
}

// Test UserHasPermission - Error fetching role permissions
func (suite *RBACServiceTestSuite) TestUserHasPermission_ErrorFetchingRolePermissions() {
	// Arrange
	userRole := &models.UserRole{
		ID:     uuid.New(),
		UserID: suite.userID,
		RoleID: suite.roleID,
	}
	expectedError := errors.New("database error")

	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{userRole}, nil)
	suite.mockRolePermissionRepo.On("ListByRole", suite.ctx, suite.tenantID, suite.roleID).Return(nil, expectedError)

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.Error(suite.T(), err)
	assert.False(suite.T(), hasPermission)
	assert.Contains(suite.T(), err.Error(), "failed to fetch role permissions")
}

// Test UserHasPermission - Nil permission returned
func (suite *RBACServiceTestSuite) TestUserHasPermission_NilPermission() {
	// Arrange
	userRole := &models.UserRole{
		ID:     uuid.New(),
		UserID: suite.userID,
		RoleID: suite.roleID,
	}
	rolePermission := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       suite.roleID,
		PermissionID: suite.permissionID,
	}

	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{userRole}, nil)
	suite.mockRolePermissionRepo.On("ListByRole", suite.ctx, suite.tenantID, suite.roleID).Return([]*models.RolePermission{rolePermission}, nil)
	suite.mockPermissionRepo.On("GetByID", suite.ctx, suite.permissionID).Return(nil, nil) // Nil permission

	// Act
	hasPermission, err := suite.service.UserHasPermission(suite.ctx, suite.userID, suite.tenantID, "product:read")

	// Assert
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission)
}

// Test GetUserPermissions - Success
func (suite *RBACServiceTestSuite) TestGetUserPermissions_Success() {
	// Arrange
	userRole := &models.UserRole{
		ID:     uuid.New(),
		UserID: suite.userID,
		RoleID: suite.roleID,
	}
	rolePermission1 := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       suite.roleID,
		PermissionID: suite.permissionID,
	}
	permissionID2 := uuid.New()
	rolePermission2 := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       suite.roleID,
		PermissionID: permissionID2,
	}
	permission1 := &models.Permission{
		ID:   suite.permissionID,
		Name: "product:read",
	}
	permission2 := &models.Permission{
		ID:   permissionID2,
		Name: "product:write",
	}

	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{userRole}, nil)
	suite.mockRolePermissionRepo.On("ListByRole", suite.ctx, suite.tenantID, suite.roleID).Return([]*models.RolePermission{rolePermission1, rolePermission2}, nil)
	suite.mockPermissionRepo.On("GetByID", suite.ctx, suite.permissionID).Return(permission1, nil)
	suite.mockPermissionRepo.On("GetByID", suite.ctx, permissionID2).Return(permission2, nil)

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
	// Arrange
	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{}, nil)

	// Act
	permissions, err := suite.service.GetUserPermissions(suite.ctx, suite.userID, suite.tenantID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, 0)
}

// Test GetUserPermissions - Error handling with partial results
func (suite *RBACServiceTestSuite) TestGetUserPermissions_PartialError() {
	// Arrange
	userRole := &models.UserRole{
		ID:     uuid.New(),
		UserID: suite.userID,
		RoleID: suite.roleID,
	}
	rolePermission1 := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       suite.roleID,
		PermissionID: suite.permissionID,
	}
	permissionID2 := uuid.New()
	rolePermission2 := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       suite.roleID,
		PermissionID: permissionID2,
	}
	permission1 := &models.Permission{
		ID:   suite.permissionID,
		Name: "product:read",
	}

	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{userRole}, nil)
	suite.mockRolePermissionRepo.On("ListByRole", suite.ctx, suite.tenantID, suite.roleID).Return([]*models.RolePermission{rolePermission1, rolePermission2}, nil)
	suite.mockPermissionRepo.On("GetByID", suite.ctx, suite.permissionID).Return(permission1, nil)
	suite.mockPermissionRepo.On("GetByID", suite.ctx, permissionID2).Return(nil, errors.New("permission not found"))

	// Act
	permissions, err := suite.service.GetUserPermissions(suite.ctx, suite.userID, suite.tenantID)

	// Assert
	assert.NoError(suite.T(), err) // Should not error out, just log and continue
	assert.Len(suite.T(), permissions, 1)
	assert.Contains(suite.T(), permissions, "product:read")
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
	userRole := &models.UserRole{
		ID:     uuid.New(),
		UserID: suite.userID,
		RoleID: suite.roleID,
	}
	rolePermission := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       suite.roleID,
		PermissionID: suite.permissionID,
	}
	permission := &models.Permission{
		ID:   suite.permissionID,
		Name: "product:read",
	}

	cacheKey := "agromart:rbac:permission:" + suite.tenantID.String() + ":" + suite.userID.String() + ":product:read"
	suite.mockCacheService.On("GetString", suite.ctx, cacheKey).Return("", errors.New("cache miss"))
	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{userRole}, nil)
	suite.mockRolePermissionRepo.On("ListByRole", suite.ctx, suite.tenantID, suite.roleID).Return([]*models.RolePermission{rolePermission}, nil)
	suite.mockPermissionRepo.On("GetByID", suite.ctx, suite.permissionID).Return(permission, nil)
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
	// Arrange - User has two roles with the same permission
	role1 := uuid.New()
	role2 := uuid.New()
	userRole1 := &models.UserRole{ID: uuid.New(), UserID: suite.userID, RoleID: role1}
	userRole2 := &models.UserRole{ID: uuid.New(), UserID: suite.userID, RoleID: role2}
	rolePermission1 := &models.RolePermission{ID: uuid.New(), RoleID: role1, PermissionID: suite.permissionID}
	rolePermission2 := &models.RolePermission{ID: uuid.New(), RoleID: role2, PermissionID: suite.permissionID}
	permission := &models.Permission{ID: suite.permissionID, Name: "product:read"}

	suite.mockUserRoleRepo.On("ListByUser", suite.ctx, suite.tenantID, suite.userID).Return([]*models.UserRole{userRole1, userRole2}, nil)
	suite.mockRolePermissionRepo.On("ListByRole", suite.ctx, suite.tenantID, role1).Return([]*models.RolePermission{rolePermission1}, nil)
	suite.mockRolePermissionRepo.On("ListByRole", suite.ctx, suite.tenantID, role2).Return([]*models.RolePermission{rolePermission2}, nil)
	suite.mockPermissionRepo.On("GetByID", suite.ctx, suite.permissionID).Return(permission, nil).Twice()

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

func (m *MockPermissionRepository) List(ctx context.Context, limit, offset int) ([]*models.Permission, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}