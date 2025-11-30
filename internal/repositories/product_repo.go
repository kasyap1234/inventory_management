package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	// ListWithCursor returns products using cursor-based pagination for better performance on large datasets
	ListWithCursor(ctx context.Context, tenantID uuid.UUID, limit int, cursorID *uuid.UUID, cursorCreatedAt *time.Time, backward bool) ([]*models.Product, error)
	// CountProducts returns the total count of products for a tenant (use sparingly)
	CountProducts(ctx context.Context, tenantID uuid.UUID) (int, error)
	GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*models.Product, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error)
	ListWithCategory(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error)
	CategoryAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]int, error)
	AdvancedSearch(ctx context.Context, tenantID uuid.UUID, filter *models.ProductSearchFilter) ([]*models.Product, error)
	UpdateQuantity(ctx context.Context, tenantID, id uuid.UUID, quantity int) error
}

type productRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (
			id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
			unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query,
		product.ID, product.TenantID, product.CategoryID, product.Name, product.BatchNumber, product.ExpiryDate,
		product.Quantity, product.UnitPrice, product.Barcode, product.UnitOfMeasure, product.Description,
		product.IsHazardous, product.HazardClass, product.SDSUrl, product.ActiveIngredients,
	)
	return err
}

func (r *productRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error) {
	product := &models.Product{}
	query := `
		SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
		       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
		       created_at, updated_at
		FROM products
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate,
		&product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description,
		&product.IsHazardous, &product.HazardClass, &product.SDSUrl, &product.ActiveIngredients,
		&product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepo) GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*models.Product, error) {
	product := &models.Product{}
	query := `
		SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
		       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
		       created_at, updated_at
		FROM products
		WHERE tenant_id = $1 AND barcode = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, barcode).Scan(
		&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate,
		&product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description,
		&product.IsHazardous, &product.HazardClass, &product.SDSUrl, &product.ActiveIngredients,
		&product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepo) Update(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET category_id = $1, name = $2, batch_number = $3, expiry_date = $4, quantity = $5, unit_price = $6, 
		    barcode = $7, unit_of_measure = $8, description = $9, is_hazardous = $10, hazard_class = $11, 
		    sds_url = $12, active_ingredients = $13, updated_at = NOW()
		WHERE tenant_id = $14 AND id = $15
	`
	_, err := r.db.Exec(ctx, query,
		product.CategoryID, product.Name, product.BatchNumber, product.ExpiryDate, product.Quantity, product.UnitPrice,
		product.Barcode, product.UnitOfMeasure, product.Description, product.IsHazardous, product.HazardClass,
		product.SDSUrl, product.ActiveIngredients, product.TenantID, product.ID,
	)
	return err
}

func (r *productRepo) UpdateQuantity(ctx context.Context, tenantID, id uuid.UUID, quantity int) error {
	query := `UPDATE products SET quantity = $1, updated_at = NOW() WHERE tenant_id = $2 AND id = $3`
	_, err := r.db.Exec(ctx, query, quantity, tenantID, id)
	return err
}

func (r *productRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM products WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *productRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	query := `
		SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
		       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
		       created_at, updated_at
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
		if err := rows.Scan(
			&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate,
			&product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description,
			&product.IsHazardous, &product.HazardClass, &product.SDSUrl, &product.ActiveIngredients,
			&product.CreatedAt, &product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// ListWithCursor returns products using cursor-based (keyset) pagination.
// This is more performant than offset-based pagination for large datasets because
// it uses an index seek rather than scanning and skipping rows.
// The cursor is based on (created_at, id) for stable ordering even with concurrent inserts.
func (r *productRepo) ListWithCursor(ctx context.Context, tenantID uuid.UUID, limit int, cursorID *uuid.UUID, cursorCreatedAt *time.Time, backward bool) ([]*models.Product, error) {
	var query string
	var args []interface{}

	if cursorID == nil || cursorCreatedAt == nil {
		// No cursor - return first/last page
		if backward {
			query = `
				SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
				       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
				       created_at, updated_at
				FROM products
				WHERE tenant_id = $1
				ORDER BY created_at ASC, id ASC
				LIMIT $2
			`
		} else {
			query = `
				SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
				       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
				       created_at, updated_at
				FROM products
				WHERE tenant_id = $1
				ORDER BY created_at DESC, id DESC
				LIMIT $2
			`
		}
		args = []interface{}{tenantID, limit}
	} else {
		// Cursor provided - use keyset pagination
		// The WHERE clause uses (created_at, id) tuple comparison for stable ordering
		if backward {
			// Backward pagination: get items BEFORE the cursor
			query = `
				SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
				       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
				       created_at, updated_at
				FROM products
				WHERE tenant_id = $1 
				  AND (created_at, id) > ($2, $3)
				ORDER BY created_at ASC, id ASC
				LIMIT $4
			`
		} else {
			// Forward pagination: get items AFTER the cursor
			query = `
				SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
				       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
				       created_at, updated_at
				FROM products
				WHERE tenant_id = $1 
				  AND (created_at, id) < ($2, $3)
				ORDER BY created_at DESC, id DESC
				LIMIT $4
			`
		}
		args = []interface{}{tenantID, *cursorCreatedAt, *cursorID, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products with cursor: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		if err := rows.Scan(
			&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate,
			&product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description,
			&product.IsHazardous, &product.HazardClass, &product.SDSUrl, &product.ActiveIngredients,
			&product.CreatedAt, &product.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products with cursor: %w", err)
	}

	// For backward pagination, reverse the results to maintain consistent ordering
	if backward && len(products) > 0 {
		for i, j := 0, len(products)-1; i < j; i, j = i+1, j-1 {
			products[i], products[j] = products[j], products[i]
		}
	}

	return products, nil
}

// CountProducts returns the total number of products for a tenant.
// Note: This can be expensive for large datasets. Use sparingly, typically only on first page load.
func (r *productRepo) CountProducts(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM products WHERE tenant_id = $1`
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}
	return count, nil
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
		SELECT p.id, p.tenant_id, p.category_id, p.name, p.batch_number, p.expiry_date, p.quantity, p.unit_price, 
		       p.barcode, p.unit_of_measure, p.description, p.is_hazardous, p.hazard_class, p.sds_url, 
		       p.active_ingredients, p.created_at, p.updated_at
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating advanced search results: %w", err)
	}

	return products, nil
}

func (r *productRepo) ListWithCategory(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error) {
	var query string
	var args []interface{}

	if categoryID != nil {
		query = `
			SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
			       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
			       created_at, updated_at
			FROM products
			WHERE tenant_id = $1 AND category_id = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{tenantID, *categoryID, limit, offset}
	} else {
		query = `
			SELECT p.id, p.tenant_id, p.category_id, p.name, p.batch_number, p.expiry_date, p.quantity, p.unit_price, 
			       p.barcode, p.unit_of_measure, p.description, p.is_hazardous, p.hazard_class, p.sds_url, 
			       p.active_ingredients, p.created_at, p.updated_at
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
		if err := rows.Scan(
			&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate,
			&product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description,
			&product.IsHazardous, &product.HazardClass, &product.SDSUrl, &product.ActiveIngredients,
			&product.CreatedAt, &product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products with category: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category analytics: %w", err)
	}

	return analytics, nil
}

func (r *productRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error) {
	var querySQL string
	var args []interface{}

	if categoryID != nil {
		querySQL = `
			SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
			       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
			       created_at, updated_at
			FROM products
			WHERE tenant_id = $1 AND category_id = $2 AND (name ILIKE $3 OR barcode ILIKE $3)
			ORDER BY created_at DESC
			LIMIT $4 OFFSET $5
		`
		args = []interface{}{tenantID, *categoryID, "%" + query + "%", limit, offset}
	} else {
		querySQL = `
			SELECT id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, 
			       unit_of_measure, description, is_hazardous, hazard_class, sds_url, active_ingredients, 
			       created_at, updated_at
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
		if err := rows.Scan(
			&product.ID, &product.TenantID, &product.CategoryID, &product.Name, &product.BatchNumber, &product.ExpiryDate,
			&product.Quantity, &product.UnitPrice, &product.Barcode, &product.UnitOfMeasure, &product.Description,
			&product.IsHazardous, &product.HazardClass, &product.SDSUrl, &product.ActiveIngredients,
			&product.CreatedAt, &product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}
