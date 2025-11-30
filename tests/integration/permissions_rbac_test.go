package integration

import (
	"context"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// RBAC Integration Test Suite
type RBACIntegrationTestSuite struct {
	suite.Suite
	mockUserRoleRepo       *MockUserRoleRepository
	mockRolePermissionRepo *MockRolePermissionRepository
	mockPermissionRepo     *MockPermissionRepository
	rbacService            services.RBACService

	tenantID uuid.UUID
	userID   uuid.UUID

	// Test fixtures
	testRole           *models.Role
	testPermission     *models.Permission
	testUserRole       *models.UserRole
	testRolePermission *models.RolePermission
}

// SetupTest initializes the RBAC integration test suite
func (suite *RBACIntegrationTestSuite) SetupTest() {
	suite.mockUserRoleRepo = &MockUserRoleRepository{}
	suite.mockRolePermissionRepo = &MockRolePermissionRepository{}
	suite.mockPermissionRepo = &MockPermissionRepository{}

	suite.rbacService = services.NewRBACService(
		suite.mockUserRoleRepo,
		suite.mockRolePermissionRepo,
		suite.mockPermissionRepo,
	)

	suite.tenantID = uuid.New()
	suite.userID = uuid.New()
	suite.initializeTestFixtures()
}

// initializeTestFixtures sets up test data
func (suite *RBACIntegrationTestSuite) initializeTestFixtures() {
	suite.testRole = &models.Role{
		ID:          uuid.New(),
		TenantID:    suite.tenantID,
		Name:        "TestRole",
		Description: stringPtr("Test role for RBAC integration tests"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	suite.testPermission = &models.Permission{
		ID:          uuid.New(),
		Name:        "test:permission",
		Description: stringPtr("Test permission for RBAC integration tests"),
		CreatedAt:   time.Now(),
	}

	suite.testUserRole = &models.UserRole{
		ID:        uuid.New(),
		UserID:    suite.userID,
		RoleID:    suite.testRole.ID,
		CreatedAt: time.Now(),
	}

	suite.testRolePermission = &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       suite.testRole.ID,
		PermissionID: suite.testPermission.ID,
		CreatedAt:    time.Now(),
	}
}

// TearDownTest cleans up test resources
func (suite *RBACIntegrationTestSuite) TearDownTest() {
	suite.mockUserRoleRepo.AssertExpectations(suite.T())
	suite.mockRolePermissionRepo.AssertExpectations(suite.T())
	suite.mockPermissionRepo.AssertExpectations(suite.T())
}

// stringPtr creates a string pointer for convenience
func stringPtr(s string) *string {
	return &s
}

// intPtr creates an int pointer for convenience
func intPtr(i int) *int {
	return &i
}

// TestEndToEndPermissionWorkflow_UserHasPermission_Granted tests successful permission grant
func (suite *RBACIntegrationTestSuite) TestEndToEndPermissionWorkflow_UserHasPermission_Granted() {
	ctx := context.Background()

	// Setup: User has role with required permission (using optimized single-query approach)
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).
		Return([]*models.Permission{suite.testPermission}, nil).Once()

	hasPermission, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, suite.testPermission.Name)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission, "User should have the required permission")
}

// TestEndToEndPermissionWorkflow_UserHasPermission_Denied_NoRole tests permission denial when user has no roles
func (suite *RBACIntegrationTestSuite) TestEndToEndPermissionWorkflow_UserHasPermission_Denied_NoRole() {
	ctx := context.Background()

	// Setup: User has no roles (empty permissions returned)
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).
		Return([]*models.Permission{}, nil).Once()

	hasPermission, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, "any:permission")

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission, "User without roles should not have any permissions")
}

// TestEndToEndPermissionWorkflow_UserHasPermission_Denied_RoleNoPermission tests permission denial when role has no permissions
func (suite *RBACIntegrationTestSuite) TestEndToEndPermissionWorkflow_UserHasPermission_Denied_RoleNoPermission() {
	ctx := context.Background()

	// Setup: User has role but role has no permissions (empty permissions returned)
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).
		Return([]*models.Permission{}, nil).Once()

	hasPermission, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, "missing:permission")

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission, "User with role but no permissions should be denied access")
}

