package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRoleRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, userRole *models.UserRole) error
	Delete(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error
	ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.UserRole, error)
	ListByRole(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.UserRole, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.UserRole, error)
}

type userRoleRepo struct {
	db *pgxpool.Pool
}

func NewUserRoleRepo(db *pgxpool.Pool) UserRoleRepository {
	return &userRoleRepo{db: db}
}

func (r *userRoleRepo) Create(ctx context.Context, tenantID uuid.UUID, userRole *models.UserRole) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, created_at)
		SELECT $1, $2, NOW()
		WHERE EXISTS (SELECT 1 FROM users WHERE id = $1 AND tenant_id = $3)
		AND EXISTS (SELECT 1 FROM roles WHERE id = $2 AND tenant_id = $3)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, userRole.UserID, userRole.RoleID, tenantID)
	return err
}

func (r *userRoleRepo) Delete(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error {
	query := `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2
		AND EXISTS (SELECT 1 FROM users WHERE id = $1 AND tenant_id = $3)
		AND EXISTS (SELECT 1 FROM roles WHERE id = $2 AND tenant_id = $3)
	`
	_, err := r.db.Exec(ctx, query, userID, roleID, tenantID)
	return err
}

func (r *userRoleRepo) ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.UserRole, error) {
	query := `
		SELECT ur.id, ur.user_id, ur.role_id, ur.created_at
		FROM user_roles ur
		JOIN users u ON ur.user_id = u.id
		WHERE u.tenant_id = $1 AND ur.user_id = $2
		ORDER BY ur.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRoles []*models.UserRole
	for rows.Next() {
		userRole := &models.UserRole{}
		if err := rows.Scan(&userRole.ID, &userRole.UserID, &userRole.RoleID, &userRole.CreatedAt); err != nil {
			return nil, err
		}
		userRoles = append(userRoles, userRole)
	}
	return userRoles, nil
}

func (r *userRoleRepo) ListByRole(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.UserRole, error) {
	query := `
		SELECT ur.id, ur.user_id, ur.role_id, ur.created_at
		FROM user_roles ur
		JOIN roles ro ON ur.role_id = ro.id
		WHERE ro.tenant_id = $1 AND ur.role_id = $2
		ORDER BY ur.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRoles []*models.UserRole
	for rows.Next() {
		userRole := &models.UserRole{}
		if err := rows.Scan(&userRole.ID, &userRole.UserID, &userRole.RoleID, &userRole.CreatedAt); err != nil {
			return nil, err
		}
		userRoles = append(userRoles, userRole)
	}
	return userRoles, nil
}

func (r *userRoleRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.UserRole, error) {
	query := `
		SELECT ur.id, ur.user_id, ur.role_id, ur.created_at
		FROM user_roles ur
		JOIN users u ON ur.user_id = u.id
		JOIN roles ro ON ur.role_id = ro.id
		WHERE u.tenant_id = $1 AND ro.tenant_id = $1
		ORDER BY ur.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRoles []*models.UserRole
	for rows.Next() {
		userRole := &models.UserRole{}
		if err := rows.Scan(&userRole.ID, &userRole.UserID, &userRole.RoleID, &userRole.CreatedAt); err != nil {
			return nil, err
		}
		userRoles = append(userRoles, userRole)
	}
	return userRoles, nil
}