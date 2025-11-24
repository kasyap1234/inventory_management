package services

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// RoleManagementService interface for role management operations
type RoleManagementService interface {
	// Role CRUD operations
	CreateRole(ctx context.Context, tenantID uuid.UUID, role *models.Role) error
	UpdateRole(ctx context.Context, tenantID uuid.UUID, role *models.Role) error
	GetRole(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.Role, error)
	ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*models.Role, error)
	DeleteRole(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error

	// Permission management
	AssignPermissionsToRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, permissions []uuid.UUID) error
	RemovePermissionsFromRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, permissions []uuid.UUID) error
	GetRolePermissions(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.Permission, error)
	ListAvailablePermissions(ctx context.Context) ([]*models.Permission, error)

	// User role assignment
	AssignUserToRole(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error
	RemoveUserFromRole(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*models.Role, error)
	GetRoleUsers(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.User, error)

	// Bulk operations
	BulkAssignUsersToRole(ctx context.Context, tenantID uuid.UUID, userIDs []uuid.UUID, roleID uuid.UUID) error
	BulkRemoveUsersFromRole(ctx context.Context, tenantID uuid.UUID, userIDs []uuid.UUID, roleID uuid.UUID) error

	// Role validation and conflict detection
	ValidateRoleAssignment(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error
	DetectRoleConflicts(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, newRoleID uuid.UUID) ([]string, error)
}

// PermissionService implements the middleware.PermissionService interface
type PermissionService interface{ // implements middleware contract without importing it here
    HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string, context map[string]interface{}) (bool, error)
    GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]models.RBACPermission, error)
    CheckResourceAccess(ctx context.Context, userID, tenantID uuid.UUID, resource, action string, resourceID *uuid.UUID) (bool, error)
}

// roleManagementService implements RoleManagementService and PermissionService
type roleManagementService struct {
	roleRepo       repositories.RoleRepository
	permissionRepo repositories.PermissionRepository
	logger         *common.StructuredLogger
}

// NewRoleManagementService creates a new role management service
func NewRoleManagementService(
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
	logger *common.StructuredLogger,
) RoleManagementService {
	return &roleManagementService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		logger:         logger,
	}
}

// NewPermissionService creates a new permission service for middleware
func NewPermissionService(
	permissionRepo repositories.PermissionRepository,
	logger *common.StructuredLogger,
) PermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
		logger:         logger,
	}
}

// permissionService implements middleware.PermissionService
type permissionService struct {
	permissionRepo repositories.PermissionRepository
	logger         *common.StructuredLogger
}

// HasPermission implements middleware.PermissionService
func (s *permissionService) HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string, context map[string]interface{}) (bool, error) {
	return s.permissionRepo.HasPermission(ctx, userID, tenantID, permission)
}

// GetUserPermissions implements middleware.PermissionService
func (s *permissionService) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]models.RBACPermission, error) {
	return s.permissionRepo.GetUserPermissions(ctx, userID, tenantID)
}

// CheckResourceAccess implements middleware.PermissionService
func (s *permissionService) CheckResourceAccess(ctx context.Context, userID, tenantID uuid.UUID, resource, action string, resourceID *uuid.UUID) (bool, error) {
	return s.permissionRepo.CheckResourceAccess(ctx, userID, tenantID, resource, action, resourceID)
}

// Role CRUD operations

// CreateRole creates a new role
func (s *roleManagementService) CreateRole(ctx context.Context, tenantID uuid.UUID, role *models.Role) error {
	role.TenantID = tenantID
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	// Validate role
	if err := s.validateRole(role); err != nil {
		return common.CreateValidationError("create_role", map[string]interface{}{
			"validation": err.Error(),
		})
	}

	// Create role
	if err := s.roleRepo.Create(ctx, role); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to create role", err, map[string]interface{}{
			"tenant_id": tenantID,
			"role_name": role.Name,
		})
		return common.CreateDatabaseError("create_role", err)
	}

	s.logger.InfoWithContext(ctx, "Role created", map[string]interface{}{
		"role_id":   role.ID,
		"tenant_id": tenantID,
		"role_name": role.Name,
	})

	// Audit log
	common.AuditCreate(ctx, "role", role.ID.String(), map[string]interface{}{
		"name":        role.Name,
		"description": role.Description,
		"is_active":   role.IsActive,
	})

	return nil
}

