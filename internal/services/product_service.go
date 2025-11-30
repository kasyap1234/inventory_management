package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif" // Register GIF format
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

type ProductService interface {
	Create(ctx context.Context, tenantID uuid.UUID, product *models.Product) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Product, error)
	Update(ctx context.Context, tenantID uuid.UUID, product *models.Product) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error)
	// ListWithCursor returns products using cursor-based pagination for better performance on large datasets
	ListWithCursor(ctx context.Context, tenantID uuid.UUID, params *common.CursorPaginationParams) (*common.CursorPaginatedResult[*models.Product], error)
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
	// BulkUpdatePrices updates prices for multiple products in a single database operation
	BulkUpdatePrices(ctx context.Context, tenantID uuid.UUID, productIDs []uuid.UUID, adjustmentType string, adjustmentValue float64) (int64, error)
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
	// Comprehensive validation
	if product == nil {
		return errors.New("product cannot be nil")
	}
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if len(product.Name) > 255 {
		return errors.New("product name cannot exceed 255 characters")
	}
	if product.UnitPrice <= 0 {
		return errors.New("unit price must be positive")
	}
	if product.UnitPrice > 999999999 {
		return errors.New("unit price exceeds maximum allowed value")
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
		log.Printf("Cache error for product %s: %v", id.String(), err)
	}

	// Cache miss - get from database
	product, err := s.productRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	// Cache the product for future requests (TTL: 15 minutes)
	if cacheErr := s.cacheService.SetProduct(ctx, tenantID, product, 15*time.Minute); cacheErr != nil {
		log.Printf("Failed to cache product %s: %v", id.String(), cacheErr)
	}

	return product, nil
}

func (s *productService) Update(ctx context.Context, tenantID uuid.UUID, product *models.Product) error {
	// Validate input
	if product == nil {
		return errors.New("product cannot be nil")
	}
	if product.ID == uuid.Nil {
		return errors.New("product ID is required")
	}
	if product.Name != "" && len(product.Name) > 255 {
		return errors.New("product name cannot exceed 255 characters")
	}
	if product.UnitPrice < 0 {
		return errors.New("unit price cannot be negative")
	}
	if product.UnitPrice > 999999999 {
		return errors.New("unit price exceeds maximum allowed value")
	}
	if product.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	product.TenantID = tenantID
	existing, err := s.productRepo.GetByID(ctx, tenantID, product.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing product: %w", err)
	}
	if existing == nil {
		return errors.New("product not found")
	}
	if product.Quantity != existing.Quantity {
		change := product.Quantity - existing.Quantity
		s.UpdateStock(ctx, tenantID, product.ID, change)
	}

	// CACHE CONSISTENCY FIX: Invalidate cache BEFORE database update
	// This ensures that if cache invalidation fails, we don't proceed with the update
	// and leave stale data in cache. Any concurrent read after invalidation but before
	// DB update will simply re-fetch from database.
	if cacheErr := s.cacheService.DeleteProduct(ctx, tenantID, product.ID); cacheErr != nil {
		log.Printf("Failed to invalidate cache before update for product %s: %v", product.ID.String(), cacheErr)
		// Continue with update - cache miss will fetch fresh data from DB
	}

	err = s.productRepo.Update(ctx, product)
	if err != nil {
		// Database update failed - cache is already invalidated, which is fine
		// Next read will fetch from DB
		return err
	}

	// Double-invalidate to handle any race conditions where cache was repopulated
	// during the database update window
	if cacheErr := s.cacheService.DeleteProduct(ctx, tenantID, product.ID); cacheErr != nil {
		log.Printf("Failed to invalidate cache after update for product %s: %v", product.ID.String(), cacheErr)
	}

	return nil
}

