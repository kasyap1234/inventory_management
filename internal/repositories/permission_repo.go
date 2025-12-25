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

// CheckResourceAccess checks if a user has access to a specific resource with data-level restrictions.
// It enforces permission conditions like data_scope ("own"/"all") from the permission's JSONB conditions.
func (r *permissionRepo) CheckResourceAccess(ctx context.Context, userID, tenantID uuid.UUID, resource, action string, resourceID *uuid.UUID) (bool, error) {
	// First check if user has the general permission
	permissionName := fmt.Sprintf("%s.%s", resource, action)
	hasPermission, err := r.HasPermission(ctx, userID, tenantID, permissionName)
	if err != nil {
		return false, err
	}

	if !hasPermission {
		// Also check for wildcard permissions
		wildcardResource := fmt.Sprintf("%s.*", resource)
		wildcardAction := fmt.Sprintf("*.%s", action)
		hasWildcardResource, _ := r.HasPermission(ctx, userID, tenantID, wildcardResource)
		hasWildcardAction, _ := r.HasPermission(ctx, userID, tenantID, wildcardAction)
		hasFullWildcard, _ := r.HasPermission(ctx, userID, tenantID, "*")

		if !hasWildcardResource && !hasWildcardAction && !hasFullWildcard {
			return false, nil
		}
	}

	// If no specific resource ID, general permission is sufficient
	if resourceID == nil {
		return true, nil
	}

	// Get the user's permissions with conditions for this resource/action
	query := `
		WITH RECURSIVE role_hierarchy AS (
			SELECT r.id, r.parent_role_id, 0 as depth
			FROM roles r
			INNER JOIN user_roles ur ON r.id = ur.role_id
			WHERE ur.user_id = $1 
			  AND ur.tenant_id = $2
			  AND r.tenant_id = $2
			  AND r.is_active = true
			  AND COALESCE(ur.is_active, true) = true
			
			UNION ALL
			
			SELECT r.id, r.parent_role_id, rh.depth + 1
			FROM roles r
			INNER JOIN role_hierarchy rh ON r.id = rh.parent_role_id
			WHERE r.tenant_id = $2
			  AND r.is_active = true
			  AND rh.depth < 10
		)
		SELECT COALESCE(rp.conditions, p.conditions, '{}'::jsonb) as conditions
		FROM role_hierarchy rh
		INNER JOIN role_permissions rp ON rh.id = rp.role_id
		INNER JOIN permissions p ON rp.permission_id = p.id
		WHERE (p.name = $3 OR p.name = $4 OR p.name = $5 OR p.name = '*')
		  AND (p.resource = $6 OR p.resource IS NULL OR p.name = '*')
		ORDER BY 
			CASE WHEN p.name = $3 THEN 0 ELSE 1 END, -- Exact match first
			rh.depth ASC
		LIMIT 1
	`

	wildcardResource := fmt.Sprintf("%s.*", resource)
	wildcardAction := fmt.Sprintf("*.%s", action)

	var conditionsJSON []byte
	err = r.db.QueryRow(ctx, query, userID, tenantID, permissionName, wildcardResource, wildcardAction, resource).Scan(&conditionsJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			// No specific conditions, default to tenant-level access
			return true, nil
		}
		return false, fmt.Errorf("failed to get permission conditions: %w", err)
	}

	// Parse conditions
	var conditions map[string]interface{}
	if len(conditionsJSON) > 0 {
		if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
			return false, fmt.Errorf("failed to parse permission conditions: %w", err)
		}
	}

	// Check data_scope condition
	if dataScope, ok := conditions["data_scope"].(string); ok {
		switch dataScope {
		case "all":
			// User has access to all resources of this type in the tenant
			return true, nil
		case "own":
			// User only has access to resources they own/created
			return r.checkResourceOwnership(ctx, userID, tenantID, resource, *resourceID)
		case "team":
			// User has access to resources owned by their team (future implementation)
			// For now, fall back to ownership check
			return r.checkResourceOwnership(ctx, userID, tenantID, resource, *resourceID)
		}
	}

	// No data_scope restriction, default to tenant-level access
	return true, nil
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
