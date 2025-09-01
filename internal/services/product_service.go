package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type ProductService interface {
	Create(ctx context.Context, tenantID uuid.UUID, product *models.Product) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error)
	Update(ctx context.Context, tenantID uuid.UUID, product *models.Product) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error)
	GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*models.Product, error)
	UpdateStock(ctx context.Context, tenantID, productID uuid.UUID, change int) error
	Search(ctx context.Context, tenantID uuid.UUID, query string, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error)
	CategoryAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]int, error)
	UploadProductImage(ctx context.Context, tenantID, productID uuid.UUID, filename string, reader io.Reader, size int64, altText *string) error
	GetProductImages(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.ProductImage, error)
	GetProductImageURL(ctx context.Context, tenantID, imageID uuid.UUID, expiry time.Duration) (string, error)
	DeleteProductImage(ctx context.Context, tenantID, imageID uuid.UUID) error

	// Bulk operations
	BulkUpdateProducts(ctx context.Context, tenantID uuid.UUID, bulkUpdate *models.ProductBulkUpdate) (*models.BulkOperationResult, error)
	BulkCreateProducts(ctx context.Context, tenantID uuid.UUID, bulkCreate *models.ProductBulkCreate) (*models.BulkOperationResult, error)
}

type productService struct {
	productRepo      repositories.ProductRepository
	inventoryRepo    repositories.InventoryRepository
	categoryRepo     repositories.CategoryRepository
	productImageRepo repositories.ProductImageRepository
	minioService     MinioService
	cacheService     caching.CacheService
}

func NewProductService(productRepo repositories.ProductRepository, inventoryRepo repositories.InventoryRepository, categoryRepo repositories.CategoryRepository, productImageRepo repositories.ProductImageRepository, minioService MinioService, cacheService caching.CacheService) ProductService {
	return &productService{
		productRepo:      productRepo,
		inventoryRepo:    inventoryRepo,
		categoryRepo:     categoryRepo,
		productImageRepo: productImageRepo,
		minioService:     minioService,
		cacheService:     cacheService,
	}
}

func (s *productService) Create(ctx context.Context, tenantID uuid.UUID, product *models.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.UnitPrice <= 0 {
		return errors.New("unit price must be positive")
	}
	if product.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	// Check for barcode duplicates if barcode is provided
	if product.Barcode != nil && strings.TrimSpace(*product.Barcode) != "" {
		_, err := s.productRepo.GetByBarcode(ctx, tenantID, *product.Barcode)
		if err == nil {
			return fmt.Errorf("barcode %s already exists for another product", *product.Barcode)
		}
	}

	product.TenantID = tenantID
	if product.CategoryID != nil {
		_, err := s.categoryRepo.GetByID(ctx, tenantID, *product.CategoryID)
		if err != nil {
			return fmt.Errorf("category not found: %w", err)
		}
	}
	product.ID = uuid.New()
	return s.productRepo.Create(ctx, product)
}

func (s *productService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error) {
	// Try to get from cache first
	if cachedProduct, err := s.cacheService.GetProduct(ctx, tenantID, id); cachedProduct != nil {
		return cachedProduct, nil
	} else if err != nil {
		// Log error but continue to database - cache errors shouldn't fail the operation
		fmt.Printf("Cache error for product %s: %v\n", id.String(), err)
	}

	// Cache miss - get from database
	product, err := s.productRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	// Cache the product for future requests (TTL: 15 minutes)
	if cacheErr := s.cacheService.SetProduct(ctx, tenantID, product, 15*time.Minute); cacheErr != nil {
		fmt.Printf("Failed to cache product %s: %v\n", id.String(), cacheErr)
	}

	return product, nil
}

func (s *productService) Update(ctx context.Context, tenantID uuid.UUID, product *models.Product) error {
	product.TenantID = tenantID
	existing, err := s.productRepo.GetByID(ctx, tenantID, product.ID)
	if err != nil {
		return err
	}
	if product.Quantity != existing.Quantity {
		change := product.Quantity - existing.Quantity
		s.UpdateStock(ctx, tenantID, product.ID, change)
	}

	err = s.productRepo.Update(ctx, product)
	if err != nil {
		return err
	}

	// Invalidate cache for this product
	if cacheErr := s.cacheService.DeleteProduct(ctx, tenantID, product.ID); cacheErr != nil {
		fmt.Printf("Failed to invalidate cache for product %s: %v\n", product.ID.String(), cacheErr)
	}

	return nil
}

