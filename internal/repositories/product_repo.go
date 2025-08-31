package repositories

import (
	"context"
	"fmt"
	"strings"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error)
	GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*models.Product, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error)
	ListWithCategory(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error)
	CategoryAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]int, error)
	AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.ProductSearchFilter) ([]*models.Product, error)
}

type productRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, unit_of_measure, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, product.ID, product.TenantID, product.CategoryID, product.Name, product.BatchNumber, product.ExpiryDate, product.Quantity, product.UnitPrice, product.Barcode, product.UnitOfMeasure, product.Description)
	return err
}

func (r *productRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error) {
	product := &models.Product{}
	query := `
		SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, unit_of_measure, description, created_at, updated_at
		FROM products
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate, &product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepo) GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*models.Product, error) {
	product := &models.Product{}
	query := `
		SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, unit_of_measure, description, created_at, updated_at
		FROM products
		WHERE tenant_id = $1 AND barcode = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, barcode).Scan(&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate, &product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepo) Update(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET category_id = $1, name = $2, batch_number = $3, expiry_date = $4, quantity = $5, unit_price = $6, barcode = $7, unit_of_measure = $8, description = $9, updated_at = NOW()
		WHERE tenant_id = $10 AND id = $11
	`
	_, err := r.db.Exec(ctx, query, product.CategoryID, product.Name, product.BatchNumber, product.ExpiryDate, product.Quantity, product.UnitPrice, product.Barcode, product.UnitOfMeasure, product.Description, product.TenantID, product.ID)
	return err
}

func (r *productRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM products WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *productRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	query := `
		SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, unit_of_measure, description, created_at, updated_at
		FROM products
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		if err := rows.Scan(&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate, &product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *productRepo) AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.ProductSearchFilter) ([]*models.Product, error) {
	// Set defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	// Build query dynamically
	queryBase := `
		SELECT p.id, p.tenant_id, p.category_id, p.name, p.batch_number, p.expiry_date, p.quantity, p.unit_price, p.barcode, p.unit_of_measure, p.description, p.created_at, p.updated_at
		FROM products p
		WHERE p.tenant_id = $1
	`
	args := []interface{}{tenantID}
	conditionCount := 1

	// Full-text search across multiple fields
	if filter.Query != "" {
		conditionCount++
		queryBase += fmt.Sprintf(` AND (
			p.name ILIKE $%d OR
			p.barcode ILIKE $%d OR
			COALESCE(p.description, '') ILIKE $%d OR
			EXISTS (
				SELECT 1 FROM categories c
				WHERE c.tenant_id = p.tenant_id AND c.id = p.category_id AND c.name ILIKE $%d
			)
		)`, conditionCount, conditionCount, conditionCount, conditionCount)
		args = append(args, "%"+filter.Query+"%")
	}

	// Category filter
	if filter.CategoryID != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND p.category_id = $%d`, conditionCount)
		args = append(args, *filter.CategoryID)
	}

	// Quantity range
	if filter.MinQuantity != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND p.quantity >= $%d`, conditionCount)
		args = append(args, *filter.MinQuantity)
	}
	if filter.MaxQuantity != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND p.quantity <= $%d`, conditionCount)
		args = append(args, *filter.MaxQuantity)
	}

	// Price range
	if filter.MinPrice != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND p.unit_price >= $%d`, conditionCount)
		args = append(args, *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND p.unit_price <= $%d`, conditionCount)
		args = append(args, *filter.MaxPrice)
	}

	// Expiry date range
	if filter.ExpiryBefore != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND p.expiry_date <= $%d`, conditionCount)
		args = append(args, *filter.ExpiryBefore)
	}
	if filter.ExpiryAfter != nil {
		conditionCount++
		queryBase += fmt.Sprintf(` AND p.expiry_date >= $%d`, conditionCount)
		args = append(args, *filter.ExpiryAfter)
	}

	// Barcode exact match
	if filter.Barcode != nil && *filter.Barcode != "" {
		conditionCount++
		queryBase += fmt.Sprintf(` AND p.barcode = $%d`, conditionCount)
		args = append(args, *filter.Barcode)
	}

	// Ordering
	validSortFields := map[string]bool{
		"name": true, "created_at": true, "quantity": true, "unit_price": true,
	}
	sortField := "p.created_at"
	if validSortFields[filter.SortBy] {
		sortField = "p." + filter.SortBy
		if filter.SortBy == "name" || filter.SortBy == "created_at" {
			sortField = "p." + filter.SortBy
		}
	}
	sortOrder := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		sortOrder = "ASC"
	}
	queryBase += fmt.Sprintf(` ORDER BY %s %s`, sortField, sortOrder)

	// Pagination
	conditionCount++
	queryBase += fmt.Sprintf(` LIMIT $%d`, conditionCount)
	args = append(args, filter.Limit)
	if filter.Offset > 0 {
		conditionCount++
		queryBase += fmt.Sprintf(` OFFSET $%d`, conditionCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, queryBase, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		if err := rows.Scan(&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate, &product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *productRepo) ListWithCategory(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error) {
	var query string
	var args []interface{}

	if categoryID != nil {
		query = `
			SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, unit_of_measure, description, created_at, updated_at
			FROM products
			WHERE tenant_id = $1 AND category_id = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{tenantID, *categoryID, limit, offset}
	} else {
		query = `
			SELECT p.id, p.tenant_id, p.category_id, p.name, p.batch_number, p.expiry_date, p.quantity, p.unit_price, p.barcode, p.unit_of_measure, p.description, p.created_at, p.updated_at
			FROM products p
			LEFT JOIN categories c ON p.category_id = c.id AND p.tenant_id = c.tenant_id
			WHERE p.tenant_id = $1
			ORDER BY c.name, p.name
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{tenantID, limit, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		if err := rows.Scan(&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate, &product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *productRepo) CategoryAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]int, error) {
	query := `
		SELECT COALESCE(c.name, 'Uncategorized'), COUNT(p.id)
		FROM categories c
		LEFT JOIN products p ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		WHERE c.tenant_id = $1
		GROUP BY c.id, c.name
		UNION ALL
		SELECT 'Uncategorized' as name, COUNT(*) as count
		FROM products
		WHERE tenant_id = $1 AND category_id IS NULL
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	analytics := make(map[string]int)
	for rows.Next() {
		var categoryName string
		var count int
		if err := rows.Scan(&categoryName, &count); err != nil {
			return nil, err
		}
		analytics[categoryName] = count
	}

	return analytics, nil
}

func (r *productRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error) {
	var querySQL string
	var args []interface{}

	if categoryID != nil {
		querySQL = `
			SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, unit_of_measure, description, created_at, updated_at
			FROM products
			WHERE tenant_id = $1 AND category_id = $2 AND (name ILIKE $3 OR barcode ILIKE $3)
			ORDER BY created_at DESC
			LIMIT $4 OFFSET $5
		`
		args = []interface{}{tenantID, *categoryID, "%" + query + "%", limit, offset}
	} else {
		querySQL = `
			SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, unit_of_measure, description, created_at, updated_at
			FROM products
			WHERE tenant_id = $1 AND (name ILIKE $2 OR barcode ILIKE $2)
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{tenantID, "%" + query + "%", limit, offset}
	}

	rows, err := r.db.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		if err := rows.Scan(&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate, &product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}