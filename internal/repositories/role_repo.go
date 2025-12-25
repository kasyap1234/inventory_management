package repositories

import (
	"context"
	"fmt"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoleRepository interface for role operations
type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	Update(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.Role, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*models.Role, error)
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	AssignUserToRole(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error
	RemoveUserFromRole(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*models.Role, error)
	GetRoleUsers(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.User, error)
	// GetUserMaxPriority returns the highest priority role for a user (for delegation checks)
	GetUserMaxPriority(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (int, error)
}

// roleRepo implements RoleRepository
type roleRepo struct {
	db *pgxpool.Pool
}

// NewRoleRepo creates a new role repository
func NewRoleRepo(db *pgxpool.Pool) RoleRepository {
	return &roleRepo{db: db}
}

// Create creates a new role
func (r *roleRepo) Create(ctx context.Context, role *models.Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}

	query := `
		INSERT INTO roles (id, tenant_id, name, description, is_active, is_system_role, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`

	_, err := r.db.Exec(ctx, query,
		role.ID,
		role.TenantID,
		role.Name,
		role.Description,
		role.IsActive,
		role.IsSystemRole,
		role.Priority,
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// Update updates an existing role
func (r *roleRepo) Update(ctx context.Context, role *models.Role) error {
	query := `
		UPDATE roles SET
			name = $3,
			description = $4,
			is_active = $5,
			updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.Exec(ctx, query,
		role.TenantID,
		role.ID,
		role.Name,
		role.Description,
		role.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// GetByID retrieves a role by ID
func (r *roleRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, is_active, COALESCE(is_system_role, false), COALESCE(priority, 0), created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND id = $2`

	var role models.Role
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&role.ID,
		&role.TenantID,
		&role.Name,
		&role.Description,
		&role.IsActive,
		&role.IsSystemRole,
		&role.Priority,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// GetByName retrieves a role by name
func (r *roleRepo) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, is_active, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND name = $2`

	var role models.Role
	err := r.db.QueryRow(ctx, query, tenantID, name).Scan(
		&role.ID,
		&role.TenantID,
		&role.Name,
		&role.Description,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// List retrieves all roles for a tenant
func (r *roleRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*models.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, is_active, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(
			&role.ID,
			&role.TenantID,
			&role.Name,
			&role.Description,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		roles = append(roles, &role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// Delete deletes a role
func (r *roleRepo) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	// First check if this is a system role (cannot be deleted)
	var isSystemRole bool
	systemCheckQuery := `SELECT COALESCE(is_system_role, false) FROM roles WHERE tenant_id = $1 AND id = $2`
	err := r.db.QueryRow(ctx, systemCheckQuery, tenantID, id).Scan(&isSystemRole)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("role not found")
		}
		return fmt.Errorf("failed to check role: %w", err)
	}
	if isSystemRole {
		return fmt.Errorf("cannot delete system role")
	}

	// Check if role has any users assigned
	var userCount int
	countQuery := `SELECT COUNT(*) FROM user_roles WHERE role_id = $1`
	err = r.db.QueryRow(ctx, countQuery, id).Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if userCount > 0 {
		return fmt.Errorf("cannot delete role: %d users are assigned to this role", userCount)
	}

	// Delete role permissions first
	_, err = r.db.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	// Delete the role
	query := `DELETE FROM roles WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// AssignUserToRole assigns a user to a role
func (r *roleRepo) AssignUserToRole(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error {
	query := `
		INSERT INTO user_roles (user_id, tenant_id, role_id, is_active, assigned_at)
		VALUES ($1, $2, $3, true, NOW())
		ON CONFLICT (user_id, tenant_id, role_id) 
		DO UPDATE SET is_active = true, assigned_at = NOW()`

	_, err := r.db.Exec(ctx, query, userID, tenantID, roleID)
	if err != nil {
		return fmt.Errorf("failed to assign user to role: %w", err)
	}

	return nil
}

// RemoveUserFromRole removes a user from a role
func (r *roleRepo) RemoveUserFromRole(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error {
	query := `
		UPDATE user_roles 
		SET is_active = false 
		WHERE user_id = $1 AND tenant_id = $2 AND role_id = $3`

	result, err := r.db.Exec(ctx, query, userID, tenantID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove user from role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user role assignment not found")
	}

	return nil
}

// GetUserRoles retrieves all roles for a user
func (r *roleRepo) GetUserRoles(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*models.Role, error) {
	query := `
		SELECT r.id, r.tenant_id, r.name, r.description, r.is_active, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND ur.tenant_id = $2 AND ur.is_active = true
		ORDER BY r.name ASC`

	rows, err := r.db.Query(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(
			&role.ID,
			&role.TenantID,
			&role.Name,
			&role.Description,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user role: %w", err)
		}

		roles = append(roles, &role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user roles: %w", err)
	}

	return roles, nil
}

// GetRoleUsers retrieves all users assigned to a role
func (r *roleRepo) GetRoleUsers(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT u.id, u.tenant_id, u.email, u.first_name, u.last_name, 
			   u.status, u.created_at, u.updated_at
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		WHERE ur.role_id = $1 AND ur.tenant_id = $2 AND ur.is_active = true
		ORDER BY u.first_name, u.last_name ASC`

	rows, err := r.db.Query(ctx, query, roleID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.TenantID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan role user: %w", err)
		}

		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating role users: %w", err)
	}

	return users, nil
}

// GetUserMaxPriority returns the highest priority role for a user (for delegation checks)
func (r *roleRepo) GetUserMaxPriority(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(MAX(r.priority), 0)
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND ur.tenant_id = $2 AND ur.is_active = true AND r.is_active = true`

	var maxPriority int
	err := r.db.QueryRow(ctx, query, userID, tenantID).Scan(&maxPriority)
	if err != nil {
		return 0, fmt.Errorf("failed to get user max priority: %w", err)
	}

	return maxPriority, nil
}