// UpdateRole updates an existing role
func (s *roleManagementService) UpdateRole(ctx context.Context, tenantID uuid.UUID, role *models.Role) error {
	// Get existing role for audit logging
	existing, err := s.roleRepo.GetByID(ctx, tenantID, role.ID)
	if err != nil {
		return common.CreateDatabaseError("update_role", err)
	}
	if existing == nil {
		return common.CreateDatabaseError("update_role", fmt.Errorf("role not found"))
	}

	role.TenantID = tenantID
	role.UpdatedAt = time.Now()

	// Validate role
	if err := s.validateRole(role); err != nil {
		return common.CreateValidationError("update_role", map[string]interface{}{
			"validation": err.Error(),
		})
	}

	// Update role
	if err := s.roleRepo.Update(ctx, role); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to update role", err, map[string]interface{}{
			"role_id":   role.ID,
			"tenant_id": tenantID,
		})
		return common.CreateDatabaseError("update_role", err)
	}

	s.logger.InfoWithContext(ctx, "Role updated", map[string]interface{}{
		"role_id":   role.ID,
		"tenant_id": tenantID,
		"role_name": role.Name,
	})

	// Audit log
	oldValues := map[string]interface{}{
		"name":        existing.Name,
		"description": existing.Description,
		"is_active":   existing.IsActive,
	}
	newValues := map[string]interface{}{
		"name":        role.Name,
		"description": role.Description,
		"is_active":   role.IsActive,
	}
	common.AuditUpdate(ctx, "role", role.ID.String(), oldValues, newValues)

	return nil
}

// GetRole retrieves a role by ID
func (s *roleManagementService) GetRole(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, common.CreateDatabaseError("get_role", err)
	}

	return role, nil
}

// ListRoles retrieves all roles for a tenant
func (s *roleManagementService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*models.Role, error) {
	roles, err := s.roleRepo.List(ctx, tenantID)
	if err != nil {
		return nil, common.CreateDatabaseError("list_roles", err)
	}

	return roles, nil
}

// DeleteRole deletes a role
func (s *roleManagementService) DeleteRole(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	// Get existing role for audit logging
	existing, err := s.roleRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return common.CreateDatabaseError("delete_role", err)
	}
	if existing == nil {
		return common.CreateDatabaseError("delete_role", fmt.Errorf("role not found"))
	}

	// Delete role
	if err := s.roleRepo.Delete(ctx, tenantID, id); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to delete role", err, map[string]interface{}{
			"role_id":   id,
			"tenant_id": tenantID,
		})
		return common.CreateDatabaseError("delete_role", err)
	}

	s.logger.InfoWithContext(ctx, "Role deleted", map[string]interface{}{
		"role_id":   id,
		"tenant_id": tenantID,
		"role_name": existing.Name,
	})

	// Audit log
	common.AuditDelete(ctx, "role", id.String(), map[string]interface{}{
		"name":        existing.Name,
		"description": existing.Description,
	})

	return nil
}

// Permission management

// AssignPermissionsToRole assigns multiple permissions to a role
func (s *roleManagementService) AssignPermissionsToRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, permissions []uuid.UUID) error {
	// Verify role exists
	_, err := s.roleRepo.GetByID(ctx, tenantID, roleID)
	if err != nil {
		return common.CreateDatabaseError("assign_permissions", err)
	}

	// Assign each permission
	for _, permissionID := range permissions {
		if err := s.permissionRepo.AssignPermissionToRole(ctx, tenantID, roleID, permissionID, nil); err != nil {
			s.logger.ErrorWithContext(ctx, "Failed to assign permission to role", err, map[string]interface{}{
				"role_id":       roleID,
				"permission_id": permissionID,
				"tenant_id":     tenantID,
			})
			return common.CreateDatabaseError("assign_permissions", err)
		}
	}

	s.logger.InfoWithContext(ctx, "Permissions assigned to role", map[string]interface{}{
		"role_id":          roleID,
		"tenant_id":        tenantID,
		"permission_count": len(permissions),
	})

	// Audit log
	common.AuditUpdate(ctx, "role_permissions", roleID.String(), nil, map[string]interface{}{
		"assigned_permissions": permissions,
	})

	return nil
}

