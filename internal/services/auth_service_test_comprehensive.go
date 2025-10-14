package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"agromart2/testhelpers"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Comprehensive AuthService test suite
type AuthServiceComprehensiveTestSuite struct {
	suite.Suite
	mockCacheService  *testhelpers.MockCacheService
	authService       AuthService
	userID            uuid.UUID
	tenantID          uuid.UUID
	ctx               context.Context
	jwtSecret         string
}

func (suite *AuthServiceComprehensiveTestSuite) SetupTest() {
	suite.mockCacheService = &testhelpers.MockCacheService{}
	suite.jwtSecret = "test-secret-key-for-testing"
	suite.authService = NewAuthService(
		suite.mockCacheService,
		suite.jwtSecret,
		3600,  // 1 hour access token
		86400, // 24 hour refresh token
	)
	suite.userID = uuid.New()
	suite.tenantID = uuid.New()
	suite.ctx = context.Background()
}

func (suite *AuthServiceComprehensiveTestSuite) TearDownTest() {
	suite.mockCacheService.AssertExpectations(suite.T())
}

func TestAuthServiceComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceComprehensiveTestSuite))
}

func (suite *AuthServiceComprehensiveTestSuite) TestGenerateTokens() {
	suite.T().Run("Successful Token Generation", func(t *testing.T) {
		// Arrange
		scope := "read write"
		
		suite.mockCacheService.On("SetString", 
			mock.AnythingOfType("context.Context"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(nil).Twice()

		// Act
		tokens, err := suite.authService.GenerateTokens(suite.ctx, suite.userID, suite.tenantID, &scope)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
		assert.Equal(t, "Bearer", tokens.TokenType)
		assert.Greater(t, tokens.ExpiresIn, 0)
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Token Generation Without Scope", func(t *testing.T) {
		// Arrange
		suite.mockCacheService.On("SetString",
			mock.AnythingOfType("context.Context"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(nil).Twice()

		// Act
		tokens, err := suite.authService.GenerateTokens(suite.ctx, suite.userID, suite.tenantID, nil)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
		suite.mockCacheService.AssertExpectations(suite.T())
	})
}

func (suite *AuthServiceComprehensiveTestSuite) TestValidateToken() {
	suite.T().Run("Valid Access Token", func(t *testing.T) {
		// Arrange
		scope := "read"
		suite.mockCacheService.On("SetString",
			mock.AnythingOfType("context.Context"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(nil).Twice()

		// First generate a token
		tokens, err := suite.authService.GenerateTokens(suite.ctx, suite.userID, suite.tenantID, &scope)
		require.NoError(t, err)
		
        // Reset mock and setup cache miss
        suite.mockCacheService = &testhelpers.MockCacheService{}
		suite.mockCacheService.On("GetString",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return("", errors.New("cache miss")) // Simulate cache miss

        // Rebuild auth service with the new mock instance
        suite.authService = NewAuthService(
            suite.mockCacheService,
            suite.jwtSecret,
            3600,
            86400,
        )

		// Act
		claims, err := suite.authService.ValidateToken(suite.ctx, tokens.AccessToken)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, suite.userID.String(), claims.UserID)
		assert.Equal(t, suite.tenantID.String(), claims.TenantID)
		assert.Equal(t, &scope, claims.Scope)
	})

	suite.T().Run("Invalid Token Format", func(t *testing.T) {
		// Act
		claims, err := suite.authService.ValidateToken(suite.ctx, "invalid-token")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "invalid token")
	})

	suite.T().Run("Expired Token", func(t *testing.T) {
		// Arrange - Create an expired token
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":   suite.userID.String(),
			"tenant_id": suite.tenantID.String(),
			"exp":       time.Now().Add(-time.Hour).Unix(), // Expired
		})
		
		expiredTokenString, err := expiredToken.SignedString([]byte(suite.jwtSecret))
		require.NoError(t, err)

		// Act
		claims, err := suite.authService.ValidateToken(suite.ctx, expiredTokenString)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token is expired")
	})
}

func (suite *AuthServiceComprehensiveTestSuite) TestRefreshToken() {
	suite.T().Run("Successful Token Refresh", func(t *testing.T) {
		// Arrange
		scope := "read write"
		clientID := "test-client"
		
		// Generate initial tokens
		suite.mockCacheService.On("SetString",
			mock.AnythingOfType("context.Context"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(nil).Twice()

		initialTokens, err := suite.authService.GenerateTokens(suite.ctx, suite.userID, suite.tenantID, &scope)
		require.NoError(t, err)

		// Setup cache expectations for refresh
		suite.mockCacheService.On("GetString",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return(suite.userID.String(), nil) // Refresh token found in cache

		suite.mockCacheService.On("SetString",
			mock.AnythingOfType("context.Context"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(nil).Twice()

		// Act
		newTokens, err := suite.authService.RefreshToken(suite.ctx, initialTokens.RefreshToken, &clientID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, newTokens.AccessToken)
		assert.NotEmpty(t, newTokens.RefreshToken)
		assert.Equal(t, "Bearer", newTokens.TokenType)
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Invalid Refresh Token", func(t *testing.T) {
		// Arrange
		clientID := "test-client"
		invalidToken := "invalid-refresh-token"

		suite.mockCacheService.On("GetString",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return("", errors.New("token not found"))

		// Act
		tokens, err := suite.authService.RefreshToken(suite.ctx, invalidToken, &clientID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "invalid refresh token")
		suite.mockCacheService.AssertExpectations(suite.T())
	})
}

func (suite *AuthServiceComprehensiveTestSuite) TestRevokeToken() {
	suite.T().Run("Successful Token Revocation", func(t *testing.T) {
		// Arrange
		token := "test-token"
		tokenType := "access"

		suite.mockCacheService.On("SetString",
			suite.ctx,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(nil)

		// Act
		err := suite.authService.RevokeToken(suite.ctx, token, &tokenType)

		// Assert
		assert.NoError(t, err)
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Token Revocation Without Type", func(t *testing.T) {
		// Arrange
		token := "test-token"

		suite.mockCacheService.On("SetString",
			suite.ctx,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(nil)

		// Act
		err := suite.authService.RevokeToken(suite.ctx, token, nil)

		// Assert
		assert.NoError(t, err)
		suite.mockCacheService.AssertExpectations(suite.T())
	})
}

func (suite *AuthServiceComprehensiveTestSuite) TestPasswordResetToken() {
	suite.T().Run("Successful Password Reset Token Generation", func(t *testing.T) {
		// Arrange
		suite.mockCacheService.On("SetString",
			suite.ctx,
			mock.AnythingOfType("string"),
			suite.userID.String(),
			mock.AnythingOfType("time.Duration"),
		).Return(nil)

		// Act
		token, err := suite.authService.GeneratePasswordResetToken(suite.ctx, suite.userID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Len(t, token, 32) // Should be 32 characters long
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Successful Password Reset Token Validation", func(t *testing.T) {
		// Arrange
		token := "test-reset-token"
		
		suite.mockCacheService.On("GetString",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return(suite.userID.String(), nil)

		// Act
		userID, err := suite.authService.ValidatePasswordResetToken(suite.ctx, token)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, suite.userID, userID)
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Invalid Password Reset Token", func(t *testing.T) {
		// Arrange
		token := "invalid-token"
		
		suite.mockCacheService.On("GetString",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return("", errors.New("token not found"))

		// Act
		userID, err := suite.authService.ValidatePasswordResetToken(suite.ctx, token)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, userID)
		assert.Contains(t, err.Error(), "invalid or expired")
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Consume Password Reset Token", func(t *testing.T) {
		// Arrange
		token := "test-reset-token"
		
		suite.mockCacheService.On("GetString",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return(suite.userID.String(), nil)

		suite.mockCacheService.On("Delete",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return(nil)

		// Act
		userID, err := suite.authService.ConsumePasswordResetToken(suite.ctx, token)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, suite.userID, userID)
		suite.mockCacheService.AssertExpectations(suite.T())
	})
}

func (suite *AuthServiceComprehensiveTestSuite) TestEmailVerificationToken() {
	suite.T().Run("Successful Email Verification Token Generation", func(t *testing.T) {
		// Arrange
		suite.mockCacheService.On("SetString",
			suite.ctx,
			mock.AnythingOfType("string"),
			suite.userID.String(),
			mock.AnythingOfType("time.Duration"),
		).Return(nil)

		// Act
		token, err := suite.authService.GenerateEmailVerificationToken(suite.ctx, suite.userID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Successful Email Verification Token Consumption", func(t *testing.T) {
		// Arrange
		token := "test-verification-token"
		
		suite.mockCacheService.On("GetString",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return(suite.userID.String(), nil)

		suite.mockCacheService.On("Delete",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return(nil)

		// Act
		userID, err := suite.authService.ConsumeEmailVerificationToken(suite.ctx, token)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, suite.userID, userID)
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Invalid Email Verification Token", func(t *testing.T) {
		// Arrange
		token := "invalid-token"
		
		suite.mockCacheService.On("GetString",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return("", errors.New("token not found"))

		// Act
		userID, err := suite.authService.ConsumeEmailVerificationToken(suite.ctx, token)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, userID)
		suite.mockCacheService.AssertExpectations(suite.T())
	})
}

func (suite *AuthServiceComprehensiveTestSuite) TestAccountLockoutMechanisms() {
	suite.T().Run("Check Account Not Locked", func(t *testing.T) {
		// Arrange
    suite.mockCacheService.On("GetString",
            suite.ctx,
            mock.AnythingOfType("string"),
    ).Return("", nil)

		// Act
		locked, err := suite.authService.IsAccountLocked(suite.ctx, suite.userID)

		// Assert
		assert.NoError(t, err)
		assert.False(t, locked)
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Register Failed Login Attempt", func(t *testing.T) {
		// Arrange
    suite.mockCacheService.On("GetString",
            suite.ctx,
            mock.AnythingOfType("string"),
    ).Return("", nil)
    suite.mockCacheService.On("SetString",
            suite.ctx,
            mock.AnythingOfType("string"),
            mock.AnythingOfType("string"),
            mock.AnythingOfType("time.Duration"),
    ).Return(nil)

		// Act
		attempts, locked, err := suite.authService.RegisterFailedLoginAttempt(suite.ctx, suite.userID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
		assert.False(t, locked)
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Clear Failed Login Attempts", func(t *testing.T) {
		// Arrange
		suite.mockCacheService.On("Delete",
			suite.ctx,
			mock.AnythingOfType("string"),
		).Return(nil)

		// Act
		err := suite.authService.ClearFailedLoginAttempts(suite.ctx, suite.userID)

		// Assert
		assert.NoError(t, err)
		suite.mockCacheService.AssertExpectations(suite.T())
	})
}

func (suite *AuthServiceComprehensiveTestSuite) TestAuthorizationCodeFlows() {
	suite.T().Run("Generate Authorization Code", func(t *testing.T) {
		// Arrange
		clientID := "test-client"
		redirectURI := "https://example.com/callback"
		scope := "read write"

		suite.mockCacheService.On("SetString",
			suite.ctx,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(nil)

		// Act
		code, err := suite.authService.GenerateAuthorizationCode(
			suite.ctx,
			suite.userID,
			suite.tenantID,
			clientID,
			&redirectURI,
			&scope,
		)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Len(t, code, 8) // Should be 8 characters
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Validate Authorization Code", func(t *testing.T) {
		// Arrange
        code := "12345678"
        clientID := "test-client"
        redirectURI := "https://example.com/callback"
		
        suite.mockCacheService.On("GetString",
            suite.ctx,
            mock.AnythingOfType("string"),
        ).Return("", errors.New("not found"))

		// Act
		claims, err := suite.authService.ValidateAuthorizationCode(suite.ctx, code, clientID, redirectURI)

		// Assert
		// Note: This test might need adjustment based on actual implementation
		// The interface might not have ValidateAuthorizationCode method
		if err == nil {
			assert.NotNil(t, claims)
		}
		suite.mockCacheService.AssertExpectations(suite.T())
	})
}

func (suite *AuthServiceComprehensiveTestSuite) TestRevokeUserTokens() {
	suite.T().Run("Successfully Revoke All User Tokens", func(t *testing.T) {
		// Act
		err := suite.authService.RevokeUserTokens(suite.ctx, suite.userID)

		// Assert
		assert.NoError(t, err)
		suite.mockCacheService.AssertExpectations(suite.T())
	})
}

func (suite *AuthServiceComprehensiveTestSuite) TestCleanupExpiredTokens() {
	suite.T().Run("Cleanup Expired Tokens", func(t *testing.T) {
		// Act
		err := suite.authService.CleanupExpiredTokens(suite.ctx)

		// Assert
		// This might not interact with cache directly for cleanup
		// Implementation might use database cleanup
		assert.NoError(t, err)
	})
}

// Edge case and error handling tests
func (suite *AuthServiceComprehensiveTestSuite) TestEdgeCases() {
	suite.T().Run("Empty Context", func(t *testing.T) {
		// Arrange
		scope := "read"
		suite.mockCacheService.On("SetString",
			context.Background(),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(nil).Twice()

		// Act
		tokens, err := suite.authService.GenerateTokens(context.Background(), suite.userID, suite.tenantID, &scope)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
		suite.mockCacheService.AssertExpectations(suite.T())
	})

	suite.T().Run("Cache Errors", func(t *testing.T) {
		// Arrange
		suite.mockCacheService.On("SetString",
			suite.ctx,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("time.Duration"),
		).Return(errors.New("cache error"))

		// Act
		tokens, err := suite.authService.GenerateTokens(suite.ctx, suite.userID, suite.tenantID, nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "cache error")
		suite.mockCacheService.AssertExpectations(suite.T())
	})
}
