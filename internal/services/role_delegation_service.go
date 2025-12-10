package services

import (
	"context"
	"fmt"

	"agromart2/internal/models"

	"github.com/google/uuid"
)

// RoleDelegationService enforces rules for role assignment and modification
// It prevents privilege escalation by ensuring users cannot assign roles higher than their own
type RoleDelegationService interface {
	// CanAssignRole checks if assigner can assign targetRole to a user
	CanAssignRole(ctx context.Context, assignerID, tenantID uuid.UUID, targetRoleID uuid.UUID) (bool, error)

	// CanModifyRole checks if user can modify (update/delete) a role
	CanModifyRole(ctx context.Context, userID, tenantID uuid.UUID, roleID uuid.UUID) (bool, error)

	// GetAssignableRoles returns roles that user can assign to others
	GetAssignableRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.Role, error)
}

// RoleRepoForDelegation defines dependencies needed from RoleRepo
type RoleRepoForDelegation interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Role, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*models.Role, error)
	GetUserMaxPriority(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (int, error)
}

// UserRepoForDelegation defines dependencies needed from UserRepo
type UserRepoForDelegation interface {
	IsPlatformAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

type roleDelegationService struct {
	roleRepo RoleRepoForDelegation
	userRepo UserRepoForDelegation
}

func NewRoleDelegationService(roleRepo RoleRepoForDelegation, userRepo UserRepoForDelegation) RoleDelegationService {
	return &roleDelegationService{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

// CanAssignRole checks if assigner can assign targetRole to a user
// Rule 1: Platform admins can assign any role
// Rule 2: Tenant users can only assign roles with priority < their max priority
func (s *roleDelegationService) CanAssignRole(ctx context.Context, assignerID, tenantID uuid.UUID, targetRoleID uuid.UUID) (bool, error) {
	// 1. Check if platform admin (bypass all checks)
	isPlatformAdmin, err := s.userRepo.IsPlatformAdmin(ctx, assignerID)
	if err != nil {
		return false, fmt.Errorf("failed to check platform admin status: %w", err)
	}
	if isPlatformAdmin {
		return true, nil
	}

	// 2. Get target role priority
	targetRole, err := s.roleRepo.GetByID(ctx, tenantID, targetRoleID)
	if err != nil {
		return false, fmt.Errorf("failed to get target role: %w", err)
	}

	// 3. Get assigner's max priority
	assignerPriority, err := s.roleRepo.GetUserMaxPriority(ctx, tenantID, assignerID)
	if err != nil {
		return false, fmt.Errorf("failed to get assigner priority: %w", err)
	}

	// 4. Enforce strict hierarchy: Assigner must have higher priority than the role they are assigning
	// Example: Admin (900) can assign Manager (700) -> 900 > 700 -> OK
	// Example: Manager (700) cannot assign Admin (900) -> 700 > 900 -> FALSE
	// Example: Manager (700) cannot assign Manager (700) -> 700 > 700 -> FALSE (prevents lateral movement without higher authority)
	return assignerPriority >= targetRole.Priority, nil
}

// CanModifyRole checks if user can modify (update/delete) a role
// Rule 1: Platform admins can modify any role
// Rule 2: System roles cannot be modified by tenant admins
// Rule 3: Tenant admins can only modify roles with priority < their max priority
func (s *roleDelegationService) CanModifyRole(ctx context.Context, userID, tenantID uuid.UUID, roleID uuid.UUID) (bool, error) {
	// 1. Check if platform admin
	isPlatformAdmin, err := s.userRepo.IsPlatformAdmin(ctx, userID)
	if err != nil {
		return false, err
	}
	if isPlatformAdmin {
		return true, nil
	}

	// 2. Get target role
	targetRole, err := s.roleRepo.GetByID(ctx, tenantID, roleID)
	if err != nil {
		return false, err
	}

	// 3. System roles are protected
	if targetRole.IsSystemRole {
		return false, nil
	}

	// 4. Check priority hierarchy
	userPriority, err := s.roleRepo.GetUserMaxPriority(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	return userPriority > targetRole.Priority, nil
}

// GetAssignableRoles returns list of roles the user is allowed to assign
func (s *roleDelegationService) GetAssignableRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.Role, error) {
	// Get all roles
	allRoles, err := s.roleRepo.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Check if platform admin
	isPlatformAdmin, err := s.userRepo.IsPlatformAdmin(ctx, userID)
	if err != nil {
		return nil, err
	}
	if isPlatformAdmin {
		return allRoles, nil
	}

	// Get user priority
	userPriority, err := s.roleRepo.GetUserMaxPriority(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Filter roles
	var assignableRoles []*models.Role
	for _, role := range allRoles {
		// Can only assign roles with equal or lower priority
		if userPriority >= role.Priority {
			assignableRoles = append(assignableRoles, role)
		}
	}

	return assignableRoles, nil
}
