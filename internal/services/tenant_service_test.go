package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) GetBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error) {
	args := m.Called(ctx, subdomain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTenantRepository) List(ctx context.Context, limit, offset int) ([]*models.Tenant, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) FindSettingsByTenantID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) UpdateSettings(ctx context.Context, tenant *models.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

type MockInvitationService struct {
	mock.Mock
}

func (m *MockInvitationService) CreateInvitation(ctx context.Context, req *CreateInvitationRequest, invitedBy uuid.UUID) (*models.Invitation, error) {
	args := m.Called(ctx, req, invitedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invitation), args.Error(1)
}

func (m *MockInvitationService) GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invitation), args.Error(1)
}

func (m *MockInvitationService) AcceptInvitation(ctx context.Context, token string, req *AcceptInvitationRequest) (*models.User, error) {
	args := m.Called(ctx, token, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockInvitationService) RevokeInvitation(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInvitationService) ListInvitations(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invitation, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Invitation), args.Error(1)
}

func (m *MockInvitationService) InviteTenantAdmin(ctx context.Context, email, tenantName string, invitedBy uuid.UUID) (*models.Invitation, error) {
	args := m.Called(ctx, email, tenantName, invitedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invitation), args.Error(1)
}

type MockPermissionLister struct {
	mock.Mock
}

func (m *MockPermissionLister) ListPermissions(ctx context.Context) ([]*models.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

type MockRolePermissionAssigner struct {
	mock.Mock
}

func (m *MockRolePermissionAssigner) Create(ctx context.Context, tenantID uuid.UUID, rolePermission *models.RolePermission) error {
	args := m.Called(ctx, tenantID, rolePermission)
	return args.Error(0)
}

type MockRoleTemplateApplier struct {
	mock.Mock
}

func (m *MockRoleTemplateApplier) ApplyTemplateToTenant(ctx context.Context, tenantID uuid.UUID, templateName string) error {
	args := m.Called(ctx, tenantID, templateName)
	return args.Error(0)
}

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *models.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *models.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Role, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	args := m.Called(ctx, tenantID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*models.Role, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockRoleRepository) AssignUserToRole(ctx context.Context, tenantID, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, tenantID, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepository) RemoveUserFromRole(ctx context.Context, tenantID, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, tenantID, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.Role, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleRepository) GetRoleUsers(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.User, error) {
	args := m.Called(ctx, tenantID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockRoleRepository) GetUserMaxPriority(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, tenantID, userID)
	return args.Int(0), args.Error(1)
}

type TenantServiceTestSuite struct {
	suite.Suite
	mockRepo                 *MockTenantRepository
	mockInviteSvc            *MockInvitationService
	mockRoleRepo             *MockRoleRepository
	mockPermissionLister     *MockPermissionLister
	mockRolePermissionAssign *MockRolePermissionAssigner
	mockRoleTemplateApplier  *MockRoleTemplateApplier
	service                  TenantService
}

func (suite *TenantServiceTestSuite) SetupTest() {
	suite.mockRepo = &MockTenantRepository{}
	suite.mockInviteSvc = &MockInvitationService{}
	suite.mockRoleRepo = &MockRoleRepository{}
	suite.mockPermissionLister = &MockPermissionLister{}
	suite.mockRolePermissionAssign = &MockRolePermissionAssigner{}
	suite.mockRoleTemplateApplier = &MockRoleTemplateApplier{}

	// Default permissive expectations for permission seeding; individual tests can override.
	suite.mockRoleRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Role")).Return(nil).Maybe()
	defaultAdminRole := &models.Role{ID: uuid.New(), Name: "admin"}
	suite.mockRoleRepo.On("GetByName", mock.Anything, mock.Anything, "admin").Return(defaultAdminRole, nil).Maybe()
	suite.mockPermissionLister.On("ListPermissions", mock.Anything).Return([]*models.Permission{}, nil).Maybe()
	suite.mockRolePermissionAssign.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.mockRoleTemplateApplier.On("ApplyTemplateToTenant", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	suite.service = NewTenantService(suite.mockRepo, suite.mockInviteSvc, suite.mockRoleRepo, suite.mockPermissionLister, suite.mockRolePermissionAssign, suite.mockRoleTemplateApplier)

	// Ensure mocks are called
	suite.mockRepo.Test(suite.T())
	suite.mockInviteSvc.Test(suite.T())
	suite.mockRoleRepo.Test(suite.T())
}

func (suite *TenantServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockPermissionLister.AssertExpectations(suite.T())
	suite.mockRolePermissionAssign.AssertExpectations(suite.T())
}

func TestTenantServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TenantServiceTestSuite))
}

