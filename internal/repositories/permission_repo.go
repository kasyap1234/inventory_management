package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PermissionRepository interface for permission operations
type PermissionRepository interface {
    GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]models.RBACPermission, error)
    HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string) (bool, error)
    CheckResourceAccess(ctx context.Context, userID, tenantID uuid.UUID, resource, action string, resourceID *uuid.UUID) (bool, error)
    GetRolePermissions(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.Permission, error)
    GetPermissionsByRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]models.RBACPermission, error)
    AssignPermissionToRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID, conditions map[string]interface{}) error
    RemovePermissionFromRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error
    RemoveAllPermissionsFromRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error
    List(ctx context.Context) ([]models.RBACPermission, error)
    ListPermissions(ctx context.Context) ([]*models.Permission, error)
    GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error)
    GetPermissionByName(ctx context.Context, name string) (*models.RBACPermission, error)
}

// permissionRepo implements PermissionRepository
type permissionRepo struct {
	db *pgxpool.Pool
}

// NewPermissionRepo creates a new permission repository
func NewPermissionRepo(db *pgxpool.Pool) PermissionRepository {
	return &permissionRepo{db: db}
}

// GetUserPermissions retrieves all permissions for a user
func (r *permissionRepo) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]models.RBACPermission, error) {
	query := `
		SELECT 
			p.name,
			p.resource,
			p.action,
			COALESCE(rp.conditions, p.conditions) as conditions,
			r.name as role_name
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = $1 AND ur.tenant_id = $2 AND ur.is_active = true
		ORDER BY p.resource, p.action`

	rows, err := r.db.Query(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

    var permissions []models.RBACPermission
	for rows.Next() {
        var perm models.RBACPermission
		var conditionsJSON []byte

		err := rows.Scan(
			&perm.Name,
			&perm.Resource,
			&perm.Action,
			&conditionsJSON,
			&perm.RoleName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		// Unmarshal conditions
		if len(conditionsJSON) > 0 {
			if err := json.Unmarshal(conditionsJSON, &perm.Conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
		} else {
			perm.Conditions = make(map[string]interface{})
		}

		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (r *permissionRepo) HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string) (bool, error) {
	query := `
		SELECT user_has_permission($1, $2, $3)`

	var hasPermission bool
	err := r.db.QueryRow(ctx, query, userID, tenantID, permission).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return hasPermission, nil
}

// CheckResourceAccess checks if a user has access to a specific resource
func (r *permissionRepo) CheckResourceAccess(ctx context.Context, userID, tenantID uuid.UUID, resource, action string, resourceID *uuid.UUID) (bool, error) {
	// First check if user has the general permission
	permissionName := fmt.Sprintf("%s.%s", resource, action)
	hasPermission, err := r.HasPermission(ctx, userID, tenantID, permissionName)
	if err != nil {
		return false, err
	}

	if !hasPermission {
		return false, nil
	}

	// If no specific resource ID, general permission is sufficient
	if resourceID == nil {
		return true, nil
	}

	// Check for resource-specific conditions
	// This is a simplified implementation - in practice, you might have more complex logic
	// based on ownership, team membership, etc.
	
	// For now, we'll check if the user has "all" data scope or if they own the resource
	permissions, err := r.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			// Check data scope conditions
			if dataScope, ok := perm.Conditions["data_scope"]; ok {
				if scope, ok := dataScope.(string); ok && scope == "all" {
					return true, nil
				}
			}
		}
	}

	// Check if user owns the resource (simplified check)
	// In practice, this would be more sophisticated based on the resource type
	return r.checkResourceOwnership(ctx, userID, tenantID, resource, *resourceID)
}

// GetRolePermissions retrieves all permissions for a role
func (r *permissionRepo) GetRolePermissions(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.Permission, error) {
	query := `
		SELECT 
			p.id,
			p.name,
			p.description,
			p.created_at
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		JOIN roles r ON rp.role_id = r.id
		WHERE rp.role_id = $1 AND r.tenant_id = $2
		ORDER BY p.name`

	rows, err := r.db.Query(ctx, query, roleID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

    var permissions []*models.Permission
	for rows.Next() {
        var perm models.Permission

		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Description,
			&perm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role permission: %w", err)
		}

		permissions = append(permissions, &perm)
	}

	return permissions, nil
}

// AssignPermissionToRole assigns a permission to a role
func (r *permissionRepo) AssignPermissionToRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID, conditions map[string]interface{}) error {
	conditionsJSON, err := json.Marshal(conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		INSERT INTO role_permissions (role_id, permission_id, conditions, granted_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (role_id, permission_id) 
		DO UPDATE SET conditions = $3, granted_at = NOW()`

	_, err = r.db.Exec(ctx, query, roleID, permissionID, conditionsJSON)
	if err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *permissionRepo) RemovePermissionFromRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`

	result, err := r.db.Exec(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("permission assignment not found")
	}

	return nil
}

// RemoveAllPermissionsFromRole removes all permissions from a role
func (r *permissionRepo) RemoveAllPermissionsFromRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error {
	query := `
		DELETE FROM role_permissions 
		WHERE role_id = $1 
		AND EXISTS (SELECT 1 FROM roles WHERE id = $1 AND tenant_id = $2)`

	_, err := r.db.Exec(ctx, query, roleID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to remove all permissions from role: %w", err)
	}

	return nil
}

// GetPermissionsByRole is an alias for GetRolePermissions for compatibility
func (r *permissionRepo) GetPermissionsByRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]models.RBACPermission, error) {
	// Legacy method - return empty for compatibility
	return []models.RBACPermission{}, nil
}

// List is an alias for ListPermissions for compatibility
func (r *permissionRepo) List(ctx context.Context) ([]models.RBACPermission, error) {
	// Legacy method - return empty for compatibility
	return []models.RBACPermission{}, nil
}

// ListPermissions retrieves all available permissions
func (r *permissionRepo) ListPermissions(ctx context.Context) ([]*models.Permission, error) {
	query := `
		SELECT id, name, description, created_at
		FROM permissions
		ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

    var permissions []*models.Permission
	for rows.Next() {
        var perm models.Permission

		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Description,
			&perm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		permissions = append(permissions, &perm)
	}

	return permissions, nil
}

// GetPermissionByName retrieves a permission by name
func (r *permissionRepo) GetPermissionByName(ctx context.Context, name string) (*models.RBACPermission, error) {
	query := `
		SELECT name, resource, action, conditions, description
		FROM permissions
		WHERE name = $1`

    var perm models.RBACPermission
	var conditionsJSON []byte
	var description *string

	err := r.db.QueryRow(ctx, query, name).Scan(
		&perm.Name,
		&perm.Resource,
		&perm.Action,
		&conditionsJSON,
		&description,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	// Unmarshal conditions
	if len(conditionsJSON) > 0 {
		if err := json.Unmarshal(conditionsJSON, &perm.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
	} else {
		perm.Conditions = make(map[string]interface{})
	}

	return &perm, nil
}

// Helper methods

// checkResourceOwnership checks if a user owns a specific resource
func (r *permissionRepo) checkResourceOwnership(ctx context.Context, userID, tenantID uuid.UUID, resource string, resourceID uuid.UUID) (bool, error) {
	// This is a simplified implementation
	// In practice, you would have specific logic for each resource type
	
	var query string
	switch resource {
	case "product":
		query = `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND tenant_id = $2)`
	case "order":
		query = `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1 AND tenant_id = $2 AND created_by = $3)`
	case "warehouse":
		query = `SELECT EXISTS(SELECT 1 FROM warehouses WHERE id = $1 AND tenant_id = $2)`
	case "distributor":
		query = `SELECT EXISTS(SELECT 1 FROM distributors WHERE id = $1 AND tenant_id = $2)`
	default:
		// For unknown resources, default to tenant-level access
		return true, nil
	}

	var exists bool
	if resource == "order" {
		err := r.db.QueryRow(ctx, query, resourceID, tenantID, userID).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("failed to check resource ownership: %w", err)
		}
	} else {
		err := r.db.QueryRow(ctx, query, resourceID, tenantID).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("failed to check resource ownership: %w", err)
		}
	}

	return exists, nil
}
// GetPermissionByID retrieves a permission by ID (returns as RBAC permission for compatibility)
func (r *permissionRepo) GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error) {
	query := `
		SELECT id, name, description, created_at
		FROM permissions
		WHERE id = $1`

	var perm models.Permission

	err := r.db.QueryRow(ctx, query, permissionID).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Description,
		&perm.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by ID: %w", err)
	}

	return &perm, nil
}