func (s *productService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	err := s.productRepo.Delete(ctx, tenantID, id)
	if err != nil {
		return err
	}

	// Invalidate cache for this product
	if cacheErr := s.cacheService.DeleteProduct(ctx, tenantID, id); cacheErr != nil {
		fmt.Printf("Failed to invalidate cache for product %s: %v\n", id.String(), cacheErr)
	}

	return nil
}

func (s *productService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	return s.productRepo.List(ctx, tenantID, limit, offset)
}

func (s *productService) GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*models.Product, error) {
	// For barcode lookups, we don't cache them directly, but we cache products by ID
	// So we need to get from DB first, then cache the product
	product, err := s.productRepo.GetByBarcode(ctx, tenantID, barcode)
	if err != nil {
		return nil, err
	}

	// Cache the product for future requests (TTL: 15 minutes)
	if cacheErr := s.cacheService.SetProduct(ctx, tenantID, product, 15*time.Minute); cacheErr != nil {
		fmt.Printf("Failed to cache product by barcode %s: %v\n", barcode, cacheErr)
	}

	return product, nil
}

func (s *productService) UpdateStock(ctx context.Context, tenantID, productID uuid.UUID, change int) error {
	// TODO: Integrate with warehouse management to get appropriate warehouse
	// For now, use a temporary default warehouse approach
	product, err := s.productRepo.GetByID(ctx, tenantID, productID)
	if err != nil {
		return err
	}

	// Update main product quantity for backward compatibility
	product.Quantity += change
	if product.Quantity < 0 {
		product.Quantity = 0
	}

	// TODO: Get actual default warehouse for tenant
	// defaultWarehouseID := s.getDefaultWarehouseID(ctx, tenantID)
	// Also move to use inventory service's AdjustStock method

	return s.productRepo.Update(ctx, product)
}

// Search products by query string with optional category filter
func (s *productService) Search(ctx context.Context, tenantID uuid.UUID, query string, categoryID *uuid.UUID, limit, offset int) ([]*models.Product, error) {

	if query == "" {
		products, err := s.List(ctx, tenantID, limit, offset)
		return products, err
	}

	products, err := s.productRepo.Search(ctx, tenantID, query, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}

	return products, nil
}

// CategoryAnalytics returns analytics about product distribution by category
func (s *productService) CategoryAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]int, error) {
	return s.productRepo.CategoryAnalytics(ctx, tenantID)
}

// UploadProductImage uploads and processes a product image with optimization
func (s *productService) UploadProductImage(ctx context.Context, tenantID, productID uuid.UUID, filename string, reader io.Reader, size int64, altText *string) error {
	// Verify product exists
	_, err := s.productRepo.GetByID(ctx, tenantID, productID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// TODO: Add image processing for resizing and optimization
	// For example using github.com/nfnt/resize library:
	// - Resize to multiple sizes (thumbnail, medium, original)
	// - Optimize quality to reduce file size
	// - Convert to appropriate formats

	// Generate tenant-isolated key for MinIO
	fileExt := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, fileExt)
	objectKey := fmt.Sprintf("%s/%s/%s%s", tenantID.String(), productID.String(), baseName, fileExt)

	// Set default bucket for product images
	bucketName := "product-images"

	// Ensure bucket exists
	err = s.minioService.EnsureBucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	// Upload original image to MinIO
	err = s.minioService.UploadImage(ctx, bucketName, objectKey, reader, size)
	if err != nil {
		return fmt.Errorf("failed to upload image to storage: %w", err)
	}

	// TODO: Generate and upload resized versions
	// e.g., thumbnail: small resolution, medium: reasonable resolution

	// Save image metadata to database
	image := &models.ProductImage{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ProductID: productID,
		ImageURL:  objectKey, // Store key instead of full URL for tenant isolation
		AltText:   altText,
	}

	return s.productImageRepo.Create(ctx, image)
}

// GetProductImages retrieves all images for a product
func (s *productService) GetProductImages(ctx context.Context, tenantID, productID uuid.UUID) ([]*models.ProductImage, error) {
	return s.productImageRepo.GetByProductID(ctx, tenantID, productID)
}

// GetProductImageURL generates a pre-signed URL for accessing the image
func (s *productService) GetProductImageURL(ctx context.Context, tenantID, imageID uuid.UUID, expiry time.Duration) (string, error) {
	// Get image metadata
	image, err := s.productImageRepo.GetByID(ctx, tenantID, imageID)
	if err != nil {
		return "", fmt.Errorf("image not found: %w", err)
	}

	// Generate pre-signed URL
	bucketName := "product-images"
	url, err := s.minioService.GetPresignedURL(bucketName, image.ImageURL, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate image URL: %w", err)
	}

	return url, nil
}

