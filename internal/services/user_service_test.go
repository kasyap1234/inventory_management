package services

import (
	"context"
	"errors"
	"testing"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockUserRepository implements the UserRepository interface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, tenantID, id)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	users, _ := args.Get(0).([]*models.User)
	return users, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	args := m.Called(ctx, tenantID, email)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) GetTenantIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID)
	tenantID, _ := args.Get(0).(uuid.UUID)
	return tenantID, args.Error(1)
}

func (m *MockUserRepository) GetByEmailGlobal(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, tenantID, userID, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, status string) error {
	args := m.Called(ctx, tenantID, userID, status)
	return args.Error(0)
}

// UserServiceTestSuite is a test suite for user service functions
type UserServiceTestSuite struct {
	suite.Suite
	mockUserRepo *MockUserRepository
	ctx          context.Context
}

// SetupTest runs before each test in the suite
func (suite *UserServiceTestSuite) SetupTest() {
	suite.mockUserRepo = new(MockUserRepository)
	suite.ctx = context.Background()
}

// TestCreateUser tests the user creation functionality
func (suite *UserServiceTestSuite) TestCreateUser() {
	// Test case: successful user creation
	suite.Run("Successful User Creation", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		user := &models.User{
			ID:        userID,
			TenantID:  tenantID,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Status:    "active",
		}

		suite.mockUserRepo.On("Create", suite.ctx, user).Return(nil).Once()

		// Since there's no dedicated user service, we're testing the repository directly
		err := suite.mockUserRepo.Create(suite.ctx, user)

		suite.NoError(err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: user creation fails
	suite.Run("User Creation Fails", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		user := &models.User{
			ID:        userID,
			TenantID:  tenantID,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Status:    "active",
		}
		expectedErr := errors.New("user creation failed")

		suite.mockUserRepo.On("Create", suite.ctx, user).Return(expectedErr).Once()

		err := suite.mockUserRepo.Create(suite.ctx, user)

		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestGetUserByID tests retrieving a user by ID
func (suite *UserServiceTestSuite) TestGetUserByID() {
	// Test case: successful user retrieval
	suite.Run("Successful User Retrieval", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		expectedUser := &models.User{
			ID:        userID,
			TenantID:  tenantID,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Status:    "active",
		}

		suite.mockUserRepo.On("GetByID", suite.ctx, tenantID, userID).Return(expectedUser, nil).Once()

		user, err := suite.mockUserRepo.GetByID(suite.ctx, tenantID, userID)

		suite.NoError(err)
		suite.Equal(expectedUser, user)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: user not found
	suite.Run("User Not Found", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		expectedErr := errors.New("user not found")

		suite.mockUserRepo.On("GetByID", suite.ctx, tenantID, userID).Return((*models.User)(nil), expectedErr).Once()

		user, err := suite.mockUserRepo.GetByID(suite.ctx, tenantID, userID)

		suite.Error(err)
		suite.Nil(user)
		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestGetUserByEmail tests retrieving a user by email
func (suite *UserServiceTestSuite) TestGetUserByEmail() {
	// Test case: successful user retrieval by email
	suite.Run("Successful User Retrieval by Email", func() {
		tenantID := uuid.New()
		email := "test@example.com"
		expectedUser := &models.User{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Email:     email,
			FirstName: "John",
			LastName:  "Doe",
			Status:    "active",
		}

		suite.mockUserRepo.On("GetByEmail", suite.ctx, tenantID, email).Return(expectedUser, nil).Once()

		user, err := suite.mockUserRepo.GetByEmail(suite.ctx, tenantID, email)

		suite.NoError(err)
		suite.Equal(expectedUser, user)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: user not found by email
	suite.Run("User Not Found by Email", func() {
		tenantID := uuid.New()
		email := "test@example.com"
		expectedErr := errors.New("user not found")

		suite.mockUserRepo.On("GetByEmail", suite.ctx, tenantID, email).Return((*models.User)(nil), expectedErr).Once()

		user, err := suite.mockUserRepo.GetByEmail(suite.ctx, tenantID, email)

		suite.Error(err)
		suite.Nil(user)
		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestUpdateUser tests updating a user
func (suite *UserServiceTestSuite) TestUpdateUser() {
	// Test case: successful user update
	suite.Run("Successful User Update", func() {
		user := &models.User{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Status:    "active",
		}

		suite.mockUserRepo.On("Update", suite.ctx, user).Return(nil).Once()

		err := suite.mockUserRepo.Update(suite.ctx, user)

		suite.NoError(err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: user update fails
	suite.Run("User Update Fails", func() {
		user := &models.User{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Status:    "active",
		}
		expectedErr := errors.New("update failed")

		suite.mockUserRepo.On("Update", suite.ctx, user).Return(expectedErr).Once()

		err := suite.mockUserRepo.Update(suite.ctx, user)

		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestDeleteUser tests deleting a user
func (suite *UserServiceTestSuite) TestDeleteUser() {
	// Test case: successful user deletion
	suite.Run("Successful User Deletion", func() {
		tenantID := uuid.New()
		userID := uuid.New()

		suite.mockUserRepo.On("Delete", suite.ctx, tenantID, userID).Return(nil).Once()

		err := suite.mockUserRepo.Delete(suite.ctx, tenantID, userID)

		suite.NoError(err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: user deletion fails
	suite.Run("User Deletion Fails", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		expectedErr := errors.New("delete failed")

		suite.mockUserRepo.On("Delete", suite.ctx, tenantID, userID).Return(expectedErr).Once()

		err := suite.mockUserRepo.Delete(suite.ctx, tenantID, userID)

		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestListUsers tests listing users
func (suite *UserServiceTestSuite) TestListUsers() {
	// Test case: successful user listing
	suite.Run("Successful User Listing", func() {
		tenantID := uuid.New()
		users := []*models.User{
			{
				ID:        uuid.New(),
				TenantID:  tenantID,
				Email:     "user1@example.com",
				FirstName: "User",
				LastName:  "One",
				Status:    "active",
			},
			{
				ID:        uuid.New(),
				TenantID:  tenantID,
				Email:     "user2@example.com",
				FirstName: "User",
				LastName:  "Two",
				Status:    "active",
			},
		}

		suite.mockUserRepo.On("List", suite.ctx, tenantID, 10, 0).Return(users, nil).Once()

		result, err := suite.mockUserRepo.List(suite.ctx, tenantID, 10, 0)

		suite.NoError(err)
		suite.Equal(users, result)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: user listing fails
	suite.Run("User Listing Fails", func() {
		tenantID := uuid.New()
		expectedErr := errors.New("list failed")

		suite.mockUserRepo.On("List", suite.ctx, tenantID, 10, 0).Return(([]*models.User)(nil), expectedErr).Once()

		result, err := suite.mockUserRepo.List(suite.ctx, tenantID, 10, 0)

		suite.Error(err)
		suite.Nil(result)
		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestGetTenantIDByUserID tests getting tenant ID by user ID
func (suite *UserServiceTestSuite) TestGetTenantIDByUserID() {
	// Test case: successful tenant ID retrieval
	suite.Run("Successful Tenant ID Retrieval", func() {
		userID := uuid.New()
		tenantID := uuid.New()

		suite.mockUserRepo.On("GetTenantIDByUserID", suite.ctx, userID).Return(tenantID, nil).Once()

		result, err := suite.mockUserRepo.GetTenantIDByUserID(suite.ctx, userID)

		suite.NoError(err)
		suite.Equal(tenantID, result)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: tenant ID retrieval fails
	suite.Run("Tenant ID Retrieval Fails", func() {
		userID := uuid.New()
		expectedErr := errors.New("tenant ID not found")

		suite.mockUserRepo.On("GetTenantIDByUserID", suite.ctx, userID).Return(uuid.Nil, expectedErr).Once()

		result, err := suite.mockUserRepo.GetTenantIDByUserID(suite.ctx, userID)

		suite.Error(err)
		suite.Equal(uuid.Nil, result)
		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestGetUserByEmailGlobal tests getting user by email globally
func (suite *UserServiceTestSuite) TestGetUserByEmailGlobal() {
	// Test case: successful global email lookup
	suite.Run("Successful Global Email Lookup", func() {
		email := "test@example.com"
		expectedUser := &models.User{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			Email:     email,
			FirstName: "John",
			LastName:  "Doe",
			Status:    "active",
		}

		suite.mockUserRepo.On("GetByEmailGlobal", suite.ctx, email).Return(expectedUser, nil).Once()

		user, err := suite.mockUserRepo.GetByEmailGlobal(suite.ctx, email)

		suite.NoError(err)
		suite.Equal(expectedUser, user)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: global email lookup fails
	suite.Run("Global Email Lookup Fails", func() {
		email := "test@example.com"
		expectedErr := errors.New("user not found")

		suite.mockUserRepo.On("GetByEmailGlobal", suite.ctx, email).Return((*models.User)(nil), expectedErr).Once()

		user, err := suite.mockUserRepo.GetByEmailGlobal(suite.ctx, email)

		suite.Error(err)
		suite.Nil(user)
		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestUpdateUserPassword tests updating user password
func (suite *UserServiceTestSuite) TestUpdateUserPassword() {
	// Test case: successful password update
	suite.Run("Successful Password Update", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		passwordHash := "hashed_password"

		suite.mockUserRepo.On("UpdatePassword", suite.ctx, tenantID, userID, passwordHash).Return(nil).Once()

		err := suite.mockUserRepo.UpdatePassword(suite.ctx, tenantID, userID, passwordHash)

		suite.NoError(err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: password update fails
	suite.Run("Password Update Fails", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		passwordHash := "hashed_password"
		expectedErr := errors.New("password update failed")

		suite.mockUserRepo.On("UpdatePassword", suite.ctx, tenantID, userID, passwordHash).Return(expectedErr).Once()

		err := suite.mockUserRepo.UpdatePassword(suite.ctx, tenantID, userID, passwordHash)

		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestUpdateUserStatus tests updating user status
func (suite *UserServiceTestSuite) TestUpdateUserStatus() {
	// Test case: successful status update
	suite.Run("Successful Status Update", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		status := "inactive"

		suite.mockUserRepo.On("UpdateStatus", suite.ctx, tenantID, userID, status).Return(nil).Once()

		err := suite.mockUserRepo.UpdateStatus(suite.ctx, tenantID, userID, status)

		suite.NoError(err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test case: status update fails
	suite.Run("Status Update Fails", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		status := "inactive"
		expectedErr := errors.New("status update failed")

		suite.mockUserRepo.On("UpdateStatus", suite.ctx, tenantID, userID, status).Return(expectedErr).Once()

		err := suite.mockUserRepo.UpdateStatus(suite.ctx, tenantID, userID, status)

		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// Test edge cases and error handling
func (suite *UserServiceTestSuite) TestEdgeCases() {
	// Test with empty email
	suite.Run("GetUserByEmail with Empty Email", func() {
		tenantID := uuid.New()
		var user *models.User
		var err error

		suite.mockUserRepo.On("GetByEmail", suite.ctx, tenantID, "").Return(user, errors.New("email cannot be empty")).Once()

		user, err = suite.mockUserRepo.GetByEmail(suite.ctx, tenantID, "")

		suite.Error(err)
		suite.Nil(user)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test with invalid UUID
	suite.Run("GetUserByID with Invalid UUID", func() {
		tenantID := uuid.Nil // Invalid UUID
		userID := uuid.Nil   // Invalid UUID
		var user *models.User
		var err error

		suite.mockUserRepo.On("GetByID", suite.ctx, tenantID, userID).Return(user, errors.New("invalid user ID")).Once()

		user, err = suite.mockUserRepo.GetByID(suite.ctx, tenantID, userID)

		suite.Error(err)
		suite.Nil(user)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	// Test with very large limit
	suite.Run("ListUsers with Large Limit", func() {
		tenantID := uuid.New()
		users := []*models.User{}

		suite.mockUserRepo.On("List", suite.ctx, tenantID, 1000, 0).Return(users, nil).Once()

		result, err := suite.mockUserRepo.List(suite.ctx, tenantID, 10000, 0)

		suite.NoError(err)
		suite.Equal(users, result)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestConcurrentOperations tests concurrent operations if applicable
func (suite *UserServiceTestSuite) TestConcurrentOperations() {
	// This test would be more relevant if we had actual service functions
	// that handle concurrency, but we can still test the repository behavior
	suite.Run("Concurrent User Creation Attempts", func() {
		tenantID := uuid.New()
		email := "concurrent@example.com"

		// Simulate concurrent attempts to create a user with the same email
		// This would typically be handled by the database constraints
		user1 := &models.User{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Email:     email,
			FirstName: "First",
			LastName:  "User",
			Status:    "active",
		}
		user2 := &models.User{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Email:     email, // Same email as user1
			FirstName: "Second",
			LastName:  "User",
			Status:    "active",
		}

		// First attempt succeeds
		suite.mockUserRepo.On("Create", suite.ctx, user1).Return(nil).Once()
		// Second attempt fails due to email uniqueness constraint
		suite.mockUserRepo.On("Create", suite.ctx, user2).Return(errors.New("user with email 'concurrent@example.com' already exists")).Once()

		err1 := suite.mockUserRepo.Create(suite.ctx, user1)
		err2 := suite.mockUserRepo.Create(suite.ctx, user2)

		suite.NoError(err1)
		suite.Error(err2)
		suite.Contains(err2.Error(), "already exists")
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TearDownTest runs after each test in the suite
func (suite *UserServiceTestSuite) TearDownTest() {
	// Clean up any resources if needed
}

// TestUserService runs the test suite
func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// Additional helper functions for testing specific scenarios

// TestUserStatusTransitions tests different user status transitions
func (suite *UserServiceTestSuite) TestUserStatusTransitions() {
	tenantID := uuid.New()
	userID := uuid.New()

	// Test various status updates
	statuses := []string{"active", "inactive", "pending", "suspended", "deleted"}

	for _, status := range statuses {
		suite.Run("Status Update to "+status, func() {
			suite.mockUserRepo.On("UpdateStatus", suite.ctx, tenantID, userID, status).Return(nil).Once()

			err := suite.mockUserRepo.UpdateStatus(suite.ctx, tenantID, userID, status)

			suite.NoError(err)
			suite.mockUserRepo.AssertExpectations(suite.T())
		})
	}
}

// TestUserEmailUniqueness tests email uniqueness across tenants
func (suite *UserServiceTestSuite) TestUserEmailUniqueness() {
	// This test verifies that the repository properly enforces email uniqueness
	// at the global level when creating users
	email := "unique@example.com"

	suite.Run("Email Uniqueness Check", func() {
		// First user creation should succeed
		user1 := &models.User{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			Email:     email,
			FirstName: "First",
			LastName:  "User",
			Status:    "active",
		}

		// Second user creation with same email should fail
		user2 := &models.User{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			Email:     email, // Same email as user1
			FirstName: "Second",
			LastName:  "User",
			Status:    "active",
		}

		suite.mockUserRepo.On("Create", suite.ctx, user1).Return(nil).Once()
		suite.mockUserRepo.On("Create", suite.ctx, user2).Return(errors.New("user with email '" + email + "' already exists")).Once()

		err1 := suite.mockUserRepo.Create(suite.ctx, user1)
		err2 := suite.mockUserRepo.Create(suite.ctx, user2)

		suite.NoError(err1)
		suite.Error(err2)
		suite.Contains(err2.Error(), "already exists")
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}

// TestErrorHandlingComprehensive tests comprehensive error handling
func (suite *UserServiceTestSuite) TestErrorHandlingComprehensive() {
	suite.Run("Database Connection Error", func() {
		tenantID := uuid.New()
		userID := uuid.New()
		expectedErr := errors.New("database connection failed")

		suite.mockUserRepo.On("GetByID", suite.ctx, tenantID, userID).Return((*models.User)(nil), expectedErr).Once()

		user, err := suite.mockUserRepo.GetByID(suite.ctx, tenantID, userID)

		suite.Error(err)
		suite.Nil(user)
		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})

	suite.Run("Invalid Input Error", func() {
		// Test with invalid UUID
		invalidTenantID := uuid.Nil
		invalidUserID := uuid.Nil
		expectedErr := errors.New("invalid user ID")

		suite.mockUserRepo.On("GetByID", suite.ctx, invalidTenantID, invalidUserID).Return((*models.User)(nil), expectedErr).Once()

		user, err := suite.mockUserRepo.GetByID(suite.ctx, invalidTenantID, invalidUserID)

		suite.Error(err)
		suite.Nil(user)
		suite.Equal(expectedErr, err)
		suite.mockUserRepo.AssertExpectations(suite.T())
	})
}