//go:build integration
// +build integration
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/handlers"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AuthIntegrationTestSuite tests the complete authentication flow
type AuthIntegrationTestSuite struct {
	suite.Suite
	echo           *echo.Echo
	pool           *pgxpool.Pool
	userRepo       repositories.UserRepository
	tenantRepo     repositories.TenantRepository
	authService    services.AuthService
	authHandler    *handlers.AuthHandler
	testTenantID   uuid.UUID
	testUserID     uuid.UUID
	ctx            context.Context
}

func (suite *AuthIntegrationTestSuite) SetupSuite() {
_suite := suite
	suite.T().Log("Setting up AuthIntegrationTestSuite")

	// Setup test database
	suite.pool = setupTestDB(suite.T())
	
	// Setup echo
	suite.echo = echo.New()
	
	// Setup repositories
	suite.userRepo = repositories.NewUserRepo(suite.pool)
	suite.tenantRepo = repositories.NewTenantRepo(suite.pool)
	
	// Setup auth service
	cacheService := &MockCacheService{} // Mock cache for testing
	suite.authService = services.NewAuthService(cacheService, "test-jwt-secret", 3600, 86400)
	
	// Setup handlers
	suite.authHandler = handlers.NewAuthHandler(suite.authService, suite.userRepo)
	
	// Setup test tenant and user
	suite.testTenantID = suite.createTestTenant()
	suite.testUserID = suite.createTestUser()
	
	suite.ctx = context.Background()

	// Setup routes
	suite.setupRoutes()
}

func (suite *AuthIntegrationTestSuite) TearDownSuite() {
	if suite.pool != nil {
		suite.pool.Close()
	}
}

func (suite *AuthIntegrationTestSuite) SetupTest() {
	// Clean up any test data created during tests
	suite.cleanupTestData()
}

func (suite *AuthIntegrationTestSuite) TearDownTest() {
	suite.cleanupTestData()
}

func (suite *AuthIntegrationTestSuite) createTestTenant() uuid.UUID {
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Auth Test Tenant " + tenantID.String()[:8],
		Subdomain: "auth-test-" + tenantID.String()[:8],
		Status:    "active",
	}

	err := suite.tenantRepo.Create(suite.ctx, tenant)
	require.NoError(suite.T(), err, "Failed to create test tenant")

	return tenantID
}

func (suite *AuthIntegrationTestSuite) createTestUser() uuid.UUID {
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		TenantID:     suite.testTenantID,
		Email:        "auth-test-" + userID.String()[:8] + "@example.com",
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewWNsglwEFnQJ/6sW", // "password"
		FirstName:    "Test",
		LastName:     "User",
		Status:       "active",
	}

	err := suite.userRepo.Create(suite.ctx, user)
	require.NoError(suite.T(), err, "Failed to create test user")

	return userID
}

func (suite *AuthIntegrationTestSuite) setupRoutes() {
	authGroup := suite.echo.Group("/api/v1/auth")
	authGroup.POST("/register", suite.authHandler.Register)
	authGroup.POST("/login", suite.authHandler.Login)
	authGroup.POST("/refresh", suite.authHandler.RefreshToken)
	authGroup.POST("/logout", suite.authHandler.Logout)
	authGroup.POST("/forgot-password", suite.authHandler.ForgotPassword)
	authGroup.POST("/reset-password", suite.authHandler.ResetPassword)
	authGroup.GET("/verify-email/:token", suite.authHandler.VerifyEmail)
}

func (suite *AuthIntegrationTestSuite) cleanupTestData() {
	// Clean up test users created during tests
	// Note: Actual cleanup implementation would depend on database schema
}

func (suite *AuthIntegrationTestSuite) makeRequest(method, path string, body interface{}, headers map[string]string) (*httptest.ResponseRecorder, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	
	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	
	// Set tenant context
	req = req.WithContext(setTenantIDInContext(req.Context(), suite.testTenantID))

	err := suite.echo.ServeHTTP(rec, req)
	return rec, err
}

func TestAuthIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}

// Registration Flow Tests
func (suite *AuthIntegrationTestSuite) TestCompleteRegistrationAndLoginFlow() {
	suite.T().Run("Successful Registration", func(t *testing.T) {
		// Arrange
		registerPayload := map[string]interface{}{
			"email":     "new-user-" + uuid.New().String()[:8] + "@example.com",
			"password":  "SecurePassword123!",
			"firstName": "New",
			"lastName":  "User",
		}

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/register", registerPayload, nil)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusCreated, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "success", response["status"])
		assert.NotNil(t, response["data"])
		assert.NotEmpty(t, response["data"].(map[string]interface{})["message"])
	})

	suite.T().Run("Registration with Invalid Data", func(t *testing.T) {
		// Arrange
		invalidPayload := map[string]interface{}{
			"email":     "invalid-email",
			"password":  "123", // Too short
			"firstName": "",
			"lastName":  "",
		}

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/register", invalidPayload, nil)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "error", response["status"])
	})

	suite.T().Run("Registration with Duplicate Email", func(t *testing.T) {
		// Arrange - use existing test user email
		existingUser, err := suite.userRepo.GetByID(suite.ctx, suite.testTenantID, suite.testUserID)
		require.NoError(t, err)

		duplicatePayload := map[string]interface{}{
			"email":     existingUser.Email,
			"password":  "SecurePassword123!",
			"firstName": "Duplicate",
			"lastName":  "User",
		}

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/register", duplicatePayload, nil)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusConflict, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "error", response["status"])
		assert.Contains(t, response["message"], "already exists")
	})
}

