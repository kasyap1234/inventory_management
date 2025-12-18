package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// CreateUserHandlerSimulation simulates the CreateUser handler's permission flow
// This tests the actual permission checking logic that happens INSIDE the handler
// (not just in the middleware wrapper)
type CreateUserRBACFlowTestSuite struct {
	suite.Suite

	mockRBACService *MockRBACServiceForHandlers
	rbacMiddleware  *middleware.RBACMiddleware

	tenantID uuid.UUID
	userID   uuid.UUID
	echo     *echo.Echo
}

func (suite *CreateUserRBACFlowTestSuite) SetupTest() {
	suite.tenantID = uuid.New()
	suite.userID = uuid.New()
	suite.echo = echo.New()

	suite.mockRBACService = &MockRBACServiceForHandlers{}
	suite.rbacMiddleware = middleware.NewRBACMiddleware(suite.mockRBACService)
}

func (suite *CreateUserRBACFlowTestSuite) TearDownTest() {
	suite.mockRBACService.AssertExpectations(suite.T())
}

// simulateCreateUserHandler simulates the permission flow of CreateUser handler
// This replicates the actual permission checking logic from user_handlers.go
func (suite *CreateUserRBACFlowTestSuite) simulateCreateUserHandler(c echo.Context, hasRoleID bool, hasPermissions bool) error {
	ctx := c.Request().Context()
	userID, _ := common.GetUserIDFromContext(ctx)
	tenantID, _ := common.GetTenantIDFromContext(ctx)

	// Handle Permission/Role Assignment - this is what happens INSIDE the handler
	if hasPermissions {
		// Check role.create permission
		hasPerm, err := suite.mockRBACService.UserHasPermission(ctx, userID, tenantID, "role.create")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error checking permission")
		}
		if !hasPerm {
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to create custom roles")
		}

		// Check role.manage_permissions
		hasPerm, err = suite.mockRBACService.UserHasPermission(ctx, userID, tenantID, "role.manage_permissions")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error checking permission")
		}
		if !hasPerm {
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to manage role permissions")
		}

		// Check user.manage_roles for assigning the created role
		hasPerm, err = suite.mockRBACService.UserHasPermission(ctx, userID, tenantID, "user.manage_roles")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error checking permission")
		}
		if !hasPerm {
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to assign roles to users")
		}
	} else if hasRoleID {
		// Assign specific role - requires user.manage_roles permission
		hasPerm, err := suite.mockRBACService.UserHasPermission(ctx, userID, tenantID, "user.manage_roles")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error checking permission")
		}
		if !hasPerm {
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions to assign roles to users")
		}
	} else {
		// Assign default 'user' role - requires user.manage_roles permission
		hasPerm, err := suite.mockRBACService.UserHasPermission(ctx, userID, tenantID, "user.manage_roles")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error checking permission")
		}
		if !hasPerm {
			// If user doesn't have manage_roles permission, skip role assignment
			return c.JSON(http.StatusCreated, map[string]interface{}{
				"message": "User created without role assignment. Assign role manually if needed.",
			})
		}
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User created with role assignment.",
	})
}

// TestFullFlow_CreateUser_WithRoleID_DeniedWithoutManageRoles tests the complete flow:
// 1. Middleware checks user.create (passes)
// 2. Handler checks user.manage_roles for role assignment (fails)
func (suite *CreateUserRBACFlowTestSuite) TestFullFlow_CreateUser_WithRoleID_DeniedWithoutManageRoles() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.UserIDKey, suite.userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, suite.tenantID)

	requestBody := map[string]interface{}{
		"email":      "newuser@example.com",
		"first_name": "New",
		"last_name":  "User",
		"role_id":    uuid.New().String(),
	}

	// Mock: user.create passes (middleware)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.create").
		Return(true, nil).Once()

	// Mock: user.manage_roles fails (handler)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.manage_roles").
		Return(false, nil).Once()

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)

	// Wrap with middleware, then call simulated handler
	handler := suite.rbacMiddleware.RequirePermission("user.create")(func(c echo.Context) error {
		return suite.simulateCreateUserHandler(c, true, false) // hasRoleID=true, hasPermissions=false
	})

	err := handler(c)

	// The handler should return 403 Forbidden for role assignment
	assert.Error(suite.T(), err)
	he, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusForbidden, he.Code)
	assert.Contains(suite.T(), he.Message, "Insufficient permissions to assign roles")
}

// TestFullFlow_CreateUser_WithRoleID_AllowedWithManageRoles tests successful role assignment
func (suite *CreateUserRBACFlowTestSuite) TestFullFlow_CreateUser_WithRoleID_AllowedWithManageRoles() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.UserIDKey, suite.userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, suite.tenantID)

	requestBody := map[string]interface{}{
		"email":      "newuser@example.com",
		"first_name": "New",
		"last_name":  "User",
		"role_id":    uuid.New().String(),
	}

	// Mock: user.create passes (middleware)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.create").
		Return(true, nil).Once()

	// Mock: user.manage_roles passes (handler)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.manage_roles").
		Return(true, nil).Once()

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)

	handler := suite.rbacMiddleware.RequirePermission("user.create")(func(c echo.Context) error {
		return suite.simulateCreateUserHandler(c, true, false)
	})

	err := handler(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)
}

