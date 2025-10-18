package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ProductHandlers handles HTTP requests for products
type ProductHandlers struct {
	productService services.ProductService
	rbacMiddleware *middleware.RBACMiddleware
}

// NewProductHandlers creates a new product handlers instance
func NewProductHandlers(productService services.ProductService, rbacMiddleware *middleware.RBACMiddleware) *ProductHandlers {
	return &ProductHandlers{
		productService: productService,
		rbacMiddleware: rbacMiddleware,
	}
}

// validateProduct validates product data
func (h *ProductHandlers) validateProduct(req *struct {
	Name          string  `json:"name"`
	CategoryID    *string `json:"category_id"`
	BatchNumber   *string `json:"batch_number"`
	ExpiryDate    *string `json:"expiry_date"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	Barcode       *string `json:"barcode"`
	UnitOfMeasure *string `json:"unit_of_measure"`
	Description   *string `json:"description"`
}) error {
	if strings.TrimSpace(req.Name) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Product name is required")
	}
	if req.UnitPrice <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Unit price must be positive")
	}
	if req.Quantity < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Quantity cannot be negative")
	}
	return nil
}

// validateUUID validates UUID string using the enhanced common validation
func (h *ProductHandlers) validateUUID(idStr string) (uuid.UUID, error) {
	id, err := common.ValidateUUID(idStr, "id")
	if err != nil {
		httpErr := echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid UUID format",
				"details": map[string]string{
					"id": err.Error(),
				},
			},
		})
		return uuid.Nil, httpErr
	}
	return id, nil
}

// CreateProduct handles POST /products
func (h *ProductHandlers) CreateProduct(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Name          string  `json:"name"`
		CategoryID    *string `json:"category_id"`
		BatchNumber   *string `json:"batch_number"`
		ExpiryDate    *string `json:"expiry_date"`
		Quantity      int     `json:"quantity"`
		UnitPrice     float64 `json:"unit_price"`
		Barcode       *string `json:"barcode"`
		UnitOfMeasure *string `json:"unit_of_measure"`
		Description   *string `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := h.validateProduct(&req); err != nil {
		return err
	}

	product := &models.Product{
		Name:          req.Name,
		BatchNumber:   req.BatchNumber,
		Quantity:      req.Quantity,
		UnitPrice:     req.UnitPrice,
		Barcode:       req.Barcode,
		UnitOfMeasure: req.UnitOfMeasure,
		Description:   req.Description,
	}

	if req.CategoryID != nil && *req.CategoryID != "" {
		categoryID, err := h.validateUUID(*req.CategoryID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID")
		}
		product.CategoryID = &categoryID
	}

	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		expiryDate, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid expiry date format")
		}
		product.ExpiryDate = &expiryDate
	}

	if err := h.productService.Create(ctx, tenantID, product); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Product created successfully",
		"product": product,
	})
}

// ListProducts handles GET /products
func (h *ProductHandlers) ListProducts(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	limit := 10 // default
	offset := 0 // default
	maxLimit := 1000 // Maximum items per page

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			if l > maxLimit {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Limit cannot exceed %d", maxLimit))
			}
			limit = l
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	products, err := h.productService.List(ctx, tenantID, limit, offset)
	if err != nil {
		log.Printf("Failed to list products for tenant %s: %v", tenantID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve products")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products": products,
		"limit":    limit,
		"offset":   offset,
	})
}