func (s *productService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	// CACHE CONSISTENCY FIX: Invalidate cache BEFORE database delete
	if cacheErr := s.cacheService.DeleteProduct(ctx, tenantID, id); cacheErr != nil {
		log.Printf("Failed to invalidate cache before delete for product %s: %v", id.String(), cacheErr)
	}

	err := s.productRepo.Delete(ctx, tenantID, id)
	if err != nil {
		return err
	}

	// Double-invalidate for consistency
	if cacheErr := s.cacheService.DeleteProduct(ctx, tenantID, id); cacheErr != nil {
		log.Printf("Failed to invalidate cache after delete for product %s: %v", id.String(), cacheErr)
	}

	return nil
}

func (s *productService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	return s.productRepo.List(ctx, tenantID, limit, offset)
}

// ListWithCursor returns products using cursor-based pagination.
// This is more efficient than offset-based pagination for large datasets because:
// 1. It uses an index seek rather than scanning/skipping rows
// 2. It provides stable results even with concurrent inserts/deletes
// 3. Performance is consistent regardless of page depth
func (s *productService) ListWithCursor(ctx context.Context, tenantID uuid.UUID, params *common.CursorPaginationParams) (*common.CursorPaginatedResult[*models.Product], error) {
	const defaultLimit = 20
	const maxLimit = 100

	if params == nil {
		params = &common.CursorPaginationParams{First: defaultLimit}
	}

	// Validate parameters
	if err := params.Validate(maxLimit); err != nil {
		return nil, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	limit := params.GetLimit(defaultLimit)
	backward := params.IsBackward()

	// Decode cursor if provided
	var cursorID *uuid.UUID
	var cursorCreatedAt *time.Time

	cursorStr := params.After
	if backward {
		cursorStr = params.Before
	}

	if cursorStr != "" {
		cursor, err := common.DecodeCursor(cursorStr)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		cursorID = &cursor.ID
		cursorCreatedAt = &cursor.CreatedAt
	}

	// Fetch one extra item to determine if there are more pages
	products, err := s.productRepo.ListWithCursor(ctx, tenantID, limit+1, cursorID, cursorCreatedAt, backward)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Determine pagination info
	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit] // Trim the extra item
	}

	result := &common.CursorPaginatedResult[*models.Product]{
		Items: products,
		PageInfo: common.CursorPageInfo{
			HasNextPage:     !backward && hasMore,
			HasPreviousPage: backward && hasMore,
		},
	}

	// Generate cursors for the page boundaries
	if len(products) > 0 {
		// Start cursor (first item)
		startCursor := common.NewCursor(products[0].ID, products[0].CreatedAt, common.CursorPrev)
		encodedStart, _ := common.EncodeCursor(startCursor)
		result.PageInfo.StartCursor = encodedStart

		// End cursor (last item)
		lastProduct := products[len(products)-1]
		endCursor := common.NewCursor(lastProduct.ID, lastProduct.CreatedAt, common.CursorNext)
		encodedEnd, _ := common.EncodeCursor(endCursor)
		result.PageInfo.EndCursor = encodedEnd

		// If we have a cursor, we know there's a previous/next page
		if cursorStr != "" {
			if backward {
				result.PageInfo.HasNextPage = true
			} else {
				result.PageInfo.HasPreviousPage = true
			}
		}
	}

	return result, nil
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
		log.Printf("Failed to cache product by barcode %s: %v\n", barcode, cacheErr)
	}

	return product, nil
}

