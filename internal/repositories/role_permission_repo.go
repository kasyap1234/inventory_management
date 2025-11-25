package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RolePermissionRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, rolePermission *models.RolePermission) error
	Delete(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error
	ListByRole(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.RolePermission, error)
	ListByPermission(ctx context.Context, permissionID uuid.UUID, limit, offset int) ([]*models.RolePermission, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.RolePermission, error)
	GetPermissionsByRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.Permission, error)
	RemoveAllPermissionsFromRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error
	AssignPermissionToRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error
	// GetAllUserPermissions fetches all permissions for a user in a single optimized query
	// This traverses the role hierarchy and returns all inherited permissions
	GetAllUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.Permission, error)
}

type rolePermissionRepo struct {
	db *pgxpool.Pool
}

func NewRolePermissionRepo(db *pgxpool.Pool) RolePermissionRepository {
	return &rolePermissionRepo{db: db}
}

func (r *rolePermissionRepo) Create(ctx context.Context, tenantID uuid.UUID, rolePermission *models.RolePermission) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		SELECT $1, $2, NOW()
		WHERE EXISTS (SELECT 1 FROM roles WHERE id = $1 AND tenant_id = $3)
		AND EXISTS (SELECT 1 FROM permissions WHERE id = $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, rolePermission.RoleID, rolePermission.PermissionID, tenantID)
	return err
}

func (r *rolePermissionRepo) Delete(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error {
	query := `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
		AND EXISTS (SELECT 1 FROM roles WHERE id = $1 AND tenant_id = $3)
		AND EXISTS (SELECT 1 FROM permissions WHERE id = $2)
	`
	_, err := r.db.Exec(ctx, query, roleID, permissionID, tenantID)
	return err
}

func (r *rolePermissionRepo) ListByRole(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.RolePermission, error) {
	query := `
		SELECT rp.id, rp.role_id, rp.permission_id, rp.created_at
		FROM role_permissions rp
		JOIN roles ro ON rp.role_id = ro.id
		WHERE ro.tenant_id = $1 AND rp.role_id = $2
		ORDER BY rp.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rolePermissions []*models.RolePermission
	for rows.Next() {
		rolePermission := &models.RolePermission{}
		if err := rows.Scan(&rolePermission.ID, &rolePermission.RoleID, &rolePermission.PermissionID, &rolePermission.CreatedAt); err != nil {
			return nil, err
		}
		rolePermissions = append(rolePermissions, rolePermission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (r *rolePermissionRepo) ListByPermission(ctx context.Context, permissionID uuid.UUID, limit, offset int) ([]*models.RolePermission, error) {
	query := `
		SELECT rp.id, rp.role_id, rp.permission_id, rp.created_at
		FROM role_permissions rp
		WHERE rp.permission_id = $1
		ORDER BY rp.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, permissionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rolePermissions []*models.RolePermission
	for rows.Next() {
		rolePermission := &models.RolePermission{}
		if err := rows.Scan(&rolePermission.ID, &rolePermission.RoleID, &rolePermission.PermissionID, &rolePermission.CreatedAt); err != nil {
			return nil, err
		}
		rolePermissions = append(rolePermissions, rolePermission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (r *rolePermissionRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.RolePermission, error) {
	query := `
		SELECT rp.id, rp.role_id, rp.permission_id, rp.created_at
		FROM role_permissions rp
		JOIN roles ro ON rp.role_id = ro.id
		WHERE ro.tenant_id = $1
		ORDER BY rp.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rolePermissions []*models.RolePermission
	for rows.Next() {
		rolePermission := &models.RolePermission{}
		if err := rows.Scan(&rolePermission.ID, &rolePermission.RoleID, &rolePermission.PermissionID, &rolePermission.CreatedAt); err != nil {
			return nil, err
		}
		rolePermissions = append(rolePermissions, rolePermission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (r *rolePermissionRepo) GetPermissionsByRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.Permission, error) {
	query := `
		SELECT p.id, p.name, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		AND EXISTS (SELECT 1 FROM roles r WHERE r.id = $1 AND r.tenant_id = $2)
	`
	rows, err := r.db.Query(ctx, query, roleID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		if err := rows.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *rolePermissionRepo) RemoveAllPermissionsFromRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error {
	query := `
		DELETE FROM role_permissions
		WHERE role_id = $1
		AND EXISTS (SELECT 1 FROM roles r WHERE r.id = $1 AND r.tenant_id = $2)
	`
	_, err := r.db.Exec(ctx, query, roleID, tenantID)
	return err
}

func (r *rolePermissionRepo) AssignPermissionToRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		SELECT $1, $2, NOW()
		WHERE EXISTS (SELECT 1 FROM roles r WHERE r.id = $1 AND r.tenant_id = $3)
		AND EXISTS (SELECT 1 FROM permissions p WHERE p.id = $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, roleID, permissionID, tenantID)
	return err
}

func (r *rolePermissionRepo) RemovePermissionFromRole(ctx context.Context, tenantID uuid.UUID, roleID, permissionID uuid.UUID) error {
	query := `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
		AND EXISTS (SELECT 1 FROM roles r WHERE r.id = $1 AND r.tenant_id = $3)
		AND EXISTS (SELECT 1 FROM permissions p WHERE p.id = $2)
	`
	_, err := r.db.Exec(ctx, query, roleID, permissionID, tenantID)
	return err
}

// GetAllUserPermissions fetches all permissions for a user in a single optimized query.
// This eliminates the N+1 query problem by using a recursive CTE to traverse role hierarchy
// and joining all tables in a single query.
func (r *rolePermissionRepo) GetAllUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]*models.Permission, error) {
	// Use a recursive CTE to traverse the role hierarchy and get all permissions
	query := `
		WITH RECURSIVE role_hierarchy AS (
			-- Base case: get user's directly assigned roles
			SELECT r.id, r.name, r.parent_role_id, r.priority, 0 as depth
			FROM roles r
			INNER JOIN user_roles ur ON r.id = ur.role_id
			WHERE ur.user_id = $1 
			  AND ur.tenant_id = $2
			  AND r.tenant_id = $2
			  AND r.is_active = true
			  AND COALESCE(ur.is_active, true) = true
			
			UNION ALL
			
			-- Recursive case: get parent roles
			SELECT r.id, r.name, r.parent_role_id, r.priority, rh.depth + 1
			FROM roles r
			INNER JOIN role_hierarchy rh ON r.id = rh.parent_role_id
			WHERE r.tenant_id = $2
			  AND r.is_active = true
			  AND rh.depth < 10  -- Prevent infinite recursion (max 10 levels)
		)
		SELECT DISTINCT ON (p.name)
			p.id, 
			p.name, 
			p.resource,
			p.action,
			COALESCE(rp.conditions, p.conditions) as conditions,
			p.description, 
			p.created_at
		FROM role_hierarchy rh
		INNER JOIN role_permissions rp ON rh.id = rp.role_id
		INNER JOIN permissions p ON rp.permission_id = p.id
		ORDER BY p.name, rh.priority DESC, rh.depth ASC
	`
	rows, err := r.db.Query(ctx, query, userID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		var conditions []byte
		if err := rows.Scan(
			&permission.ID,
			&permission.Name,
			&permission.Resource,
			&permission.Action,
			&conditions,
			&permission.Description,
			&permission.CreatedAt,
		); err != nil {
			return nil, err
		}
		// Parse conditions JSON if present
		if len(conditions) > 0 {
			// Store raw conditions for later use
			permission.ConditionsRaw = conditions
		}
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}