// BulkPriceUpdate handles POST /products/bulk-price-update
func (h *ProductHandlers) BulkPriceUpdate(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		ProductIDs []string `json:"product_ids"`
		Adjustment struct {
			Type  string  `json:"type"`  // "percentage" or "fixed"
			Value float64 `json:"value"` // percentage or fixed amount
		} `json:"adjustment"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if len(req.ProductIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "At least one product ID is required")
	}

	if req.Adjustment.Type != "percentage" && req.Adjustment.Type != "fixed" {
		return echo.NewHTTPError(http.StatusBadRequest, "Adjustment type must be 'percentage' or 'fixed'")
	}

	// Validate and convert product IDs
	productUUIDs := make([]uuid.UUID, len(req.ProductIDs))
	for i, idStr := range req.ProductIDs {
		id, err := h.validateUUID(idStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID: "+idStr)
		}
		productUUIDs[i] = id
	}

	// Update prices for all products
	updatedCount := 0
	for _, productID := range productUUIDs {
		// Get existing product
		product, err := h.productService.GetByID(ctx, tenantID, productID)
		if err != nil {
			log.Printf("Failed to get product %s: %v", productID, err)
			continue
		}

		// Calculate new price
		newPrice := product.UnitPrice
		if req.Adjustment.Type == "percentage" {
			newPrice = product.UnitPrice * (1 + req.Adjustment.Value/100)
		} else {
			newPrice = product.UnitPrice + req.Adjustment.Value
		}

		// Ensure price doesn't go negative
		if newPrice < 0 {
			newPrice = 0
		}

		// Update product price
		product.UnitPrice = newPrice
		if err := h.productService.Update(ctx, tenantID, product); err != nil {
			log.Printf("Failed to update product %s: %v", productID, err)
			continue
		}
		updatedCount++
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Bulk price update completed",
		"updated_count": updatedCount,
		"total_count":   len(req.ProductIDs),
	})
}

// GetProductByID handles GET /products/:id
func (h *ProductHandlers) GetProductByID(c echo.Context) error {
	ctx := c.Request().Context()

	productID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	product, err := h.productService.GetByID(ctx, tenantID, productID)
	if err != nil {
		log.Printf("Failed to get product %s for tenant %s: %v", productID, tenantID, err)
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve product")
	}

	return c.JSON(http.StatusOK, product)

}

// GetProduct handles GET /products/:id (alias for GetProductByID)
func (h *ProductHandlers) GetProduct(c echo.Context) error {
	return h.GetProductByID(c)
}

// UpdateProduct handles PUT /products/:id
func (h *ProductHandlers) UpdateProduct(c echo.Context) error {
	ctx := c.Request().Context()

	productID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Name          string  `json:"name"`
		CategoryID    *string `json:"category_id"`
		BatchNumber   *string `json:"batch_number"`
		ExpiryDate    *string `json:"expiry_date"`
		Quantity      int     `json:"quantity"`
		UnitPrice     float64 `json:"unit_price"`
		Barcode       *string `json:"barcode"`
		UnitOfMeasure *string `json:"unit_of_measure"`
		Description   *string `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := h.validateProduct(&req); err != nil {
		return err
	}

	existing, err := h.productService.GetByID(ctx, tenantID, productID)
	if err != nil {
		log.Printf("Failed to get product %s for update: %v", productID, err)
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve product")
	}

	existing.Name = req.Name
	existing.BatchNumber = req.BatchNumber
	existing.Quantity = req.Quantity
	existing.UnitPrice = req.UnitPrice
	existing.Barcode = req.Barcode
	existing.UnitOfMeasure = req.UnitOfMeasure
	existing.Description = req.Description

	if req.CategoryID != nil && *req.CategoryID != "" {
		categoryID, err := h.validateUUID(*req.CategoryID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID")
		}
		existing.CategoryID = &categoryID
	}

	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		expiryDate, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid expiry date format")
		}
		existing.ExpiryDate = &expiryDate
	}

	if err := h.productService.Update(ctx, tenantID, existing); err != nil {
		log.Printf("Failed to update product %s: %v", productID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update product")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Product updated successfully",
		"product": existing,
	})
}

// DeleteProduct handles DELETE /products/:id
func (h *ProductHandlers) DeleteProduct(c echo.Context) error {
	ctx := c.Request().Context()

	productID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.productService.Delete(ctx, tenantID, productID); err != nil {
		log.Printf("Failed to delete product %s: %v", productID, err)
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		if strings.Contains(err.Error(), "constraint") || strings.Contains(err.Error(), "foreign key") {
			return echo.NewHTTPError(http.StatusConflict, "Cannot delete product as it is referenced by other records")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete product")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}

// SearchProducts handles GET /products/search
func (h *ProductHandlers) SearchProducts(c echo.Context) error {
	ctx := c.Request().Context()
	start := time.Now()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	query := c.QueryParam("q")
	categoryIDStr := c.QueryParam("category_id")

	var categoryID *uuid.UUID
	if categoryIDStr != "" {
		catID, err := h.validateUUID(categoryIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID")
		}
		categoryID = &catID
	}

	limit := 10
	offset := 0

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	products, err := h.productService.Search(ctx, tenantID, query, categoryID, limit, offset)
	if err != nil {
		log.Printf("Failed to search products for tenant %s: %v", tenantID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search products")
	}

	// Record search usage (non-blocking via event)
	duration := time.Since(start)
	filterCount := 0
	if categoryID != nil {
		filterCount++
	}
	common.PublishEvent(ctx, "search_performed", map[string]interface{}{
		"entity_type":      "products",
		"search_term":      query,
		"filter_count":     filterCount,
		"result_count":     len(products),
		"response_time_ms": duration.Milliseconds(),
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products": products,
		"limit":    limit,
		"offset":   offset,
		"query":    query,
	})
}

// GetProductAnalytics handles GET /products/analytics
func (h *ProductHandlers) GetProductAnalytics(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	analytics, err := h.productService.CategoryAnalytics(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"analytics":   analytics,
		"description": "Category distribution of products",
	})
}

// UploadProductImage handles POST /products/:id/images
func (h *ProductHandlers) UploadProductImage(c echo.Context) error {
	ctx := c.Request().Context()

	productID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get file from form
	file, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Image file is required")
	}

	// Validate file size (5MB limit)
	const maxFileSize = 5 * 1024 * 1024 // 5MB in bytes
	if file.Size > maxFileSize {
		return echo.NewHTTPError(http.StatusBadRequest, "File size exceeds maximum limit of 5MB")
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	// Get file content type by opening file
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open image file")
	}
	defer src.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read file content")
	}
	contentType := http.DetectContentType(buffer)

	if !allowedTypes[contentType] {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed")
	}

	// Reset file pointer to beginning for re-reading
	src.Seek(0, 0)

	altText := c.FormValue("alt_text")

	err = h.productService.UploadProductImage(ctx, tenantID, productID, file.Filename, src, file.Size, &altText)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Image uploaded successfully",
	})
}

