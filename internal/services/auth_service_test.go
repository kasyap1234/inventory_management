package services

import (
	"context"
	"fmt"
	"testing"

	"agromart2/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	suite.Suite
	authService AuthService
	mockCache   *testhelpers.MockCacheService
	jwtSecret   string
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.mockCache = new(testhelpers.MockCacheService)
	suite.jwtSecret = "test-secret"
	suite.authService = NewAuthService(suite.mockCache, suite.jwtSecret, 3600, 86400)
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

func (suite *AuthServiceTestSuite) TestGenerateTokens() {
	userID := uuid.New()
	tenantID := uuid.New()
	scope := "read:write"

	suite.T().Run("Successful Token Generation", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		// Act
		tokenResponse, err := suite.authService.GenerateTokens(context.Background(), userID, tenantID, &scope)

		// Assert
		suite.NoError(err)
		suite.NotNil(tokenResponse)
		suite.NotEmpty(tokenResponse.AccessToken)
		suite.NotEmpty(tokenResponse.RefreshToken)
		suite.Equal("Bearer", tokenResponse.TokenType)
		suite.Equal(3600, tokenResponse.ExpiresIn)
		suite.Equal(&scope, tokenResponse.Scope)
		suite.Equal(userID.String(), tokenResponse.UserID)
		suite.Equal(tenantID.String(), tokenResponse.TenantID)

		// Validate the token claims
		claims, err := suite.authService.ValidateToken(context.Background(), tokenResponse.AccessToken)
		suite.NoError(err)
		suite.Equal(userID.String(), claims.UserID)
		suite.Equal(tenantID.String(), claims.TenantID)
		suite.Equal(&scope, claims.Scope)

		suite.mockCache.AssertExpectations(t)
	})
}
func (suite *AuthServiceTestSuite) TestLogin() {
	userID := uuid.New()
	tenantID := uuid.New()
	scope := "read:write"

	suite.T().Run("Successful Login", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		// Act
		tokenResponse, err := suite.authService.GenerateTokens(context.Background(), userID, tenantID, &scope)

		// Assert
		suite.NoError(err)
		suite.NotNil(tokenResponse)
		suite.NotEmpty(tokenResponse.AccessToken)
		suite.NotEmpty(tokenResponse.RefreshToken)
		suite.Equal("Bearer", tokenResponse.TokenType)
		suite.Equal(3600, tokenResponse.ExpiresIn)

		suite.mockCache.AssertExpectations(t)
	})

	suite.T().Run("Login with Account Lock", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return("locked", nil).Once()

		// Act - This would typically be called from a handler, but we're testing the service logic
		locked, err := suite.authService.IsAccountLocked(context.Background(), userID)

		// Assert
		suite.NoError(err)
		suite.True(locked)
		suite.mockCache.AssertExpectations(t)
	})

	suite.T().Run("Login with Failed Attempt Registration", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return("", fmt.Errorf("key not found")).Once()
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		// Act
		attempts, locked, err := suite.authService.RegisterFailedLoginAttempt(context.Background(), userID)

		// Assert
		suite.NoError(err)
		suite.Equal(1, attempts)
		suite.False(locked)
		suite.mockCache.AssertExpectations(t)
	})
}

func (suite *AuthServiceTestSuite) TestRegistration() {
	userID := uuid.New()

	suite.T().Run("Successful Registration", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		// Act
		token, err := suite.authService.GenerateEmailVerificationToken(context.Background(), userID)

		// Assert
		suite.NoError(err)
		suite.NotEmpty(token)
		suite.mockCache.AssertExpectations(t)
	})

	suite.T().Run("Registration with Email Verification", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()
		token, err := suite.authService.GenerateEmailVerificationToken(context.Background(), userID)
		suite.NoError(err)

		// Reset mock
		suite.mockCache.ExpectedCalls = nil
		suite.mockCache.Calls = nil

		// Arrange for consumption
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return(userID.String(), nil).Once()
		suite.mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()

		// Act
		returnedUserID, err := suite.authService.ConsumeEmailVerificationToken(context.Background(), token)

		// Assert
		suite.NoError(err)
		suite.Equal(userID, returnedUserID)
		suite.mockCache.AssertExpectations(t)
	})
}

func (suite *AuthServiceTestSuite) TestLogout() {
	userID := uuid.New()
	tenantID := uuid.New()
	scope := "read:write"

	suite.T().Run("Successful Logout", func(t *testing.T) {
		// Arrange - generate a token first
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()
		tokenResponse, err := suite.authService.GenerateTokens(context.Background(), userID, tenantID, &scope)
		suite.NoError(err)

		// Reset mock
		suite.mockCache.ExpectedCalls = nil
		suite.mockCache.Calls = nil

		// Arrange for revoke
		suite.mockCache.On("SetString", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		// Act
		err = suite.authService.RevokeToken(context.Background(), tokenResponse.AccessToken, nil)

		// Assert
		suite.NoError(err)
		suite.mockCache.AssertExpectations(t)
	})

	suite.T().Run("Logout with Clear Failed Attempts", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Twice()

		// Act
		err := suite.authService.ClearFailedLoginAttempts(context.Background(), userID)

		// Assert
		suite.NoError(err)
		suite.mockCache.AssertExpectations(t)
	})
}