// TestMultiTenantPermissionIsolation tests that permissions are properly isolated between tenants
func (suite *RBACIntegrationTestSuite) TestMultiTenantPermissionIsolation() {
	ctx := context.Background()

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()

	// Setup tenant 1: User has permission
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, tenantID1).
		Return([]*models.Permission{suite.testPermission}, nil).Once()

	// Setup tenant 2: Same user but different tenant has no permissions
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, tenantID2).
		Return([]*models.Permission{}, nil).Once()

	// Test: User has permission in tenant 1 but not in tenant 2
	hasPermission1, err := suite.rbacService.UserHasPermission(ctx, suite.userID, tenantID1, suite.testPermission.Name)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission1, "User should have permission in tenant 1")

	hasPermission2, err := suite.rbacService.UserHasPermission(ctx, suite.userID, tenantID2, suite.testPermission.Name)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission2, "User should not have permission in tenant 2")
}

// TestGetUserPermissions tests permission retrieval functionality
func (suite *RBACIntegrationTestSuite) TestGetUserPermissions_Success() {
	ctx := context.Background()

	// Using optimized single-query approach
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).
		Return([]*models.Permission{suite.testPermission}, nil).Once()

	permissions, err := suite.rbacService.GetUserPermissions(ctx, suite.userID, suite.tenantID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, 1)
	assert.Equal(suite.T(), suite.testPermission.Name, permissions[0])
}

// TestGetUserPermissions_NoRoles tests user with no roles
func (suite *RBACIntegrationTestSuite) TestGetUserPermissions_NoRoles() {
	ctx := context.Background()

	// User has no roles, returns empty permissions
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).
		Return([]*models.Permission{}, nil).Once()

	permissions, err := suite.rbacService.GetUserPermissions(ctx, suite.userID, suite.tenantID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, 0, "User with no roles should have no permissions")
}

// TestMultipleRolesMultiplePermissions tests user with multiple roles and permissions
func (suite *RBACIntegrationTestSuite) TestMultipleRolesMultiplePermissions() {
	ctx := context.Background()

	// Create additional permission for testing
	permission2 := &models.Permission{
		ID:          uuid.New(),
		Name:        "additional:permission",
		Description: stringPtr("Additional test permission"),
	}

	// User has two roles with different permissions - the optimized query returns all at once
	allPermissions := []*models.Permission{suite.testPermission, permission2}

	// Mock setup using optimized GetAllUserPermissions
	// 3 calls: GetUserPermissions + 2x UserHasPermission
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).
		Return(allPermissions, nil).Times(3)

	permissions, err := suite.rbacService.GetUserPermissions(ctx, suite.userID, suite.tenantID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, 2, "User with multiple roles should have multiple permissions")

	// Test individual permission checks
	hasPerm1, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, suite.testPermission.Name)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPerm1, "User should have first permission")

	hasPerm2, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, permission2.Name)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPerm2, "User should have second permission")
}

// TestRepositoryErrorHandling tests RBAC service error handling
func (suite *RBACIntegrationTestSuite) TestRepositoryErrorHandling() {
	ctx := context.Background()

	// Test error in GetAllUserPermissions
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).
		Return([]*models.Permission(nil), assert.AnError).Once()

	hasPermission, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, "any:permission")
	assert.Error(suite.T(), err)
	assert.False(suite.T(), hasPermission, "Error should result in permission denial")
}

// TestPermissionsDeduplication tests that duplicate permissions are correctly handled
func (suite *RBACIntegrationTestSuite) TestPermissionsDeduplication() {
	ctx := context.Background()

	// With the optimized query, deduplication happens at the database level via DISTINCT
	// The same permission returned once even if user has multiple roles with the same permission
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).
		Return([]*models.Permission{suite.testPermission}, nil).Once()

	permissions, err := suite.rbacService.GetUserPermissions(ctx, suite.userID, suite.tenantID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, 1, "Duplicate permissions should be deduplicated")
	assert.Equal(suite.T(), suite.testPermission.Name, permissions[0])
}