// RemovePermissionsFromRole removes multiple permissions from a role
func (s *roleManagementService) RemovePermissionsFromRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, permissions []uuid.UUID) error {
	// Remove each permission
	for _, permissionID := range permissions {
		if err := s.permissionRepo.RemovePermissionFromRole(ctx, tenantID, roleID, permissionID); err != nil {
			s.logger.ErrorWithContext(ctx, "Failed to remove permission from role", err, map[string]interface{}{
				"role_id":       roleID,
				"permission_id": permissionID,
				"tenant_id":     tenantID,
			})
			// Continue with other permissions even if one fails
		}
	}

	s.logger.InfoWithContext(ctx, "Permissions removed from role", map[string]interface{}{
		"role_id":          roleID,
		"tenant_id":        tenantID,
		"permission_count": len(permissions),
	})

	// Audit log
	common.AuditUpdate(ctx, "role_permissions", roleID.String(), nil, map[string]interface{}{
		"removed_permissions": permissions,
	})

	return nil
}

// GetRolePermissions retrieves all permissions for a role
func (s *roleManagementService) GetRolePermissions(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.Permission, error) {
	permissions, err := s.permissionRepo.GetRolePermissions(ctx, tenantID, roleID)
	if err != nil {
		return nil, common.CreateDatabaseError("get_role_permissions", err)
	}

	return permissions, nil
}

// ListAvailablePermissions retrieves all available permissions
func (s *roleManagementService) ListAvailablePermissions(ctx context.Context) ([]*models.Permission, error) {
	permissions, err := s.permissionRepo.ListPermissions(ctx)
	if err != nil {
		return nil, common.CreateDatabaseError("list_permissions", err)
	}

	return permissions, nil
}

// User role assignment

// AssignUserToRole assigns a user to a role
func (s *roleManagementService) AssignUserToRole(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error {
	// Validate role assignment
	if err := s.ValidateRoleAssignment(ctx, tenantID, userID, roleID); err != nil {
		return err
	}

	// Assign user to role
	if err := s.roleRepo.AssignUserToRole(ctx, tenantID, userID, roleID); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to assign user to role", err, map[string]interface{}{
			"user_id":   userID,
			"role_id":   roleID,
			"tenant_id": tenantID,
		})
		return common.CreateDatabaseError("assign_user_to_role", err)
	}

	s.logger.InfoWithContext(ctx, "User assigned to role", map[string]interface{}{
		"user_id":   userID,
		"role_id":   roleID,
		"tenant_id": tenantID,
	})

	// Audit log
	common.AuditCreate(ctx, "user_role", "", map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
		"action":  "assigned",
	})

	return nil
}

// RemoveUserFromRole removes a user from a role
func (s *roleManagementService) RemoveUserFromRole(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error {
	// Remove user from role
	if err := s.roleRepo.RemoveUserFromRole(ctx, tenantID, userID, roleID); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to remove user from role", err, map[string]interface{}{
			"user_id":   userID,
			"role_id":   roleID,
			"tenant_id": tenantID,
		})
		return common.CreateDatabaseError("remove_user_from_role", err)
	}

	s.logger.InfoWithContext(ctx, "User removed from role", map[string]interface{}{
		"user_id":   userID,
		"role_id":   roleID,
		"tenant_id": tenantID,
	})

	// Audit log
	common.AuditCreate(ctx, "user_role", "", map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
		"action":  "removed",
	})

	return nil
}

// GetUserRoles retrieves all roles for a user
func (s *roleManagementService) GetUserRoles(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*models.Role, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return nil, common.CreateDatabaseError("get_user_roles", err)
	}

	return roles, nil
}

// GetRoleUsers retrieves all users assigned to a role
func (s *roleManagementService) GetRoleUsers(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.User, error) {
	users, err := s.roleRepo.GetRoleUsers(ctx, tenantID, roleID)
	if err != nil {
		return nil, common.CreateDatabaseError("get_role_users", err)
	}

	return users, nil
}