func (suite *AuthServiceTestSuite) TestChangePassword() {
	userID := uuid.New()

	suite.T().Run("Successful Password Reset Token Generation", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		// Act
		token, err := suite.authService.GeneratePasswordResetToken(context.Background(), userID)

		// Assert
		suite.NoError(err)
		suite.NotEmpty(token)
		suite.mockCache.AssertExpectations(t)
	})

	suite.T().Run("Successful Password Reset Token Validation", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()
		token, err := suite.authService.GeneratePasswordResetToken(context.Background(), userID)
		suite.NoError(err)

		// Reset mock
		suite.mockCache.ExpectedCalls = nil
		suite.mockCache.Calls = nil

		// Arrange for validation
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return(userID.String(), nil).Once()

		// Act
		returnedUserID, err := suite.authService.ValidatePasswordResetToken(context.Background(), token)

		// Assert
		suite.NoError(err)
		suite.Equal(userID, returnedUserID)
		suite.mockCache.AssertExpectations(t)
	})

	suite.T().Run("Successful Password Reset Token Consumption", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()
		token, err := suite.authService.GeneratePasswordResetToken(context.Background(), userID)
		suite.NoError(err)

		// Reset mock
		suite.mockCache.ExpectedCalls = nil
		suite.mockCache.Calls = nil

		// Arrange for consumption
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return(userID.String(), nil).Once()
		suite.mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()

		// Act
		returnedUserID, err := suite.authService.ConsumePasswordResetToken(context.Background(), token)

		// Assert
		suite.NoError(err)
		suite.Equal(userID, returnedUserID)
		suite.mockCache.AssertExpectations(t)
	})
}

func (suite *AuthServiceTestSuite) TestResetPassword() {
	userID := uuid.New()

	suite.T().Run("Successful Password Reset Flow", func(t *testing.T) {
		// Arrange - generate reset token
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()
		token, err := suite.authService.GeneratePasswordResetToken(context.Background(), userID)
		suite.NoError(err)

		// Reset mock
		suite.mockCache.ExpectedCalls = nil
		suite.mockCache.Calls = nil

		// Arrange for consumption
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return(userID.String(), nil).Once()
		suite.mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()

		// Act
		returnedUserID, err := suite.authService.ConsumePasswordResetToken(context.Background(), token)

		// Assert
		suite.NoError(err)
		suite.Equal(userID, returnedUserID)
		suite.mockCache.AssertExpectations(t)
	})

	suite.T().Run("Invalid Password Reset Token", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return("", fmt.Errorf("token not found")).Once()

		// Act
		returnedUserID, err := suite.authService.ValidatePasswordResetToken(context.Background(), "invalid-token")

		// Assert
		suite.Error(err)
		suite.Equal(uuid.Nil, returnedUserID)
		suite.Contains(err.Error(), "invalid or expired password reset token")
		suite.mockCache.AssertExpectations(t)
	})
}

func (suite *AuthServiceTestSuite) TestVerifyEmail() {
	userID := uuid.New()

	suite.T().Run("Successful Email Verification", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("SetString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()
		token, err := suite.authService.GenerateEmailVerificationToken(context.Background(), userID)
		suite.NoError(err)

		// Reset mock
		suite.mockCache.ExpectedCalls = nil
		suite.mockCache.Calls = nil

		// Arrange for consumption
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return(userID.String(), nil).Once()
		suite.mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()

		// Act
		returnedUserID, err := suite.authService.ConsumeEmailVerificationToken(context.Background(), token)

		// Assert
		suite.NoError(err)
		suite.Equal(userID, returnedUserID)
		suite.mockCache.AssertExpectations(t)
	})

	suite.T().Run("Invalid Email Verification Token", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return("", fmt.Errorf("token not found")).Once()

		// Act
		returnedUserID, err := suite.authService.ConsumeEmailVerificationToken(context.Background(), "invalid-token")

		// Assert
		suite.Error(err)
		suite.Equal(uuid.Nil, returnedUserID)
		suite.Contains(err.Error(), "invalid or expired email verification token")
		suite.mockCache.AssertExpectations(t)
	})

	suite.T().Run("Malformed Email Verification Token Data", func(t *testing.T) {
		// Arrange
		suite.mockCache.On("GetString", mock.Anything, mock.AnythingOfType("string")).Return("invalid-uuid", nil).Once()
		suite.mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()

		// Act
		returnedUserID, err := suite.authService.ConsumeEmailVerificationToken(context.Background(), "some-token")

		// Assert
		suite.Error(err)
		suite.Equal(uuid.Nil, returnedUserID)
		suite.Contains(err.Error(), "invalid user ID in email verification token")
		suite.mockCache.AssertExpectations(t)
	})
}