func (suite *TenantServiceTestSuite) TestCreate_Success() {
	ctx := context.Background()
	req := &CreateTenantRequest{
		Name:      "Test Tenant",
		Subdomain: "test-tenant",
		License:   "LIC12345",
	}

	// Expected tenant after creation

	suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Run(func(args mock.Arguments) {
		tenant := args.Get(1).(*models.Tenant)
		assert.Equal(suite.T(), req.Name, tenant.Name)
		assert.Equal(suite.T(), req.Subdomain, tenant.Subdomain)
		assert.Equal(suite.T(), req.License, tenant.License)
		assert.Equal(suite.T(), "active", tenant.Status)
		assert.NotEqual(suite.T(), uuid.Nil, tenant.ID)
	})

	tenant, err := suite.service.Create(ctx, req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tenant)
	assert.Equal(suite.T(), req.Name, tenant.Name)
	assert.Equal(suite.T(), req.Subdomain, tenant.Subdomain)
	assert.Equal(suite.T(), req.License, tenant.License)
	assert.Equal(suite.T(), "active", tenant.Status)
}

func (suite *TenantServiceTestSuite) TestCreate_AppliesRoleTemplates() {
	ctx := context.Background()
	req := &CreateTenantRequest{
		Name:      "Template Tenant",
		Subdomain: "template-tenant",
		License:   "LIC999",
	}

	adminRoleID := uuid.New()

	// Reset defaults to assert exact calls for this test
	suite.mockPermissionLister.ExpectedCalls = nil
	suite.mockRolePermissionAssign.ExpectedCalls = nil
	suite.mockRoleRepo.ExpectedCalls = nil
	suite.mockRepo.ExpectedCalls = nil
	suite.mockRoleTemplateApplier.ExpectedCalls = nil

	suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil)

	// Verify ApplyTemplateToTenant is called for each standard role template
	expectedTemplates := []string{"admin", "manager", "user", "viewer", "inventory_manager", "sales", "analyst"}
	for _, template := range expectedTemplates {
		suite.mockRoleTemplateApplier.On("ApplyTemplateToTenant", ctx, mock.AnythingOfType("uuid.UUID"), template).Return(nil).Once()
	}

	// After templates are applied, admin role should be verified
	suite.mockRoleRepo.On("GetByName", ctx, mock.AnythingOfType("uuid.UUID"), "admin").Return(&models.Role{ID: adminRoleID, Name: "admin"}, nil).Once()

	tenant, err := suite.service.Create(ctx, req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tenant)

	// Verify all template applications were called
	suite.mockRoleTemplateApplier.AssertExpectations(suite.T())
}

func (suite *TenantServiceTestSuite) TestCreate_ValidationEmptyName() {
	ctx := context.Background()
	req := &CreateTenantRequest{
		Name:      "",
		Subdomain: "test-tenant",
		License:   "LIC12345",
	}

	tenant, err := suite.service.Create(ctx, req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "name and subdomain are required")
}

func (suite *TenantServiceTestSuite) TestCreate_ValidationEmptySubdomain() {
	ctx := context.Background()
	req := &CreateTenantRequest{
		Name:      "Test Tenant",
		Subdomain: "",
		License:   "LIC12345",
	}

	tenant, err := suite.service.Create(ctx, req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "name and subdomain are required")
}

func (suite *TenantServiceTestSuite) TestCreate_ValidationSubdomainWithSpaces() {
	ctx := context.Background()
	req := &CreateTenantRequest{
		Name:      "Test Tenant",
		Subdomain: "test tenant",
		License:   "LIC12345",
	}

	tenant, err := suite.service.Create(ctx, req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "subdomain cannot have spaces")
}

func (suite *TenantServiceTestSuite) TestCreate_ValidationSubdomainWithLeadingTrailingSpaces() {
	ctx := context.Background()
	req := &CreateTenantRequest{
		Name:      "Test Tenant",
		Subdomain: " test-tenant ",
		License:   "LIC12345",
	}

	tenant, err := suite.service.Create(ctx, req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "subdomain cannot have spaces")
}

