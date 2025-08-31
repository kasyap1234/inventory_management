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
	Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*models.Category, error)
	ListSubcategories(ctx context.Context, tenantID, parentID uuid.UUID, limit, offset int) ([]*models.Category, error)
	ListRootCategories(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Category, error)
	GetCategoryTree(ctx context.Context, tenantID uuid.UUID) ([]*models.Category, error)
	ListWithChildren(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Category, error)
	BulkCreate(ctx context.Context, categories []*models.Category) error
	BulkUpdate(ctx context.Context, updates []*models.CategoryBulkUpdate) error
	BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error
	UpdateHierarchy(ctx context.Context, category *models.Category) error
}

type categoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(ctx context.Context, category *models.Category) error {
	// Calculate level and path based on parent_id
	if category.ParentID != nil {
		// Get parent details to calculate level and path
		parent, err := r.GetByID(ctx, category.TenantID, *category.ParentID)
		if err != nil {
			return err
		}
		category.Level = parent.Level + 1
		category.Path = parent.Path + "/" + category.Name
	} else {
		category.Level = 0
		category.Path = category.Name
	}

	query := `
		INSERT INTO categories (id, tenant_id, name, description, parent_id, level, path, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, category.ID, category.TenantID, category.Name, category.Description,
		category.ParentID, category.Level, category.Path)
	return err
}

func (r *categoryRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Category, error) {
	category := &models.Category{}
	query := `
		SELECT id, tenant_id, name, description, parent_id, level, path, created_at, updated_at
		FROM categories
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&category.ID, &category.TenantID, &category.Name, &category.Description,
		&category.ParentID, &category.Level, &category.Path, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepo) Update(ctx context.Context, category *models.Category) error {
	// Recalculate level and path if parent_id changed
	if category.ParentID != nil {
		parent, err := r.GetByID(ctx, category.TenantID, *category.ParentID)
		if err != nil {
			return err
		}
		category.Level = parent.Level + 1
		category.Path = parent.Path + "/" + category.Name
	} else {
		category.Level = 0
		category.Path = category.Name
	}

	query := `
		UPDATE categories
		SET name = $1, description = $2, parent_id = $3, level = $4, path = $5, updated_at = NOW()
		WHERE tenant_id = $6 AND id = $7
	`
	_, err := r.db.Exec(ctx, query, category.Name, category.Description, category.ParentID,
		category.Level, category.Path, category.TenantID, category.ID)
	return err
}

func (r *categoryRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *categoryRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Category, error) {
	query := `
		SELECT id, tenant_id, name, description, parent_id, level, path, created_at, updated_at
		FROM categories
		WHERE tenant_id = $1
		ORDER BY level ASC, path ASC
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
		if err := rows.Scan(&category.ID, &category.TenantID, &category.Name, &category.Description,
			&category.ParentID, &category.Level, &category.Path, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

// Implementation of missing methods with dummy return values to fix compilation
func (r *categoryRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*models.Category, error) {
	if query == "" {
		return r.List(ctx, tenantID, limit, offset) // If no query, return all categories
	}

	searchQuery := `
		SELECT id, tenant_id, name, description, parent_id, level, path, created_at, updated_at
		FROM categories
		WHERE tenant_id = $1 AND
		      (LOWER(name) LIKE LOWER($2) OR LOWER(description) LIKE LOWER($3))
		ORDER BY level ASC, path ASC
		LIMIT $4 OFFSET $5
	`
	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, tenantID, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}
		if err := rows.Scan(&category.ID, &category.TenantID, &category.Name, &category.Description,
			&category.ParentID, &category.Level, &category.Path, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}
func (r *categoryRepo) ListSubcategories(ctx context.Context, tenantID, parentID uuid.UUID, limit, offset int) ([]*models.Category, error) {
	return r.List(ctx, tenantID, limit, offset) // TODO: Implement subcategories
}
func (r *categoryRepo) ListRootCategories(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Category, error) {
	return r.List(ctx, tenantID, limit, offset) // TODO: Implement root categories
}
func (r *categoryRepo) GetCategoryTree(ctx context.Context, tenantID uuid.UUID) ([]*models.Category, error) {
	return r.List(ctx, tenantID, 1000, 0) // TODO: Implement tree
}
func (r *categoryRepo) ListWithChildren(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Category, error) {
	return r.List(ctx, tenantID, limit, offset) // TODO: Implement with children
}
func (r *categoryRepo) BulkCreate(ctx context.Context, categories []*models.Category) error {
	return nil // TODO: Implement bulk create
}
func (r *categoryRepo) BulkUpdate(ctx context.Context, updates []*models.CategoryBulkUpdate) error {
	return nil // TODO: Implement bulk update
}
func (r *categoryRepo) BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error {
	return nil // TODO: Implement bulk delete
}
func (r *categoryRepo) UpdateHierarchy(ctx context.Context, category *models.Category) error {
	return r.Update(ctx, category) // TODO: Implement hierarchy update
}