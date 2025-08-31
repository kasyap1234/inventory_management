package handlers

import (
	"log"
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CategoryHandlers handles category-related HTTP requests
type CategoryHandlers struct {
	categoryRepo  repositories.CategoryRepository
	rbacMiddleware *middleware.RBACMiddleware
}

// NewCategoryHandlers creates a new category handlers instance
func NewCategoryHandlers(categoryRepo repositories.CategoryRepository, rbacMiddleware *middleware.RBACMiddleware) *CategoryHandlers {
	return &CategoryHandlers{
		categoryRepo:  categoryRepo,
		rbacMiddleware: rbacMiddleware,
	}
}

// ListCategoriesRequest represents query parameters for listing categories
type ListCategoriesRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ListCategories handles getting a list of categories with tenant filtering
func (h *CategoryHandlers) ListCategories(c echo.Context) error {
	log.Printf("DEBUG: ListCategories handler called")

	// TODO: Enable RBAC for categories once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("categories:list")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err // RBAC middleware will return appropriate error
	// }

	ctx := c.Request().Context()
	log.Printf("DEBUG: ListCategories handler context retrieved")

	var req ListCategoriesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100 // Maximum limit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get categories from the tenant
	categories, err := h.categoryRepo.List(ctx, tenantID, req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list categories")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"categories": categories,
		"limit":      req.Limit,
		"offset":     req.Offset,
	})
}

// CreateCategoryRequest represents the category creation request payload
type CreateCategoryRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id"` // Optional parent category ID
}

// CreateCategory handles creating a new category
func (h *CategoryHandlers) CreateCategory(c echo.Context) error {
	// TODO: Enable RBAC for categories once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("categories:create")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err
	// }

	ctx := c.Request().Context()

	var req CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name is required")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Generate category ID
	categoryID := uuid.New()

	// Validate and parse parent_id if provided
	var parentID *uuid.UUID
	var level int
	var path string

	if req.ParentID != nil && *req.ParentID != "" {
		parentUUID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid parent_id format")
		}

		parentID = &parentUUID

		// Get parent category to calculate level and path
		parentCategory, err := h.categoryRepo.GetByID(ctx, tenantID, parentUUID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Parent category not found")
		}

		level = parentCategory.Level + 1
		path = parentCategory.Path + "/" + req.Name
	} else {
		// Root category
		level = 0
		path = req.Name
	}

	// Create new category
	category := &models.Category{
		ID:          categoryID,
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
		Level:       level,
		Path:        path,
	}

	if err := h.categoryRepo.Create(ctx, category); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create category")
	}

	return c.JSON(http.StatusCreated, category)
}

// GetCategory handles getting category details by ID
func (h *CategoryHandlers) GetCategory(c echo.Context) error {
	// TODO: Enable RBAC for categories once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("categories:read")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err
	// }

	ctx := c.Request().Context()

	categoryIDStr := c.Param("id")
	if categoryIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Category ID is required")
	}

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	log.Printf("DEBUG: ListCategories tenant ID: %s, ok: %v", tenantID.String(), ok)
	if !ok {
		log.Printf("DEBUG: ListCategories - tenant not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get category details
	category, err := h.categoryRepo.GetByID(ctx, tenantID, categoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Category not found")
	}

	return c.JSON(http.StatusOK, category)
}

// UpdateCategoryRequest represents the category update request payload
type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// UpdateCategory handles updating category details
func (h *CategoryHandlers) UpdateCategory(c echo.Context) error {
	// TODO: Enable RBAC for categories once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("categories:update")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err
	// }

	ctx := c.Request().Context()

	categoryIDStr := c.Param("id")
	if categoryIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Category ID is required")
	}

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID format")
	}

	var req UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get existing category
	category, err := h.categoryRepo.GetByID(ctx, tenantID, categoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Category not found")
	}

	// Update fields if provided
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}

	if err := h.categoryRepo.Update(ctx, category); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update category")
	}

	return c.JSON(http.StatusOK, category)
}

// SearchCategoriesRequest represents query parameters for searching categories
type SearchCategoriesRequest struct {
	Query  string `query:"q"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}

// SearchCategories handles searching categories by name or description
func (h *CategoryHandlers) SearchCategories(c echo.Context) error {
	// TODO: Enable RBAC for categories once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("categories:list")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err
	// }

	ctx := c.Request().Context()

	var req SearchCategoriesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100 // Maximum limit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Search categories
	categories, err := h.categoryRepo.Search(ctx, tenantID, req.Query, req.Limit, req.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search categories")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"categories": categories,
		"limit":      req.Limit,
		"offset":     req.Offset,
		"query":      req.Query,
	})
}

// DeleteCategory handles deleting a category
func (h *CategoryHandlers) DeleteCategory(c echo.Context) error {
	// TODO: Enable RBAC for categories once permissions are configured
	// err := h.rbacMiddleware.RequirePermission("categories:delete")(func(c echo.Context) error {
	// 	return nil
	// })(c)
	// if err != nil {
	// 	return err
	// }

	ctx := c.Request().Context()

	categoryIDStr := c.Param("id")
	if categoryIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Category ID is required")
	}

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID format")
	}

	// Get tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Optional: Check if category exists before deleting
	_, err = h.categoryRepo.GetByID(ctx, tenantID, categoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Category not found")
	}

	if err := h.categoryRepo.Delete(ctx, tenantID, categoryID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete category")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Category deleted successfully",
	})
}