// GetProductImages handles GET /products/:id/images
func (h *ProductHandlers) GetProductImages(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse and validate product ID with enhanced validation
	idParam := c.Param("id")
	productID, err := common.ValidateUUID(idParam, "product_id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid product ID",
				"details": map[string]string{
					"product_id": err.Error(),
				},
			},
		})
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	images, err := h.productService.GetProductImages(ctx, tenantID, productID)
	if err != nil {
		// Log the error with context
		if logger := common.GetGlobalLogger(); logger != nil {
			logger.ErrorWithContext(ctx, "Failed to get product images", err, map[string]interface{}{
				"product_id": productID,
				"tenant_id":  tenantID,
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve product images")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"images":     images,
		"count":      len(images),
		"product_id": productID,
	})
}

// GetProductImageURL handles GET /products/:id/images/:imageId/url
func (h *ProductHandlers) GetProductImageURL(c echo.Context) error {
	ctx := c.Request().Context()

	imageID, err := h.validateUUID(c.Param("imageId"))
	if err != nil {
		return err
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	expiry := time.Hour * 24 // 24 hours default
	expiryStr := c.QueryParam("expiry_minutes")
	if expiryStr != "" {
		if minutes, err := strconv.Atoi(expiryStr); err == nil && minutes > 0 {
			expiry = time.Minute * time.Duration(minutes)
		}
	}

	url, err := h.productService.GetProductImageURL(ctx, tenantID, imageID, expiry)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"url":        url,
		"expires_in": expiry.String(),
	})
}

// DeleteProductImage handles DELETE /products/:id/images/:imageId
func (h *ProductHandlers) DeleteProductImage(c echo.Context) error {
	ctx := c.Request().Context()

	imageID, err := h.validateUUID(c.Param("imageId"))
	if err != nil {
		return err
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	err = h.productService.DeleteProductImage(ctx, tenantID, imageID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Image deleted successfully",
	})
}

// BulkUpdateProducts handles POST /products/bulk/update
func (h *ProductHandlers) BulkUpdateProducts(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req models.ProductBulkUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := h.validateBulkUpdateRequest(&req); err != nil {
		return err
	}

	result, err := h.productService.BulkUpdateProducts(ctx, tenantID, &req)
	if err != nil {
		log.Printf("Failed to bulk update products: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to bulk update products")
	}

	statusCode := http.StatusOK
	if result.Status == "partial" {
		statusCode = http.StatusPartialContent
	}

	return c.JSON(statusCode, result)
}

// BulkCreateProducts handles POST /products/bulk/create
func (h *ProductHandlers) BulkCreateProducts(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req models.ProductBulkCreate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := h.validateBulkCreateRequest(&req); err != nil {
		return err
	}

	result, err := h.productService.BulkCreateProducts(ctx, tenantID, &req)
	if err != nil {
		log.Printf("Failed to bulk create products: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to bulk create products")
	}

	statusCode := http.StatusCreated
	if result.Status == "partial" {
		statusCode = http.StatusPartialContent
	}

	return c.JSON(statusCode, result)
}

// validateBulkUpdateRequest validates bulk update request
func (h *ProductHandlers) validateBulkUpdateRequest(req *models.ProductBulkUpdate) error {
	if req == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Request body is required")
	}

	if len(req.ProductIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Product IDs are required")
	}

	if len(req.ProductIDs) > 1000 {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot update more than 1000 products at once")
	}

	// ProductIDs are already validated as uuid.UUID during binding
	// No additional validation needed here

	return nil
}

// validateBulkCreateRequest validates bulk create request
func (h *ProductHandlers) validateBulkCreateRequest(req *models.ProductBulkCreate) error {
	if req == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Request body is required")
	}

	if len(req.Products) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Products list is required")
	}

	if len(req.Products) > 500 {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot create more than 500 products at once")
	}

	return nil
}