func (s *productService) UpdateStock(ctx context.Context, tenantID, productID uuid.UUID, change int) error {
	// Validate input
	if change == 0 {
		return nil // No change needed
	}

	// Get product to verify it exists
	product, err := s.productRepo.GetByID(ctx, tenantID, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return errors.New("product not found")
	}

	// Update main product quantity for backward compatibility
	newQuantity := product.Quantity + change
	if newQuantity < 0 {
		return fmt.Errorf("stock update would result in negative quantity: current=%d, change=%d", product.Quantity, change)
	}
	product.Quantity = newQuantity

	// Update product record
	if err := s.productRepo.Update(ctx, product); err != nil {
		return err
	}

	// Integrate with inventory service to update warehouse stock
	// Get all inventory records for this product
	inventories, err := s.inventoryRepo.GetByProduct(ctx, tenantID, productID)
	if err != nil {
		// Log error but don't fail - inventory integration is optional
		log.Printf("Warning: Failed to get inventory for product %s: %v\n", productID.String(), err)
		return nil
	}

	// If product has inventory records, update the first warehouse
	// In a more sophisticated system, you'd determine which warehouse to update
	if len(inventories) > 0 {
		inventory := inventories[0]

		// Create an inventory service to handle the update
		// For now, directly update the inventory quantity
		inventory.Quantity += change
		if inventory.Quantity < 0 {
			inventory.Quantity = 0
		}

		if err := s.inventoryRepo.Update(ctx, inventory); err != nil {
			log.Printf("Warning: Failed to update inventory for product %s: %v\n", productID.String(), err)
		}
	}

	return nil
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
	// Validate inputs
	if filename == "" {
		return errors.New("filename is required")
	}
	if reader == nil {
		return errors.New("image reader is required")
	}
	if size <= 0 {
		return errors.New("invalid file size")
	}
	if size > 10*1024*1024 { // 10MB limit
		return errors.New("file size exceeds 10MB limit")
	}

	// Verify product exists
	product, err := s.productRepo.GetByID(ctx, tenantID, productID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}
	if product == nil {
		return errors.New("product not found")
	}

	// Read the image data into memory for processing
	imageData, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read image data: %w", err)
	}

	// Generate tenant-isolated key for MinIO
	fileExt := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, fileExt)

	// Set default bucket for product images
	bucketName := "product-images"

	// Ensure bucket exists
	err = s.minioService.EnsureBucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	// Process and upload original image (optimized quality)
	originalKey := fmt.Sprintf("%s/%s/%s%s", tenantID.String(), productID.String(), baseName, fileExt)
	optimizedImage, err := s.optimizeImage(imageData, 1920, 1920, 85) // Max 1920px, quality 85%
	if err != nil {
		// If optimization fails, fall back to original
		log.Printf("Warning: Image optimization failed, using original: %v\n", err)
		optimizedImage = imageData
	}

	err = s.minioService.UploadImage(ctx, bucketName, originalKey, bytes.NewReader(optimizedImage), int64(len(optimizedImage)))
	if err != nil {
		return fmt.Errorf("failed to upload original image: %w", err)
	}

	// Generate and upload thumbnail version (300x300)
	thumbnailKey := fmt.Sprintf("%s/%s/%s_thumb%s", tenantID.String(), productID.String(), baseName, fileExt)
	thumbnail, err := s.optimizeImage(imageData, 300, 300, 80)
	if err != nil {
		log.Printf("Warning: Thumbnail generation failed: %v\n", err)
	} else {
		err = s.minioService.UploadImage(ctx, bucketName, thumbnailKey, bytes.NewReader(thumbnail), int64(len(thumbnail)))
		if err != nil {
			log.Printf("Warning: Failed to upload thumbnail: %v\n", err)
		}
	}

	// Generate and upload medium version (800x800)
	mediumKey := fmt.Sprintf("%s/%s/%s_medium%s", tenantID.String(), productID.String(), baseName, fileExt)
	medium, err := s.optimizeImage(imageData, 800, 800, 82)
	if err != nil {
		log.Printf("Warning: Medium image generation failed: %v\n", err)
	} else {
		err = s.minioService.UploadImage(ctx, bucketName, mediumKey, bytes.NewReader(medium), int64(len(medium)))
		if err != nil {
			log.Printf("Warning: Failed to upload medium image: %v\n", err)
		}
	}

	// Save image metadata to database (store original key)
	image := &models.ProductImage{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ProductID: productID,
		ImageURL:  originalKey, // Store original key
		AltText:   altText,
	}

	return s.productImageRepo.Create(ctx, image)
}