// TestFullFlow_CreateUser_WithPermissions_DeniedWithoutRoleCreate tests custom role creation denied
func (suite *CreateUserRBACFlowTestSuite) TestFullFlow_CreateUser_WithPermissions_DeniedWithoutRoleCreate() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.UserIDKey, suite.userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, suite.tenantID)

	requestBody := map[string]interface{}{
		"email":       "newuser@example.com",
		"first_name":  "New",
		"last_name":   "User",
		"permissions": []string{"product.read", "product.create"},
	}

	// Mock: user.create passes (middleware)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.create").
		Return(true, nil).Once()

	// Mock: role.create fails (handler)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "role.create").
		Return(false, nil).Once()

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)

	handler := suite.rbacMiddleware.RequirePermission("user.create")(func(c echo.Context) error {
		return suite.simulateCreateUserHandler(c, false, true) // hasRoleID=false, hasPermissions=true
	})

	err := handler(c)

	assert.Error(suite.T(), err)
	he, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusForbidden, he.Code)
	assert.Contains(suite.T(), he.Message, "Insufficient permissions to create custom roles")
}

// TestFullFlow_CreateUser_WithPermissions_AllAllowed tests full permission flow
func (suite *CreateUserRBACFlowTestSuite) TestFullFlow_CreateUser_WithPermissions_AllAllowed() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.UserIDKey, suite.userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, suite.tenantID)

	requestBody := map[string]interface{}{
		"email":       "newuser@example.com",
		"first_name":  "New",
		"last_name":   "User",
		"permissions": []string{"product.read", "product.create"},
	}

	// Mock: user.create passes (middleware)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.create").
		Return(true, nil).Once()

	// Mock: role.create passes (handler)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "role.create").
		Return(true, nil).Once()

	// Mock: role.manage_permissions passes (handler)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "role.manage_permissions").
		Return(true, nil).Once()

	// Mock: user.manage_roles passes (handler)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.manage_roles").
		Return(true, nil).Once()

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)

	handler := suite.rbacMiddleware.RequirePermission("user.create")(func(c echo.Context) error {
		return suite.simulateCreateUserHandler(c, false, true)
	})

	err := handler(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)
}

// TestFullFlow_CreateUser_NoRole_CreatedWithoutRoleAssignment tests graceful fallback
func (suite *CreateUserRBACFlowTestSuite) TestFullFlow_CreateUser_NoRole_CreatedWithoutRoleAssignment() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.UserIDKey, suite.userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, suite.tenantID)

	requestBody := map[string]interface{}{
		"email":      "newuser@example.com",
		"first_name": "New",
		"last_name":  "User",
	}

	// Mock: user.create passes (middleware)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.create").
		Return(true, nil).Once()

	// Mock: user.manage_roles fails (handler) - but this is OK for no-role case
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.manage_roles").
		Return(false, nil).Once()

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)

	handler := suite.rbacMiddleware.RequirePermission("user.create")(func(c echo.Context) error {
		return suite.simulateCreateUserHandler(c, false, false)
	})

	err := handler(c)

	// Should succeed but with message about no role assignment
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Contains(suite.T(), response["message"], "without role assignment")
}

// TestFullFlow_CreateUser_DeniedWithoutUserCreate tests middleware denial
func (suite *CreateUserRBACFlowTestSuite) TestFullFlow_CreateUser_DeniedWithoutUserCreate() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.UserIDKey, suite.userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, suite.tenantID)

	requestBody := map[string]interface{}{
		"email":      "newuser@example.com",
		"first_name": "New",
		"last_name":  "User",
	}

	// Mock: user.create fails (middleware)
	suite.mockRBACService.On("UserHasPermission", mock.Anything, suite.userID, suite.tenantID, "user.create").
		Return(false, nil).Once()

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)

	handler := suite.rbacMiddleware.RequirePermission("user.create")(func(c echo.Context) error {
		return suite.simulateCreateUserHandler(c, false, false)
	})

	err := handler(c)

	assert.Error(suite.T(), err)
	he, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusForbidden, he.Code)
}

func TestCreateUserRBACFlowTestSuite(t *testing.T) {
	suite.Run(t, new(CreateUserRBACFlowTestSuite))
}

// MockRBACServiceForHandlers implements middleware.RBACService
type MockRBACServiceForHandlers struct {
	mock.Mock
}

func (m *MockRBACServiceForHandlers) UserHasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, userID, tenantID, permission)
	return args.Bool(0), args.Error(1)
}

// Helper to satisfy the time.Time requirement for tests
var _ = time.Now
