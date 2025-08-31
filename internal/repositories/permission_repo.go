package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission *models.Permission) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	GetByName(ctx context.Context, name string) (*models.Permission, error)
	Update(ctx context.Context, permission *models.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Permission, error)
}

type permissionRepo struct {
	db *pgxpool.Pool
}

func NewPermissionRepo(db *pgxpool.Pool) PermissionRepository {
	return &permissionRepo{db: db}
}

func (r *permissionRepo) Create(ctx context.Context, permission *models.Permission) error {
	query := `
		INSERT INTO permissions (id, name, description, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (name) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, permission.ID, permission.Name, permission.Description)
	return err
}

func (r *permissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	permission := &models.Permission{}
	query := `
		SELECT id, name, description, created_at
		FROM permissions
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (r *permissionRepo) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	permission := &models.Permission{}
	query := `
		SELECT id, name, description, created_at
		FROM permissions
		WHERE name = $1
	`
	err := r.db.QueryRow(ctx, query, name).Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (r *permissionRepo) Update(ctx context.Context, permission *models.Permission) error {
	query := `
		UPDATE permissions
		SET description = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, permission.Description, permission.ID)
	return err
}

func (r *permissionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM permissions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *permissionRepo) List(ctx context.Context, limit, offset int) ([]*models.Permission, error) {
	query := `
		SELECT id, name, description, created_at
		FROM permissions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
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
	return permissions, nil
}