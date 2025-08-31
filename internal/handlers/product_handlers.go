package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	Name           string   `json:"name"`
	CategoryID     *string  `json:"category_id"`
	BatchNumber    *string  `json:"batch_number"`
	ExpiryDate     *string  `json:"expiry_date"`
	Quantity       int      `json:"quantity"`
	UnitPrice      float64  `json:"unit_price"`
	Barcode        *string  `json:"barcode"`
	UnitOfMeasure  *string  `json:"unit_of_measure"`
	Description    *string  `json:"description"`
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

// validateUUID validates UUID string with enhanced checks
func (h *ProductHandlers) validateUUID(idStr string) (uuid.UUID, error) {
	// Enhanced logging for UUID format validation
	log.Printf("PRODUCT_HANDLER_UUID_VALIDATION: Received UUID string: '%s' (length: %d)", idStr, len(idStr))

	// Trim whitespace to handle edge case of leading/trailing spaces
	idStr = strings.TrimSpace(idStr)
	log.Printf("PRODUCT_HANDLER_UUID_VALIDATION: After trimming: '%s' (length: %d)", idStr, len(idStr))

	// Check for empty string after trimming
	if idStr == "" {
		log.Printf("PRODUCT_HANDLER_UUID_VALIDATION: Empty UUID string received after trimming")
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format: empty string")
	}

	// Validate exact length of 36 characters
	if len(idStr) != 36 {
		log.Printf("PRODUCT_HANDLER_UUID_VALIDATION: Incorrect length - expected 36, got %d", len(idStr))
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format: length must be 36 characters (including hyphens)")
	}

	// Check for proper format (8-4-4-4-12) - hyphens at positions 8, 13, 18, 23
	if idStr[8] != '-' || idStr[13] != '-' || idStr[18] != '-' || idStr[23] != '-' {
		log.Printf("PRODUCT_HANDLER_UUID_VALIDATION: Invalid hyphen placement: positions 8,13,18,23 should be hyphens")
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format: hyphens must be at positions 9, 14, 19, and 24")
	}

	// Fallback validation using standard UUID parser to catch any remaining format issues
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("PRODUCT_HANDLER_UUID_VALIDATION: UUID parsing failed - input: '%s', error: %v", idStr, err)
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format: contains invalid characters or format")
	}

	log.Printf("PRODUCT_HANDLER_UUID_VALIDATION: UUID validation successful: %s", id.String())
	return id, nil
}

// CreateProduct handles POST /products
func (h *ProductHandlers) CreateProduct(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract tenant ID from context
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Name           string   `json:"name"`
		CategoryID     *string  `json:"category_id"`
		BatchNumber    *string  `json:"batch_number"`
		ExpiryDate     *string  `json:"expiry_date"`
		Quantity       int      `json:"quantity"`
		UnitPrice      float64  `json:"unit_price"`
		Barcode        *string  `json:"barcode"`
		UnitOfMeasure  *string  `json:"unit_of_measure"`
		Description    *string  `json:"description"`
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	limit := 10  // default
	offset := 0  // default

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

	products, err := h.productService.List(ctx, tenantID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products": products,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetProductByID handles GET /products/:id
func (h *ProductHandlers) GetProductByID(c echo.Context) error {
	ctx := c.Request().Context()

	productID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	product, err := h.productService.GetByID(ctx, tenantID, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, product)
}

// UpdateProduct handles PUT /products/:id
func (h *ProductHandlers) UpdateProduct(c echo.Context) error {
	ctx := c.Request().Context()

	productID, err := h.validateUUID(c.Param("id"))
	if err != nil {
		return err
	}

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Name           string   `json:"name"`
		CategoryID     *string  `json:"category_id"`
		BatchNumber    *string  `json:"batch_number"`
		ExpiryDate     *string  `json:"expiry_date"`
		Quantity       int      `json:"quantity"`
		UnitPrice      float64  `json:"unit_price"`
		Barcode        *string  `json:"barcode"`
		UnitOfMeasure  *string  `json:"unit_of_measure"`
		Description    *string  `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := h.validateProduct(&req); err != nil {
		return err
	}

	existing, err := h.productService.GetByID(ctx, tenantID, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.productService.Delete(ctx, tenantID, productID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}

// SearchProducts handles GET /products/search
func (h *ProductHandlers) SearchProducts(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	analytics, err := h.productService.CategoryAnalytics(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"analytics":    analytics,
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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

	// Enhanced debugging for UUID parsing
	idParam := c.Param("id")
	log.Printf("DEBUG: GetProductImages called with ID param: '%s', length: %d", idParam, len(idParam))
	log.Printf("DEBUG: Context type: %T", ctx)

	// Check for unexpected characters
	for i, r := range idParam {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == '-') {
			log.Printf("DEBUG: Invalid character at position %d: '%c' (ASCII: %d)", i, r, r)
			return echo.NewHTTPError(http.StatusBadRequest, "UUID contains invalid characters")
		}
	}

	// Check if it's a valid UUID format first
	if _, parseErr := uuid.Parse(idParam); parseErr != nil {
		log.Printf("DEBUG: UUID format validation failed for: '%s', parse error: %v", idParam, parseErr)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	// Now try to parse it
	productID, err := h.validateUUID(idParam)
	if err != nil {
		log.Printf("DEBUG: UUID parsing failed for: '%s', error: %v", idParam, err)
		return echo.NewHTTPError(http.StatusBadRequest, "UUID could not be parsed")
	}

	log.Printf("DEBUG: UUID validation successful: %s", productID.String())

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	images, err := h.productService.GetProductImages(ctx, tenantID, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"images":   images,
		"count":    len(images),
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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
		"url":     url,
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

	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
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