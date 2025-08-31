package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductImageRepository interface {
	Create(ctx context.Context, image *models.ProductImage) error
	GetByProductID(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.ProductImage, error)
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.ProductImage, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteAllByProductID(ctx context.Context, tenantID, productID uuid.UUID) error
}

type productImageRepo struct {
	db *pgxpool.Pool
}

func NewProductImageRepo(db *pgxpool.Pool) ProductImageRepository {
	return &productImageRepo{db: db}
}

func (r *productImageRepo) Create(ctx context.Context, image *models.ProductImage) error {
	query := `
		INSERT INTO product_images (id, tenant_id, product_id, image_url, alt_text, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	image.ID = uuid.New()
	_, err := r.db.Exec(ctx, query, image.ID, image.TenantID, image.ProductID, image.ImageURL, image.AltText)
	return err
}

func (r *productImageRepo) GetByProductID(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.ProductImage, error) {
	query := `
		SELECT id, tenant_id, product_id, image_url, alt_text, created_at
		FROM product_images
		WHERE tenant_id = $1 AND product_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*models.ProductImage
	for rows.Next() {
		image := &models.ProductImage{}
		if err := rows.Scan(&image.ID, &image.TenantID, &image.ProductID, &image.ImageURL, &image.AltText, &image.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, nil
}

func (r *productImageRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.ProductImage, error) {
	query := `
		SELECT id, tenant_id, product_id, image_url, alt_text, created_at
		FROM product_images
		WHERE tenant_id = $1 AND id = $2
	`
	image := &models.ProductImage{}
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&image.ID, &image.TenantID, &image.ProductID, &image.ImageURL, &image.AltText, &image.CreatedAt)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (r *productImageRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM product_images WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *productImageRepo) DeleteAllByProductID(ctx context.Context, tenantID, productID uuid.UUID) error {
	query := `DELETE FROM product_images WHERE tenant_id = $1 AND product_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, productID)
	return err
}