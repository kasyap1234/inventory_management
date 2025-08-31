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
	return rolePermissions, nil
}