// DeleteProductImage removes a product image from storage and database
func (s *productService) DeleteProductImage(ctx context.Context, tenantID, imageID uuid.UUID) error {
	// Get image metadata first
	image, err := s.productImageRepo.GetByID(ctx, tenantID, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	// Delete from storage
	bucketName := "product-images"
	err = s.minioService.DeleteImage(ctx, bucketName, image.ImageURL)
	if err != nil {
		// Log error but continue to delete from database
		fmt.Printf("Warning: failed to delete image from storage: %v\n", err)
	}

	// Delete from database
	return s.productImageRepo.Delete(ctx, tenantID, imageID)
}

// BulkUpdateProducts performs bulk updates on multiple products
func (s *productService) BulkUpdateProducts(ctx context.Context, tenantID uuid.UUID, bulkUpdate *models.ProductBulkUpdate) (*models.BulkOperationResult, error) {
	// Set defaults
	if bulkUpdate.ValidationMode == "" {
		bulkUpdate.ValidationMode = "strict"
	}
	if bulkUpdate.TransactionMode == "" {
		bulkUpdate.TransactionMode = "atomic"
	}

	result := &models.BulkOperationResult{
		OperationID:   fmt.Sprintf("bulk_update_products_%d", time.Now().UnixNano()),
		Status:       "processing",
		TotalItems:    len(bulkUpdate.ProductIDs),
		StartTime:     time.Now(),
		Progress:      0,
		Errors:        []models.BulkOperationError{},
		Items:         []models.BulkOperationItem{},
	}

	totalItems := len(bulkUpdate.ProductIDs)

	for i, productID := range bulkUpdate.ProductIDs {
		// Get existing product
		product, err := s.productRepo.GetByID(ctx, tenantID, productID)
		if err != nil {
			result.FailedItems++
			errorMsg := fmt.Sprintf("Failed to get product: %v", err)
			result.Errors = append(result.Errors, models.BulkOperationError{
				ItemIndex: i,
				ItemID:    productID.String(),
				Error:     errorMsg,
			})
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    productID.String(),
				Status:    "failed",
				Error:     &errorMsg,
			})
			if bulkUpdate.ValidationMode == "strict" {
				continue // Skip and continue for strict mode, but we'll update anyway
			}
		}

		// Apply updates
		updated := false
		if bulkUpdate.CategoryID != nil {
			product.CategoryID = bulkUpdate.CategoryID
			updated = true
		}

		if bulkUpdate.UnitPriceChange != nil {
			if bulkUpdate.UnitPriceMode == "percentage" {
				if *bulkUpdate.UnitPriceChange > -100 {
					newPrice := product.UnitPrice * (1 + *bulkUpdate.UnitPriceChange/100)
					product.UnitPrice = newPrice
					updated = true
				}
			} else {
				newPrice := product.UnitPrice + *bulkUpdate.UnitPriceChange
				if newPrice >= 0 {
					product.UnitPrice = newPrice
					updated = true
				}
			}
		}

		if bulkUpdate.BatchNumber != nil {
			product.BatchNumber = bulkUpdate.BatchNumber
			updated = true
		}

		if bulkUpdate.ExpiryDate != nil {
			product.ExpiryDate = bulkUpdate.ExpiryDate
			updated = true
		}

		if bulkUpdate.UnitOfMeasure != nil {
			product.UnitOfMeasure = bulkUpdate.UnitOfMeasure
			updated = true
		}

		if bulkUpdate.Description != nil {
			product.Description = bulkUpdate.Description
			updated = true
		}

		if updated {
			err = s.productRepo.Update(ctx, product)
			if err != nil {
				result.FailedItems++
				errorMsg := fmt.Sprintf("Failed to update product: %v", err)
				result.Errors = append(result.Errors, models.BulkOperationError{
					ItemIndex: i,
					ItemID:    productID.String(),
					Error:     errorMsg,
				})
				result.Items = append(result.Items, models.BulkOperationItem{
					ItemIndex: i,
					ItemID:    productID.String(),
					Status:    "failed",
					Error:     &errorMsg,
				})
			} else {
				result.ProcessedItems++
				result.Items = append(result.Items, models.BulkOperationItem{
					ItemIndex: i,
					ItemID:    productID.String(),
					Status:    "success",
				})
			}
		} else {
			result.ProcessedItems++
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    productID.String(),
				Status:    "success",
			})
		}

		// Update progress
		result.Progress = float64(i+1) / float64(totalItems) * 100
	}

	result.Status = "completed"
	if result.FailedItems > 0 && result.ProcessedItems > 0 {
		result.Status = "partial"
	}
	result.CompletionTime = &time.Time{}
	*result.CompletionTime = time.Now()

	return result, nil
}