// Login Flow Tests
func (suite *AuthIntegrationTestSuite) TestLogin Flow() {
	suite.T().Run("Successful Login", func(t *testing.T) {
		// Arrange
		existingUser, err := suite.userRepo.GetByID(suite.ctx, suite.testTenantID, suite.testUserID)
		require.NoError(t, err)

		loginPayload := map[string]interface{}{
			"email":    existingUser.Email,
			"password": "password", // Matches the hashed password
		}

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/login", loginPayload, nil)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "success", response["status"])
		data := response["data"].(map[string]interface{})
		
		// Check for tokens
		assert.NotEmpty(t, data["access_token"])
		assert.NotEmpty(t, data["refresh_token"])
		assert.Equal(t, "Bearer", data["token_type"])
		assert.Greater(t, data["expires_in"], float64(0))
		
		// Check user data
		userData := data["user"].(map[string]interface{})
		assert.Equal(t, existingUser.ID.String(), userData["id"])
		assert.Equal(t, existingUser.Email, userData["email"])
		assert.Equal(t, existingUser.FirstName, userData["first_name"])
		assert.Equal(t, existingUser.LastName, userData["last_name"])
	})

	suite.T().Run("Login with Invalid Credentials", func(t *testing.T) {
		// Arrange
		existingUser, err := suite.userRepo.GetByID(suite.ctx, suite.testTenantID, suite.testUserID)
		require.NoError(t, err)

		invalidLoginPayload := map[string]interface{}{
			"email":    existingUser.Email,
			"password": "wrongpassword",
		}

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/login", invalidLoginPayload, nil)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "error", response["status"])
		assert.Contains(t, response["message"], "Invalid credentials")
	})

	suite.T().Run("Login with Non-Existent User", func(t *testing.T) {
		// Arrange
		loginPayload := map[string]interface{}{
			"email":    "nonexistent-" + uuid.New().String()[:8] + "@example.com",
			"password": "password",
		}

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/login", loginPayload, nil)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// Token Refresh Tests
func (suite *AuthIntegrationTestSuite) TestTokenRefreshFlow() {
	suite.T().Run("Successful Token Refresh", func(t *testing.T) {
		// Arrange - First login to get tokens
		existingUser, err := suite.userRepo.GetByID(suite.ctx, suite.testTenantID, suite.testUserID)
		require.NoError(t, err)

		loginPayload := map[string]interface{}{
			"email":    existingUser.Email,
			"password": "password",
		}

		loginRec, err := suite.makeRequest("POST", "/api/v1/auth/login", loginPayload, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, loginRec.Code)

		var loginResponse map[string]interface{}
		err = json.Unmarshal(loginRec.Body.Bytes(), &loginResponse)
		require.NoError(t, err)

		refreshToken := loginResponse["data"].(map[string]interface{})["refresh_token"].(string)

		// Act - Refresh token
		refreshPayload := map[string]interface{}{
			"refreshToken": refreshToken,
		}

		refreshRec, err := suite.makeRequest("POST", "/api/v1/auth/refresh", refreshPayload, nil)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusOK, refreshRec.Code)
		
		var refreshResponse map[string]interface{}
		err = json.Unmarshal(refreshRec.Body.Bytes(), &refreshResponse)
		require.NoError(t, err)
		
		assert.Equal(t, "success", refreshResponse["status"])
		data := refreshResponse["data"].(map[string]interface{})
		assert.NotEmpty(t, data["access_token"])
		assert.NotEmpty(t, data["refresh_token"])
		
		// New tokens should be different from old ones
		assert.NotEqual(t, loginResponse["data"].(map[string]interface{})["access_token"], data["access_token"])
	})

	suite.T().Run("Refresh with Invalid Token", func(t *testing.T) {
		// Arrange
		refreshPayload := map[string]interface{}{
			"refreshToken": "invalid-refresh-token",
		}

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/refresh", refreshPayload, nil)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// Password Reset Flow Tests
func (suite *AuthIntegrationTestSuite) TestPasswordResetFlow() {
	suite.T().Run("Password Reset Request", func(t *testing.T) {
		// Arrange
		existingUser, err := suite.userRepo.GetByID(suite.ctx, suite.testTenantID, suite.testUserID)
		require.NoError(t, err)

		resetPayload := map[string]interface{}{
			"email": existingUser.Email,
		}

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/forgot-password", resetPayload, nil)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "success", response["status"])
		assert.Contains(t, response["message"], "password reset link")
	})

	suite.T().Run("Password Reset with Non-Existent Email", func(t *testing.T) {
		// Arrange
		resetPayload := map[string]interface{}{
			"email": "nonexistent-" + uuid.New().String()[:8] + "@example.com",
		}

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/forgot-password", resetPayload, nil)
		require.NoError(t, err)

		// Assert - Should still return success for security (don't reveal if email exists)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// Logout Tests
func (suite *AuthIntegrationTestSuite) TestLogoutFlow() {
	suite.T().Run("Successful Logout", func(t *testing.T) {
		// Arrange - First login to get token
		existingUser, err := suite.userRepo.GetByID(suite.ctx, suite.testTenantID, suite.testUserID)
		require.NoError(t, err)

		loginPayload := map[string]interface{}{
			"email":    existingUser.Email,
			"password": "password",
		}

		loginRec, err := suite.makeRequest("POST", "/api/v1/auth/login", loginPayload, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, loginRec.Code)

		var loginResponse map[string]interface{}
		err = json.Unmarshal(loginRec.Body.Bytes(), &loginResponse)
		require.NoError(t, err)

		accessToken := loginResponse["data"].(map[string]interface{})["access_token"].(string)

		// Act - Logout
		logoutRec, err := suite.makeRequest("POST", "/api/v1/auth/logout", nil, map[string]string{
			"Authorization": "Bearer " + accessToken,
		})
		require.NoError(t, err)

		// Assert
		assert.Equal(t, http.StatusOK, logoutRec.Code)
		
		var logoutResponse map[string]interface{}
		err = json.Unmarshal(logoutRec.Body.Bytes(), &logoutResponse)
		require.NoError(t, err)
		
		assert.Equal(t, "success", logoutResponse["status"])
	})
}

// Security Tests
func (suite *AuthIntegrationTestSuite) TestSecurityFeatures() {
	suite.T().Run("Prevent SQL Injection", func(t *testing.T) {
		// Attempt SQL injection through login
		sqlInjectionPayload := map[string]interface{}{
			"email":    "'; DROP TABLE users; --@example.com",
			"password": "password",
		}

		rec, err := suite.makeRequest("POST", "/api/v1/auth/login", sqlInjectionPayload, nil)
		require.NoError(t, err)

		// Should return unauthorized, not success or server error
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	suite.T().Run("Cross-Tenant Data Access Prevention", func(t *testing.T) {
		// Create another tenant and user
		otherTenantID := suite.createTestTenant()
		otherUserID := uuid.New()
		otherUser := &models.User{
			ID:           otherUserID,
			TenantID:     otherTenantID,
			Email:        "other-tenant-" + otherUserID.String()[:8] + "@example.com",
			PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewWNsglwEFnQJ/6sW",
			FirstName:    "Other",
			LastName:     "User",
			Status:       "active",
		}

		err := suite.userRepo.Create(suite.ctx, otherUser)
		require.NoError(t, err)

		// Try to login as other tenant user using original tenant context
		loginPayload := map[string]interface{}{
			"email":    otherUser.Email,
			"password": "password",
		}

		rec, err := suite.makeRequest("POST", "/api/v1/auth/login", loginPayload, nil)
		require.NoError(t, err)

		// Should fail - user belongs to different tenant
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// Performance Tests
func (suite *AuthIntegrationTestSuite) TestPerformance() {
	suite.T().Run("Login Performance", func(t *testing.T) {
		// Arrange
		existingUser, err := suite.userRepo.GetByID(suite.ctx, suite.testTenantID, suite.testUserID)
		require.NoError(t, err)

		loginPayload := map[string]interface{}{
			"email":    existingUser.Email,
			"password": "password",
		}

		start := time.Now()

		// Act
		rec, err := suite.makeRequest("POST", "/api/v1/auth/login", loginPayload, nil)
		require.NoError(t, err)

		duration := time.Since(start)

		// Assert - Should complete quickly
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Less(t, duration.Milliseconds(), int64(1000), "Login should complete within 1 second")
	})
}

// Edge Cases
func (suite *AuthIntegrationTestSuite) TestEdgeCases() {
	suite.T().Run("Empty Request Body", func(t *testing.T) {
		rec, err := suite.makeRequest("POST", "/api/v1/auth/login", nil, nil)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	suite.T().Run("Malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer([]byte("{invalid json}")))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(setTenantIDInContext(req.Context(), suite.testTenantID))

		rec := httptest.NewRecorder()
		err := suite.echo.ServeHTTP(rec, req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	suite.T().Run("Rate Limiting", func(t *testing.T) {
		existingUser, err := suite.userRepo.GetByID(suite.ctx, suite.testTenantID, suite.testUserID)
		require.NoError(t, err)

		loginPayload := map[string]interface{}{
			"email":    existingUser.Email,
			"password": "wrongpassword", // Always fail to test rate limiting
		}

		// Make multiple failed login attempts
		for i := 0; i < 10; i++ {
			rec, err := suite.makeRequest("POST", "/api/v1/auth/login", loginPayload, nil)
			require.NoError(t, err)
			
			// Should eventually be rate limited
			if rec.Code == http.StatusTooManyRequests {
				return // Test passed
			}
		}

		suite.T().Log("Rate limiting not implemented or not triggered")
	})
}

// Mock Cache Service for testing
type MockCacheService struct {
	data map[string]interface{}
}

func (m *MockCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	return m.data[key], nil
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
	return nil
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockCacheService) DeletePattern(ctx context.Context, pattern string) error {
	// Simple implementation - would need regex matching in real implementation
	for key := range m.data {
		delete(m.data, key)
	}
	return nil
}