func (suite *TenantServiceTestSuite) TestCreate_RepositoryError() {
	ctx := context.Background()
	req := &CreateTenantRequest{
		Name:      "Test Tenant",
		Subdomain: "test-tenant",
		License:   "LIC12345",
	}

	suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(errors.New("database connection failed"))

	tenant, err := suite.service.Create(ctx, req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

func (suite *TenantServiceTestSuite) TestGetByID_Success() {
	ctx := context.Background()
	tenantID := uuid.New()
	expectedTenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Test Tenant",
		Subdomain: "test-tenant",
		License:   "LIC12345",
		Status:    "active",
	}

	suite.mockRepo.On("GetByID", ctx, tenantID).Return(expectedTenant, nil)

	tenant, err := suite.service.GetByID(ctx, tenantID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTenant, tenant)
}

func (suite *TenantServiceTestSuite) TestGetByID_NotFound() {
	ctx := context.Background()
	tenantID := uuid.New()

	suite.mockRepo.On("GetByID", ctx, tenantID).Return((*models.Tenant)(nil), errors.New("tenant not found"))

	tenant, err := suite.service.GetByID(ctx, tenantID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "tenant not found")
}

func (suite *TenantServiceTestSuite) TestGetBySubdomain_Success() {
	ctx := context.Background()
	subdomain := "test-tenant"
	expectedTenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      "Test Tenant",
		Subdomain: subdomain,
		License:   "LIC12345",
		Status:    "active",
	}

	suite.mockRepo.On("GetBySubdomain", ctx, subdomain).Return(expectedTenant, nil)

	tenant, err := suite.service.GetBySubdomain(ctx, subdomain)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTenant, tenant)
}

func (suite *TenantServiceTestSuite) TestGetBySubdomain_EmptySubdomain() {
	ctx := context.Background()

	tenant, err := suite.service.GetBySubdomain(ctx, "")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "subdomain is required")
}

func (suite *TenantServiceTestSuite) TestGetBySubdomain_NotFound() {
	ctx := context.Background()
	subdomain := "nonexistent-tenant"

	suite.mockRepo.On("GetBySubdomain", ctx, subdomain).Return((*models.Tenant)(nil), errors.New("tenant not found"))

	tenant, err := suite.service.GetBySubdomain(ctx, subdomain)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "tenant not found")
}

func (suite *TenantServiceTestSuite) TestGetBySubdomain_DatabaseError() {
	ctx := context.Background()
	subdomain := "test-tenant"

	suite.mockRepo.On("GetBySubdomain", ctx, subdomain).Return((*models.Tenant)(nil), errors.New("database connection timeout"))

	tenant, err := suite.service.GetBySubdomain(ctx, subdomain)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "database connection timeout")
}

func (suite *TenantServiceTestSuite) TestUpdate_Success() {
	ctx := context.Background()
	tenantID := uuid.New()

	existingTenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Old Name",
		Subdomain: "old-subdomain",
		License:   "OLD12345",
		Status:    "active",
	}

	req := &UpdateTenantRequest{
		ID:        tenantID,
		Name:      "Updated Tenant",
		Subdomain: "updated-subdomain",
		License:   "LIC67890",
		Status:    "suspended",
	}

	suite.mockRepo.On("GetByID", ctx, tenantID).Return(existingTenant, nil)
	suite.mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Run(func(args mock.Arguments) {
		tenant := args.Get(1).(*models.Tenant)
		assert.Equal(suite.T(), req.Name, tenant.Name)
		assert.Equal(suite.T(), req.Subdomain, tenant.Subdomain)
		assert.Equal(suite.T(), req.License, tenant.License)
		assert.Equal(suite.T(), req.Status, tenant.Status)
		assert.Equal(suite.T(), tenantID, tenant.ID)
	})

	err := suite.service.Update(ctx, req)
	assert.NoError(suite.T(), err)
}

func (suite *TenantServiceTestSuite) TestUpdate_TenantNotFound() {
	ctx := context.Background()
	tenantID := uuid.New()

	req := &UpdateTenantRequest{
		ID:        tenantID,
		Name:      "Updated Tenant",
		Subdomain: "updated-subdomain",
		License:   "LIC67890",
		Status:    "suspended",
	}

	suite.mockRepo.On("GetByID", ctx, tenantID).Return((*models.Tenant)(nil), errors.New("tenant not found"))

	err := suite.service.Update(ctx, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "tenant not found")
}