// BulkCreateProducts creates multiple products in bulk
func (s *productService) BulkCreateProducts(ctx context.Context, tenantID uuid.UUID, bulkCreate *models.ProductBulkCreate) (*models.BulkOperationResult, error) {
	// Set defaults
	if bulkCreate.ValidationMode == "" {
		bulkCreate.ValidationMode = "strict"
	}
	if bulkCreate.TransactionMode == "" {
		bulkCreate.TransactionMode = "atomic"
	}

	result := &models.BulkOperationResult{
		OperationID:   fmt.Sprintf("bulk_create_products_%d", time.Now().UnixNano()),
		Status:       "processing",
		TotalItems:    len(bulkCreate.Products),
		StartTime:     time.Now(),
		Progress:      0,
		Errors:        []models.BulkOperationError{},
		Items:         []models.BulkOperationItem{},
	}

	totalItems := len(bulkCreate.Products)

	for i, product := range bulkCreate.Products {
		// Set tenant ID
		product.TenantID = tenantID
		product.ID = uuid.New()

		// Basic validation
		if product.Name == "" || product.UnitPrice <= 0 || product.Quantity < 0 {
			if bulkCreate.ValidationMode == "skip_invalid" {
				continue
			}
			result.FailedItems++
			errorMsg := "Invalid product data: name is required, unit_price must be positive, quantity cannot be negative"
			result.Errors = append(result.Errors, models.BulkOperationError{
				ItemIndex: i,
				ItemID:    product.ID.String(),
				Error:     errorMsg,
			})
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    product.ID.String(),
				Status:    "failed",
				Error:     &errorMsg,
			})
			continue
		}

		// Check for duplicate barcodes within the batch
		if product.Barcode != nil && strings.TrimSpace(*product.Barcode) != "" {
			for j := 0; j < i; j++ {
				if bulkCreate.Products[j].Barcode != nil &&
				   strings.TrimSpace(*bulkCreate.Products[j].Barcode) == strings.TrimSpace(*product.Barcode) {
					result.FailedItems++
					errorMsg := fmt.Sprintf("Duplicate barcode %s in batch", *product.Barcode)
					result.Errors = append(result.Errors, models.BulkOperationError{
						ItemIndex: i,
						ItemID:    product.ID.String(),
						Error:     errorMsg,
					})
					result.Items = append(result.Items, models.BulkOperationItem{
						ItemIndex: i,
						ItemID:    product.ID.String(),
						Status:    "failed",
						Error:     &errorMsg,
					})
					continue
				}
			}
			// Also check against existing products
			_, err := s.productRepo.GetByBarcode(ctx, tenantID, *product.Barcode)
			if err == nil {
				result.FailedItems++
				errorMsg := fmt.Sprintf("Barcode %s already exists", *product.Barcode)
				result.Errors = append(result.Errors, models.BulkOperationError{
					ItemIndex: i,
					ItemID:    product.ID.String(),
					Error:     errorMsg,
				})
				result.Items = append(result.Items, models.BulkOperationItem{
					ItemIndex: i,
					ItemID:    product.ID.String(),
					Status:    "failed",
					Error:     &errorMsg,
				})
				continue
			}
		}

		// Create product
		err := s.productRepo.Create(ctx, product)
		if err != nil {
			result.FailedItems++
			errorMsg := fmt.Sprintf("Failed to create product: %v", err)
			result.Errors = append(result.Errors, models.BulkOperationError{
				ItemIndex: i,
				ItemID:    product.ID.String(),
				Error:     errorMsg,
			})
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    product.ID.String(),
				Status:    "failed",
				Error:     &errorMsg,
			})
		} else {
			result.ProcessedItems++
			result.Items = append(result.Items, models.BulkOperationItem{
				ItemIndex: i,
				ItemID:    product.ID.String(),
				Status:    "success",
			})
		}

		// Update progress
		result.Progress = float64(i+1) / float64(totalItems) * 100
	}

	result.Status = "completed"
	if result.FailedItems > 0 && result.ProcessedItems > 0 {
		result.Status = "partial"
	}
	result.CompletionTime = &time.Time{}
	*result.CompletionTime = time.Now()

	return result, nil
}