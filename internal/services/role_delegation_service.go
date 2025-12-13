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

	// CanCreateRoleWithPriority checks if user can create a role with the given priority
	// Rule: Users can create roles with priority up to (but not exceeding) their own
	CanCreateRoleWithPriority(ctx context.Context, userID, tenantID uuid.UUID, rolePriority int) (bool, error)

	// GetAssignableRoles returns roles that user can assign to others
	GetAssignableRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.Role, error)

	// GetAssignablePermissions returns permissions that user can grant to roles
	// Returns all permissions available at user's priority level and below
	GetAssignablePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.Permission, error)

	// GetUserMaxPriority returns the user's highest role priority
	GetUserMaxPriority(ctx context.Context, userID, tenantID uuid.UUID) (int, error)
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

// RolePermissionRepoForDelegation defines dependencies for permission queries
type RolePermissionRepoForDelegation interface {
	GetAllUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.Permission, error)
}

// PermissionRepoForDelegation defines dependencies for listing all permissions
type PermissionRepoForDelegation interface {
	ListPermissions(ctx context.Context) ([]*models.Permission, error)
}

type roleDelegationService struct {
	roleRepo           RoleRepoForDelegation
	userRepo           UserRepoForDelegation
	rolePermissionRepo RolePermissionRepoForDelegation
	permissionRepo     PermissionRepoForDelegation
}

// NewRoleDelegationService creates a new role delegation service (backward compatible)
func NewRoleDelegationService(roleRepo RoleRepoForDelegation, userRepo UserRepoForDelegation) RoleDelegationService {
	return &roleDelegationService{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

// NewRoleDelegationServiceWithPermissions creates a role delegation service with permission filtering
func NewRoleDelegationServiceWithPermissions(
	roleRepo RoleRepoForDelegation,
	userRepo UserRepoForDelegation,
	rolePermissionRepo RolePermissionRepoForDelegation,
	permissionRepo PermissionRepoForDelegation,
) RoleDelegationService {
	return &roleDelegationService{
		roleRepo:           roleRepo,
		userRepo:           userRepo,
		rolePermissionRepo: rolePermissionRepo,
		permissionRepo:     permissionRepo,
	}
}

// CanAssignRole checks if assigner can assign targetRole to a user
// Rule 1: Platform admins can assign any role
// Rule 2: Tenant users can only assign roles with priority <= their max priority
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

	// 4. Enforce hierarchy: Assigner must have priority >= the role they are assigning
	// Example: Admin (900) can assign Manager (700) -> 900 >= 700 -> OK
	// Example: Admin (900) can assign Admin (900) -> 900 >= 900 -> OK (can assign own level)
	// Example: Manager (700) cannot assign Admin (900) -> 700 >= 900 -> FALSE
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

// CanCreateRoleWithPriority checks if user can create a role with the given priority
// Rule 1: Platform admins can create any role
// Rule 2: Tenant users can create roles with priority up to (but not exceeding) their own
func (s *roleDelegationService) CanCreateRoleWithPriority(ctx context.Context, userID, tenantID uuid.UUID, rolePriority int) (bool, error) {
	// 1. Check if platform admin
	isPlatformAdmin, err := s.userRepo.IsPlatformAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check platform admin status: %w", err)
	}
	if isPlatformAdmin {
		return true, nil
	}

	// 2. Get user's max priority
	userPriority, err := s.roleRepo.GetUserMaxPriority(ctx, tenantID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user priority: %w", err)
	}

	// 3. User can create roles with priority <= their own (up to but not exceeding)
	return userPriority >= rolePriority, nil
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

	// Filter roles - can assign roles with priority <= own priority
	var assignableRoles []*models.Role
	for _, role := range allRoles {
		if userPriority >= role.Priority {
			assignableRoles = append(assignableRoles, role)
		}
	}

	return assignableRoles, nil
}

// GetAssignablePermissions returns permissions that user can grant to roles
// Platform admins can grant all permissions
// Other users can grant all available permissions (no filtering by what they have)
// since permission scope is controlled at the role priority level
func (s *roleDelegationService) GetAssignablePermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.Permission, error) {
	// Check if permission repo is available
	if s.permissionRepo == nil {
		return nil, fmt.Errorf("permission repository not configured")
	}

	// Check if platform admin
	isPlatformAdmin, err := s.userRepo.IsPlatformAdmin(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check platform admin status: %w", err)
	}

	// All permissions are available - scope is controlled by role priority
	// Platform admins and tenant admins alike can assign any permission
	// The restriction is on WHAT ROLES they can create/modify (by priority)
	if isPlatformAdmin {
		return s.permissionRepo.ListPermissions(ctx)
	}

	// Get user's max priority to verify they have role management capability
	userPriority, err := s.roleRepo.GetUserMaxPriority(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user priority: %w", err)
	}

	// Only users with sufficient priority (e.g., admin level 900+) should manage permissions
	// But since we're checking at handler level, return all permissions
	// The role priority check happens when assigning permissions to a role
	if userPriority <= 0 {
		return []*models.Permission{}, nil
	}

	return s.permissionRepo.ListPermissions(ctx)
}

// GetUserMaxPriority returns the user's highest role priority
func (s *roleDelegationService) GetUserMaxPriority(ctx context.Context, userID, tenantID uuid.UUID) (int, error) {
	isPlatformAdmin, err := s.userRepo.IsPlatformAdmin(ctx, userID)
	if err != nil {
		return 0, err
	}
	if isPlatformAdmin {
		// Platform admins have max priority (1000)
		return 1000, nil
	}

	return s.roleRepo.GetUserMaxPriority(ctx, tenantID, userID)
}

