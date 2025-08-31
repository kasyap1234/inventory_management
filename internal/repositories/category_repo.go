package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Category, error)
}

type categoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (id, tenant_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, category.ID, category.TenantID, category.Name, category.Description)
	return err
}

func (r *categoryRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Category, error) {
	category := &models.Category{}
	query := `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM categories
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&category.ID, &category.TenantID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepo) Update(ctx context.Context, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, updated_at = NOW()
		WHERE tenant_id = $3 AND id = $4
	`
	_, err := r.db.Exec(ctx, query, category.Name, category.Description, category.TenantID, category.ID)
	return err
}

func (r *categoryRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *categoryRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Category, error) {
	query := `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM categories
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}
		if err := rows.Scan(&category.ID, &category.TenantID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}