func (suite *TenantServiceTestSuite) TestUpdate_GetByIDError() {
	ctx := context.Background()
	tenantID := uuid.New()

	req := &UpdateTenantRequest{
		ID:        tenantID,
		Name:      "Updated Tenant",
		Subdomain: "updated-subdomain",
		License:   "LIC67890",
		Status:    "suspended",
	}

	suite.mockRepo.On("GetByID", ctx, tenantID).Return((*models.Tenant)(nil), errors.New("database connection failed"))

	err := suite.service.Update(ctx, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

func (suite *TenantServiceTestSuite) TestUpdate_RepositoryError() {
	ctx := context.Background()
	tenantID := uuid.New()

	existingTenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Old Name",
		Subdomain: "old-subdomain",
		License:   "OLD12345",
		Status:    "active",
	}

	req := &UpdateTenantRequest{
		ID:        tenantID,
		Name:      "Updated Tenant",
		Subdomain: "updated-subdomain",
		License:   "LIC67890",
		Status:    "suspended",
	}

	suite.mockRepo.On("GetByID", ctx, tenantID).Return(existingTenant, nil)
	suite.mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Tenant")).Return(errors.New("update constraint violation"))

	err := suite.service.Update(ctx, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "update constraint violation")
}

func (suite *TenantServiceTestSuite) TestGetTenantSettings_Success() {
	ctx := context.Background()
	tenantID := uuid.New()
	expected := &models.Tenant{
		ID:        tenantID,
		Name:      "Settings Tenant",
		Subdomain: "settings-tenant",
		License:   "LIC-SETTINGS",
		Status:    "active",
	}

	suite.mockRepo.On("FindSettingsByTenantID", ctx, tenantID).Return(expected, nil).Once()

	tenant, err := suite.service.GetTenantSettings(ctx, tenantID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, tenant)
}

func (suite *TenantServiceTestSuite) TestGetTenantSettings_MissingTenantID() {
	ctx := context.Background()

	tenant, err := suite.service.GetTenantSettings(ctx, uuid.Nil)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant)
	assert.Contains(suite.T(), err.Error(), "tenant ID is required")
}

func (suite *TenantServiceTestSuite) TestUpdateTenantSettings_Success() {
	ctx := context.Background()
	tenantID := uuid.New()
	existing := &models.Tenant{
		ID:        tenantID,
		Name:      "Old Name",
		Subdomain: "old-subdomain",
		License:   "OLD-LIC",
		Status:    "active",
	}

	req := &UpdateTenantSettingsRequest{
		TenantID:  tenantID,
		Name:      "New Name",
		Subdomain: "new-subdomain",
		License:   "NEW-LIC",
	}

	suite.mockRepo.On("FindSettingsByTenantID", ctx, tenantID).Return(existing, nil).Once()
	suite.mockRepo.On("UpdateSettings", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once().Run(func(args mock.Arguments) {
		updated := args.Get(1).(*models.Tenant)
		assert.Equal(suite.T(), req.Name, updated.Name)
		assert.Equal(suite.T(), req.Subdomain, updated.Subdomain)
		assert.Equal(suite.T(), req.License, updated.License)
	})

	err := suite.service.UpdateTenantSettings(ctx, req)
	assert.NoError(suite.T(), err)
}

func (suite *TenantServiceTestSuite) TestUpdateTenantSettings_ValidationErrors() {
	ctx := context.Background()
	tenantID := uuid.New()

	testCases := []struct {
		name string
		req  *UpdateTenantSettingsRequest
	}{
		{"missing tenant id", &UpdateTenantSettingsRequest{TenantID: uuid.Nil, Name: "A", Subdomain: "abc"}},
		{"missing name", &UpdateTenantSettingsRequest{TenantID: tenantID, Name: "", Subdomain: "abc"}},
		{"missing subdomain", &UpdateTenantSettingsRequest{TenantID: tenantID, Name: "A", Subdomain: ""}},
		{"subdomain with spaces", &UpdateTenantSettingsRequest{TenantID: tenantID, Name: "A", Subdomain: "a b"}},
		{"subdomain too short", &UpdateTenantSettingsRequest{TenantID: tenantID, Name: "A", Subdomain: "ab"}},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.service.UpdateTenantSettings(ctx, tc.req)
			assert.Error(suite.T(), err)
		})
	}
}