// Bulk operations

// BulkAssignUsersToRole assigns multiple users to a role
func (s *roleManagementService) BulkAssignUsersToRole(ctx context.Context, tenantID uuid.UUID, userIDs []uuid.UUID, roleID uuid.UUID) error {
	successCount := 0
	errorCount := 0

	for _, userID := range userIDs {
		if err := s.AssignUserToRole(ctx, tenantID, userID, roleID); err != nil {
			s.logger.ErrorWithContext(ctx, "Failed to assign user to role in bulk operation", err, map[string]interface{}{
				"user_id":   userID,
				"role_id":   roleID,
				"tenant_id": tenantID,
			})
			errorCount++
		} else {
			successCount++
		}
	}

	s.logger.InfoWithContext(ctx, "Bulk user role assignment completed", map[string]interface{}{
		"role_id":       roleID,
		"tenant_id":     tenantID,
		"success_count": successCount,
		"error_count":   errorCount,
		"total_count":   len(userIDs),
	})

	if errorCount > 0 {
		return fmt.Errorf("bulk assignment completed with %d errors out of %d operations", errorCount, len(userIDs))
	}

	return nil
}

// BulkRemoveUsersFromRole removes multiple users from a role
func (s *roleManagementService) BulkRemoveUsersFromRole(ctx context.Context, tenantID uuid.UUID, userIDs []uuid.UUID, roleID uuid.UUID) error {
	successCount := 0
	errorCount := 0

	for _, userID := range userIDs {
		if err := s.RemoveUserFromRole(ctx, tenantID, userID, roleID); err != nil {
			s.logger.ErrorWithContext(ctx, "Failed to remove user from role in bulk operation", err, map[string]interface{}{
				"user_id":   userID,
				"role_id":   roleID,
				"tenant_id": tenantID,
			})
			errorCount++
		} else {
			successCount++
		}
	}

	s.logger.InfoWithContext(ctx, "Bulk user role removal completed", map[string]interface{}{
		"role_id":       roleID,
		"tenant_id":     tenantID,
		"success_count": successCount,
		"error_count":   errorCount,
		"total_count":   len(userIDs),
	})

	if errorCount > 0 {
		return fmt.Errorf("bulk removal completed with %d errors out of %d operations", errorCount, len(userIDs))
	}

	return nil
}

// Validation and conflict detection

// ValidateRoleAssignment validates if a user can be assigned to a role
func (s *roleManagementService) ValidateRoleAssignment(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error {
	// Check if role exists and is active
	role, err := s.roleRepo.GetByID(ctx, tenantID, roleID)
	if err != nil {
		return common.CreateValidationError("validate_role_assignment", map[string]interface{}{
			"role_id": "Role not found",
		})
	}

	if !role.IsActive {
		return common.CreateValidationError("validate_role_assignment", map[string]interface{}{
			"role_id": "Role is not active",
		})
	}

	// Check for role conflicts
	conflicts, err := s.DetectRoleConflicts(ctx, tenantID, userID, roleID)
	if err != nil {
		return err
	}

	if len(conflicts) > 0 {
		return common.CreateValidationError("validate_role_assignment", map[string]interface{}{
			"conflicts": conflicts,
		})
	}

	return nil
}

// DetectRoleConflicts detects potential conflicts when assigning a role to a user
func (s *roleManagementService) DetectRoleConflicts(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, newRoleID uuid.UUID) ([]string, error) {
	var conflicts []string

	// No hardcoded conflicts - allow any role combination
	// In a real-world scenario, this might be driven by a policy engine or configuration
	
	return conflicts, nil
}

// Helper methods

// validateRole validates a role
func (s *roleManagementService) validateRole(role *models.Role) error {
	if role.Name == "" {
		return fmt.Errorf("role name is required")
	}

	if len(role.Name) > 100 {
		return fmt.Errorf("role name cannot exceed 100 characters")
	}

	if role.Description != nil && len(*role.Description) > 500 {
		return fmt.Errorf("role description cannot exceed 500 characters")
	}

	return nil
}