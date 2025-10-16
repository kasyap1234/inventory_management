package services

import (
	"context"

	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// UserService defines the interface for user-related business logic
type UserService interface {
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string) error
}

// userService implements UserService
type userService struct {
	userRepo repositories.UserRepository
}

// NewUserService creates a new user service instance
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// UpdateUserProfile updates the user's profile information
func (s *userService) UpdateUserProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string) error {
	// Get tenant ID for the user
	tenantID, err := s.userRepo.GetTenantIDByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Get the existing user
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	// Update the fields
	user.FirstName = firstName
	user.LastName = lastName

	// Save the updated user
	return s.userRepo.Update(ctx, user)
}