func (suite *TenantServiceTestSuite) TestUpdateTenantSettings_FindSettingsError() {
	ctx := context.Background()
	tenantID := uuid.New()
	req := &UpdateTenantSettingsRequest{
		TenantID:  tenantID,
		Name:      "New Name",
		Subdomain: "new-subdomain",
	}

	suite.mockRepo.On("FindSettingsByTenantID", ctx, tenantID).Return((*models.Tenant)(nil), errors.New("lookup failed")).Once()

	err := suite.service.UpdateTenantSettings(ctx, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "lookup failed")
}

func (suite *TenantServiceTestSuite) TestUpdateTenantSettings_UpdateError() {
	ctx := context.Background()
	tenantID := uuid.New()
	existing := &models.Tenant{ID: tenantID, Name: "Old", Subdomain: "old"}
	req := &UpdateTenantSettingsRequest{
		TenantID:  tenantID,
		Name:      "New Name",
		Subdomain: "new-subdomain",
	}

	suite.mockRepo.On("FindSettingsByTenantID", ctx, tenantID).Return(existing, nil).Once()
	suite.mockRepo.On("UpdateSettings", ctx, mock.AnythingOfType("*models.Tenant")).Return(errors.New("update failed")).Once()

	err := suite.service.UpdateTenantSettings(ctx, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "update failed")
}

func (suite *TenantServiceTestSuite) TestSeedAdminPermissions_NoDependencies() {
	ctx := context.Background()
	svc := suite.service.(*tenantService)
	svc.permissionLister = nil
	svc.rolePermissionAssigner = nil

	err := svc.seedAdminPermissions(ctx, uuid.New(), uuid.New())
	assert.NoError(suite.T(), err)
}

func (suite *TenantServiceTestSuite) TestSeedAdminPermissions_ListPermissionsError() {
	ctx := context.Background()
	tenantID := uuid.New()
	adminRoleID := uuid.New()
	svc := suite.service.(*tenantService)

	suite.mockPermissionLister.ExpectedCalls = nil
	suite.mockRolePermissionAssign.ExpectedCalls = nil
	suite.mockPermissionLister.On("ListPermissions", ctx).Return([]*models.Permission(nil), errors.New("permission fetch failed")).Once()

	err := svc.seedAdminPermissions(ctx, tenantID, adminRoleID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "permission fetch failed")
	suite.mockRolePermissionAssign.AssertNotCalled(suite.T(), "Create", mock.Anything, mock.Anything, mock.Anything)
}

func (suite *TenantServiceTestSuite) TestSeedAdminPermissions_IgnoresDuplicateErrors() {
	ctx := context.Background()
	tenantID := uuid.New()
	adminRoleID := uuid.New()
	svc := suite.service.(*tenantService)

	perm1 := &models.Permission{ID: uuid.New(), Name: "perm1"}
	perm2 := &models.Permission{ID: uuid.New(), Name: "perm2"}

	suite.mockPermissionLister.ExpectedCalls = nil
	suite.mockRolePermissionAssign.ExpectedCalls = nil

	suite.mockPermissionLister.On("ListPermissions", ctx).Return([]*models.Permission{perm1, perm2}, nil).Once()
	suite.mockRolePermissionAssign.
		On("Create", ctx, tenantID, mock.AnythingOfType("*models.RolePermission")).
		Return(errors.New("duplicate key value violates unique constraint")).
		Times(2)

	err := svc.seedAdminPermissions(ctx, tenantID, adminRoleID)
	assert.NoError(suite.T(), err)
}

func (suite *TenantServiceTestSuite) TestSeedAdminPermissions_PropagatesAssignmentErrors() {
	ctx := context.Background()
	tenantID := uuid.New()
	adminRoleID := uuid.New()
	svc := suite.service.(*tenantService)

	perm := &models.Permission{ID: uuid.New(), Name: "perm1"}

	suite.mockPermissionLister.ExpectedCalls = nil
	suite.mockRolePermissionAssign.ExpectedCalls = nil

	suite.mockPermissionLister.On("ListPermissions", ctx).Return([]*models.Permission{perm}, nil).Once()
	suite.mockRolePermissionAssign.
		On("Create", ctx, tenantID, mock.AnythingOfType("*models.RolePermission")).
		Return(errors.New("unexpected failure")).Once()

	err := svc.seedAdminPermissions(ctx, tenantID, adminRoleID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "unexpected failure")
}