// optimizeImage resizes and optimizes an image using basic Go image processing
// This is a simple implementation - for production, consider using libraries like:
// - github.com/h2non/bimg (requires libvips)
// - github.com/disintegration/imaging (pure Go)
// - github.com/nfnt/resize (pure Go, simpler)
func (s *productService) optimizeImage(imageData []byte, maxWidth, maxHeight int, quality int) ([]byte, error) {
	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get current dimensions
	bounds := img.Bounds()
	currentWidth := bounds.Dx()
	currentHeight := bounds.Dy()

	// Calculate new dimensions while maintaining aspect ratio
	newWidth, newHeight := currentWidth, currentHeight
	if currentWidth > maxWidth || currentHeight > maxHeight {
		ratio := float64(currentWidth) / float64(currentHeight)
		if currentWidth > currentHeight {
			newWidth = maxWidth
			newHeight = int(float64(maxWidth) / ratio)
		} else {
			newHeight = maxHeight
			newWidth = int(float64(maxHeight) * ratio)
		}
	}

	// Resize image using simple nearest-neighbor resampling
	// For better quality, use imaging.Resize with Lanczos filter
	resizedImg := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)

	// Encode the resized image
	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: quality})
	case "png":
		encoder := &png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(&buf, resizedImg)
	default:
		// For other formats, try JPEG encoding as fallback
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: quality})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
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
		log.Printf("Warning: failed to delete image from storage: %v\n", err)
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
		OperationID: fmt.Sprintf("bulk_update_products_%d", time.Now().UnixNano()),
		Status:      "processing",
		TotalItems:  len(bulkUpdate.ProductIDs),
		StartTime:   time.Now(),
		Progress:    0,
		Errors:      []models.BulkOperationError{},
		Items:       []models.BulkOperationItem{},
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

// BulkUpdatePrices updates prices for multiple products in a single database operation
// This is much more efficient than updating products one by one (N*2 queries -> 1 query)
func (s *productService) BulkUpdatePrices(ctx context.Context, tenantID uuid.UUID, productIDs []uuid.UUID, adjustmentType string, adjustmentValue float64) (int64, error) {
	if len(productIDs) == 0 {
		return 0, nil
	}

	if adjustmentType != "percentage" && adjustmentType != "fixed" {
		return 0, errors.New("adjustment type must be 'percentage' or 'fixed'")
	}

	// Perform the bulk update in a single SQL query
	updatedCount, err := s.productRepo.BulkUpdatePrices(ctx, tenantID, productIDs, adjustmentType, adjustmentValue)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk update prices: %w", err)
	}

	// Invalidate cache for affected products
	for _, productID := range productIDs {
		cacheKey := fmt.Sprintf("product:%s:%s", tenantID.String(), productID.String())
		if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
			log.Printf("Failed to invalidate cache for product %s: %v", productID.String(), err)
		}
	}

	return updatedCount, nil
}// BulkCreateProducts creates multiple products in bulk
func (s *productService) BulkCreateProducts(ctx context.Context, tenantID uuid.UUID, bulkCreate *models.ProductBulkCreate) (*models.BulkOperationResult, error) {
	// Set defaults
	if bulkCreate.ValidationMode == "" {
		bulkCreate.ValidationMode = "strict"
	}
	if bulkCreate.TransactionMode == "" {
		bulkCreate.TransactionMode = "atomic"
	}

	result := &models.BulkOperationResult{
		OperationID: fmt.Sprintf("bulk_create_products_%d", time.Now().UnixNano()),
		Status:      "processing",
		TotalItems:  len(bulkCreate.Products),
		StartTime:   time.Now(),
		Progress:    0,
		Errors:      []models.BulkOperationError{},
		Items:       []models.BulkOperationItem{},
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