// TestRoleHierarchyInheritance tests permission inheritance through multiple roles
func (suite *RBACIntegrationTestSuite) TestRoleHierarchyInheritance() {
	ctx := context.Background()

	// Employee permission
	employeePerm := &models.Permission{
		ID:          suite.testPermission.ID,
		Name:        "employee:access",
		Description: stringPtr("Basic employee access"),
	}

	// Manager permission
	managerPerm := &models.Permission{
		ID:          uuid.New(),
		Name:        "manager:approve",
		Description: stringPtr("Manager approval permission"),
	}

	// Manager has both employee and manager permissions via optimized query
	allPermissions := []*models.Permission{employeePerm, managerPerm}

	// 3 calls: 2x UserHasPermission + 1x GetUserPermissions
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).
		Return(allPermissions, nil).Times(3)

	// Test that manager has both employee and manager permissions
	hasEmployeePerm, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, employeePerm.Name)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasEmployeePerm, "Manager should have employee permissions")

	hasManagerPerm, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, managerPerm.Name)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasManagerPerm, "Manager should have manager permissions")

	// Test permission list includes both permissions
	permissions, err := suite.rbacService.GetUserPermissions(ctx, suite.userID, suite.tenantID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, 2, "Manager should have both employee and manager permissions")
}

// TestAccessControlPattern_ComprehensiveCRUDOperations tests all CRUD operation permissions
func (suite *RBACIntegrationTestSuite) TestAccessControlPattern_ComprehensiveCRUDOperations() {
	ctx := context.Background()

	// Create permissions for CRUD operations on products
	productPermissions := []string{"products:create", "products:read", "products:update", "products:delete"}
	userPermissions := make([]*models.Permission, len(productPermissions))

	for i, permName := range productPermissions {
		permObj := &models.Permission{
			ID:          uuid.New(),
			Name:        permName,
			Description: stringPtr("Product " + permName + " permission"),
			CreatedAt:   time.Now(),
		}
		userPermissions[i] = permObj
	}

	// Mock GetAllUserPermissions - 6 calls total: 1 for GetUserPermissions + 4 for valid perms + 1 for invalid perm
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", ctx, suite.userID, suite.tenantID).Return(userPermissions, nil).Times(6)

	// Test that user has all CRUD permissions
	permissions, err := suite.rbacService.GetUserPermissions(ctx, suite.userID, suite.tenantID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), permissions, len(productPermissions), "User should have all CRUD permissions")

	// Test individual permission checks
	for _, permName := range productPermissions {
		hasPermission, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, permName)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), hasPermission, "User should have %s permission", permName)
	}

	// Test assertion that user does NOT have a permission they shouldn't have
	hasWrongPermission, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, "products:admin")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasWrongPermission, "User should not have unauthorized permission")
}

// TestConcurrentPermissionAccess tests that RBAC handles concurrent access properly
func (suite *RBACIntegrationTestSuite) TestConcurrentPermissionAccess() {
	ctx := context.Background()

	// Setup expectations using the optimized GetAllUserPermissions that work for concurrent calls
	suite.mockRolePermissionRepo.On("GetAllUserPermissions", mock.Anything, suite.userID, suite.tenantID).
		Return([]*models.Permission{suite.testPermission}, nil).
		Times(10)

	// Run multiple concurrent permission checks
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			hasPermission, err := suite.rbacService.UserHasPermission(ctx, suite.userID, suite.tenantID, suite.testPermission.Name)
			assert.NoError(suite.T(), err)
			assert.True(suite.T(), hasPermission)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestModule function to run the RBAC integration test suite
func TestRBACIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RBACIntegrationTestSuite))
}

// Mock User Role Repository
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
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

func (m *MockUserRoleRepository) ListByRole(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.UserRole, error) {
	args := m.Called(ctx, tenantID, roleID)
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

func (m *MockUserRoleRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.UserRole, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

// Mock Role Permission Repository
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
	return args.Get(0).([]*models.RolePermission), args.Error(1)
}

func (m *MockRolePermissionRepository) ListByPermission(ctx context.Context, permissionID uuid.UUID, limit, offset int) ([]*models.RolePermission, error) {
	args := m.Called(ctx, permissionID, limit, offset)
	return args.Get(0).([]*models.RolePermission), args.Error(1)
}

func (m *MockRolePermissionRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.RolePermission, error) {
	args := m.Called(ctx, tenantID, limit, offset)
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

// Mock Permission Repository
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) GetPermissionByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
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

// Note: GetPermissionByID is already defined at line 577, removed duplicate

func (m *MockPermissionRepository) GetPermissionByName(ctx context.Context, name string) (*models.RBACPermission, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RBACPermission), args.Error(1)
}