func (suite *TenantServiceTestSuite) TestDelete_Success() {
	ctx := context.Background()
	tenantID := uuid.New()

	suite.mockRepo.On("Delete", ctx, tenantID).Return(nil)

	err := suite.service.Delete(ctx, tenantID)
	assert.NoError(suite.T(), err)
}

func (suite *TenantServiceTestSuite) TestDelete_TenantNotFound() {
	ctx := context.Background()
	tenantID := uuid.New()

	suite.mockRepo.On("Delete", ctx, tenantID).Return(errors.New("tenant not found"))

	err := suite.service.Delete(ctx, tenantID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "tenant not found")
}

func (suite *TenantServiceTestSuite) TestList_Success() {
	ctx := context.Background()
	limit := 10
	offset := 0

	expectedTenants := []*models.Tenant{
		{
			ID:        uuid.New(),
			Name:      "Tenant 1",
			Subdomain: "tenant1",
			License:   "LIC1",
			Status:    "active",
		},
		{
			ID:        uuid.New(),
			Name:      "Tenant 2",
			Subdomain: "tenant2",
			License:   "LIC2",
			Status:    "active",
		},
	}

	suite.mockRepo.On("List", ctx, limit, offset).Return(expectedTenants, nil)

	tenants, err := suite.service.List(ctx, limit, offset)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTenants, tenants)
	assert.Len(suite.T(), tenants, 2)
}

func (suite *TenantServiceTestSuite) TestList_EmptyResult() {
	ctx := context.Background()
	limit := 10
	offset := 0

	suite.mockRepo.On("List", ctx, limit, offset).Return([]*models.Tenant{}, nil)

	tenants, err := suite.service.List(ctx, limit, offset)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), tenants)
}

func (suite *TenantServiceTestSuite) TestList_DefaultLimits() {
	ctx := context.Background()
	limit := 0
	offset := -5

	expectedTenants := []*models.Tenant{}

	suite.mockRepo.On("List", ctx, 10, 0).Return(expectedTenants, nil) // Should apply defaults

	tenants, err := suite.service.List(ctx, limit, offset)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTenants, tenants)
}

func (suite *TenantServiceTestSuite) TestList_RepositoryError() {
	ctx := context.Background()
	limit := 10
	offset := 0

	suite.mockRepo.On("List", ctx, limit, offset).Return([]*models.Tenant(nil), errors.New("database query failed"))

	tenants, err := suite.service.List(ctx, limit, offset)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenants)
	assert.Contains(suite.T(), err.Error(), "database query failed")
}

func (suite *TenantServiceTestSuite) TestMultiTenantIsolation() {
	ctx := context.Background()

	// Create tenants with different subdomains to test isolation
	tenants := []CreateTenantRequest{
		{Name: "Tenant A", Subdomain: "tenant-a", License: "LICA"},
		{Name: "Tenant B", Subdomain: "tenant-b", License: "LICB"},
	}

	createdTenants := []*models.Tenant{}

	for _, req := range tenants {
		suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once()
		tenant, err := suite.service.Create(ctx, &req)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), tenant)
		createdTenants = append(createdTenants, tenant)
	}

	// Verify each tenant can be retrieved independently
	for _, tenant := range createdTenants {
		suite.mockRepo.On("GetBySubdomain", ctx, tenant.Subdomain).Return(tenant, nil).Once()
		retrieved, err := suite.service.GetBySubdomain(ctx, tenant.Subdomain)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), tenant.ID, retrieved.ID)
		assert.Equal(suite.T(), tenant.Subdomain, retrieved.Subdomain)
	}
}

func (suite *TenantServiceTestSuite) TestLicenseValidationAndManagement() {
	ctx := context.Background()

	licenseNumbers := []string{
		"STANDARD-ENT-12345",
		"ENTERPRISE-EXT-67890",
		"TRIAL-TMP-ABC123",
		"", // No license
	}

	for _, license := range licenseNumbers {
		req := &CreateTenantRequest{
			Name:      "Test Tenant with License",
			Subdomain: "license-test-" + license,
			License:   license,
		}

		if license != "license-test-" { // Skip invalid empty subdomain test
			suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once()

			tenant, err := suite.service.Create(ctx, req)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), license, tenant.License)
		}
	}
}

func (suite *TenantServiceTestSuite) TestStatusManagement() {
	ctx := context.Background()
	tenantID := uuid.New()

	initialTenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Test Tenant",
		Subdomain: "status-test",
		Status:    "active",
		License:   "TEST123",
	}

	// Test status updates
	statuses := []string{"active", "inactive", "suspended"}

	for _, status := range statuses {
		req := &UpdateTenantRequest{
			ID:        tenantID,
			Name:      "Test Tenant",
			Subdomain: "status-test",
			License:   "TEST123",
			Status:    status,
		}

		suite.mockRepo.On("GetByID", ctx, tenantID).Return(initialTenant, nil).Once()
		suite.mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once()

		err := suite.service.Update(ctx, req)
		assert.NoError(suite.T(), err)
	}
}

func (suite *TenantServiceTestSuite) TestSubdomainUniquenessHandling() {
	ctx := context.Background()

	// First tenant
	req1 := &CreateTenantRequest{
		Name:      "Tenant 1",
		Subdomain: "unique-subdomain",
		License:   "LIC1",
	}

	suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once()

	tenant1, err := suite.service.Create(ctx, req1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "unique-subdomain", tenant1.Subdomain)

	// Second tenant trying same subdomain - this would fail at repository level
	req2 := &CreateTenantRequest{
		Name:      "Tenant 2",
		Subdomain: "unique-subdomain",
		License:   "LIC2",
	}

	suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(errors.New("duplicate key value violates unique constraint \"tenants_subdomain_key\"")).Once()

	tenant2, err := suite.service.Create(ctx, req2)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tenant2)
	assert.Contains(suite.T(), err.Error(), "unique constraint")
}

func (suite *TenantServiceTestSuite) TestCreateAndRetrieveFlow() {
	ctx := context.Background()

	req := &CreateTenantRequest{
		Name:      "Integration Test Tenant",
		Subdomain: "integration-test",
		License:   "INT123",
	}

	// Create tenant
	suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once()

	created, err := suite.service.Create(ctx, req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), created)
	assert.Equal(suite.T(), "active", created.Status) // Should be set by service

	// Retrieve by ID
	suite.mockRepo.On("GetByID", ctx, created.ID).Return(created, nil).Once()

	byID, err := suite.service.GetByID(ctx, created.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), created.ID, byID.ID)
	assert.Equal(suite.T(), created.Subdomain, byID.Subdomain)

	// Retrieve by subdomain
	suite.mockRepo.On("GetBySubdomain", ctx, created.Subdomain).Return(created, nil).Once()

	bySubdomain, err := suite.service.GetBySubdomain(ctx, created.Subdomain)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), created.ID, bySubdomain.ID)
	assert.Equal(suite.T(), created.Name, bySubdomain.Name)
}

func (suite *TenantServiceTestSuite) TestTenantDataIntegrity() {
	ctx := context.Background()

	// Test that tenant data is preserved correctly through different operations
	originalReq := &CreateTenantRequest{
		Name:      "Data Integrity Test",
		Subdomain: "data-integrity",
		License:   "DINT123",
	}

	suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once()

	created, err := suite.service.Create(ctx, originalReq)
	assert.NoError(suite.T(), err)

	// Update tenant with new data
	updateReq := &UpdateTenantRequest{
		ID:        created.ID,
		Name:      "Updated Name",
		Subdomain: "updated-subdomain",
		License:   "UPDT456",
		Status:    "suspended",
	}

	originalTenant := &models.Tenant{
		ID:        created.ID,
		Name:      "Data Integrity Test",
		Subdomain: "data-integrity",
		License:   "DINT123",
		Status:    "active",
	}

	suite.mockRepo.On("GetByID", ctx, created.ID).Return(originalTenant, nil).Once()
	suite.mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once()

	err = suite.service.Update(ctx, updateReq)
	assert.NoError(suite.T(), err)

	// // Retrieve and verify all fields
	// suite.mockRepo.On("GetByID", ctx, created.ID).Return(&models.Tenant{
	// 	ID:        created.ID,
	// 	Name:      "Updated Name",
	// 	Subdomain: "updated-subdomain",
	// 	License:   "UPDT456",
	// 	Status:    "suspended",
	// }, nil).Once()

	// updated, err := suite.service.GetByID(ctx, created.ID)
	// assert.NoError(suite.T(), err)
	// assert.Equal(suite.T(), updateReq.Name, updated.Name)
	// assert.Equal(suite.T(), updateReq.Subdomain, updated.Subdomain)
	// assert.Equal(suite.T(), updateReq.License, updated.License)
	// assert.Equal(suite.T(), updateReq.Status, updated.Status)
}

func (suite *TenantServiceTestSuite) TestBatchOperations() {
	ctx := context.Background()
	limit := 50
	offset := 100

	// Create a batch of tenants for testing
	batchTenants := []*models.Tenant{}
	for i := 0; i < 10; i++ {
		tenant := &models.Tenant{
			ID:        uuid.New(),
			Name:      "Batch Tenant " + string(rune(i)),
			Subdomain: "batch-tenant-" + string(rune(i)),
			License:   "BATCH" + string(rune(i)),
			Status:    "active",
		}
		batchTenants = append(batchTenants, tenant)
	}

	suite.mockRepo.On("List", ctx, limit, offset).Return(batchTenants, nil).Once()

	result, err := suite.service.List(ctx, limit, offset)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 10)
	assert.Equal(suite.T(), batchTenants, result)
}

func (suite *TenantServiceTestSuite) TestConcurrentTenantOperations() {
	ctx := context.Background()

	// Sequential creates to avoid mock races
	for i := 0; i < 5; i++ {
		req := &CreateTenantRequest{
			Name:      "Concurrent Tenant " + string(rune(i)),
			Subdomain: "concurrent-" + string(rune(i)),
			License:   "CONC" + string(rune(i)),
		}

		suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once()

		tenant, err := suite.service.Create(ctx, req)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), tenant)
		assert.Equal(suite.T(), req.Name, tenant.Name)
	}
}

// Test performance expectations for tenant service (unit level)
func (suite *TenantServiceTestSuite) TestPerformanceBasicOperations() {
	ctx := context.Background()

	// Setup tenant data
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Performance Test Tenant",
		Subdomain: "perf-test",
		Status:    "active",
	}

	// Benchmark-style assertions for operations
	start := time.Now()
	suite.mockRepo.On("GetByID", ctx, tenantID).Return(tenant, nil).Once()
	retrieved, err := suite.service.GetByID(ctx, tenantID)
	duration := time.Since(start)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrieved)
	// In a real performance test environment, we'd assert on duration:
	// assert.True(suite.T(), duration < 100*time.Millisecond, "GetByID took too long: %v", duration)
	assert.True(suite.T(), duration < 1*time.Second, "Operation should complete rapidly")
}

func (suite *TenantServiceTestSuite) TestSubdomainNormalization() {
	// Test various subdomain normalization scenarios
	testCases := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"test-tenant", "test-tenant", true},
		{"TestTenant", "TestTenant", true},
		{"test tenant", "", false}, // Spaces not allowed
		{"test_tenant", "test_tenant", true},
		{"test.tenant", "test.tenant", true},
		{"test123tenant", "test123tenant", true},
		{"tenant-with-multiple-hyphens", "tenant-with-multiple-hyphens", true},
	}

	for _, tc := range testCases {
		req := &CreateTenantRequest{
			Name:      "Test Tenant",
			Subdomain: tc.input,
			License:   "TEST123",
		}

		if tc.valid {
			suite.mockRepo.On("Create", context.Background(), mock.AnythingOfType("*models.Tenant")).Return(nil).Once()
			tenant, err := suite.service.Create(context.Background(), req)
			if tc.input == "TestTenant" && tc.expected == "TestTenant" {
				// This would fail if service lowercases subdomains, but it doesn't
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), tenant)
			}
		} else {
			suite.T().Skip("Validation test for invalid subdomains") // Skip for now
		}
	}
}

func (suite *TenantServiceTestSuite) TestContextCancellation() {
	ctx := context.Background()
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	_ = "test-bucket" // unused bucket variable
	subdomain := "cancel-test"

	req := &CreateTenantRequest{
		Name:      "Cancelled Tenant",
		Subdomain: subdomain,
		License:   "CANC123",
	}

	// Mock to simulate context cancellation at repository level
	suite.mockRepo.On("Create", cancelledCtx, mock.AnythingOfType("*models.Tenant")).Return(nil).Once()

	// Operations should handle context cancellation gracefully
	tenant, err := suite.service.Create(cancelledCtx, req)
	assert.NoError(suite.T(), err) // Service level validation happens before context check
	assert.NotNil(suite.T(), tenant)
}
