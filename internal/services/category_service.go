package services

import (
	"context"
	"errors"
	"fmt"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// CategoryService handles business logic for category hierarchies
type CategoryService struct {
	categoryRepo repositories.CategoryRepository
}

// NewCategoryService creates a new category service instance
func NewCategoryService(categoryRepo repositories.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// GetCategoryHierarchy retrieves the complete category hierarchy for a tenant
func (s *CategoryService) GetCategoryHierarchy(ctx context.Context, tenantID uuid.UUID) ([]*models.CategoryTree, error) {
	// Get all categories for the tenant
	categories, err := s.categoryRepo.List(ctx, tenantID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	// Build hierarchy tree
	hierarchy := s.buildHierarchyTree(categories)
	return hierarchy, nil
}

// GetCategoryPath retrieves the full path from root to a specific category
func (s *CategoryService) GetCategoryPath(ctx context.Context, tenantID uuid.UUID, categoryID uuid.UUID) ([]string, error) {
	category, err := s.categoryRepo.GetByID(ctx, tenantID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// Parse path string
	if category.Path == "" {
		return []string{category.Name}, nil
	}

	// Split path and append current category name
	pathParts := []string{}
	if category.Path != "" {
		pathParts = append(pathParts, category.Path)
	}
	pathParts = append(pathParts, category.Name)

	return pathParts, nil
}

// GetSubcategories retrieves all direct children of a category
func (s *CategoryService) GetSubcategories(ctx context.Context, tenantID uuid.UUID, parentID uuid.UUID) ([]*models.Category, error) {
	// Get all categories
	allCategories, err := s.categoryRepo.List(ctx, tenantID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	// Filter for direct children
	var subcategories []*models.Category
	for _, cat := range allCategories {
		if cat.ParentID != nil && *cat.ParentID == parentID {
			subcategories = append(subcategories, cat)
		}
	}

	return subcategories, nil
}

// GetRootCategories retrieves all root-level categories (no parent)
func (s *CategoryService) GetRootCategories(ctx context.Context, tenantID uuid.UUID) ([]*models.Category, error) {
	allCategories, err := s.categoryRepo.List(ctx, tenantID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	var rootCategories []*models.Category
	for _, cat := range allCategories {
		if cat.ParentID == nil {
			rootCategories = append(rootCategories, cat)
		}
	}

	return rootCategories, nil
}

// ValidateCategoryHierarchy checks for circular references and depth limits
func (s *CategoryService) ValidateCategoryHierarchy(ctx context.Context, tenantID uuid.UUID, categoryID uuid.UUID, newParentID *uuid.UUID) error {
	if newParentID == nil {
		return nil // No parent, valid
	}

	// Check if trying to set a category as its own parent
	if *newParentID == categoryID {
		return errors.New("category cannot be its own parent")
	}

	// Check for circular references
	visited := make(map[uuid.UUID]bool)
	currentID := *newParentID

	for {
		if visited[currentID] {
			return errors.New("circular reference detected in category hierarchy")
		}
		visited[currentID] = true

		parent, err := s.categoryRepo.GetByID(ctx, tenantID, currentID)
		if err != nil {
			return fmt.Errorf("invalid parent category: %w", err)
		}

		if parent.ParentID == nil {
			break // Reached root
		}

		currentID = *parent.ParentID
	}

	// Check depth limit (max 5 levels)
	depth := 0
	currentID = *newParentID
	for currentID != uuid.Nil {
		depth++
		if depth > 4 { // 0-indexed, so 4 means 5 levels
			return errors.New("category hierarchy depth exceeds maximum of 5 levels")
		}

		parent, err := s.categoryRepo.GetByID(ctx, tenantID, currentID)
		if err != nil {
			break
		}

		if parent.ParentID == nil {
			break
		}

		currentID = *parent.ParentID
	}

	return nil
}

// MoveCategoryToParent moves a category to a new parent
func (s *CategoryService) MoveCategoryToParent(ctx context.Context, tenantID uuid.UUID, categoryID uuid.UUID, newParentID *uuid.UUID) error {
	// Validate the move
	if err := s.ValidateCategoryHierarchy(ctx, tenantID, categoryID, newParentID); err != nil {
		return err
	}

	// Get the category
	category, err := s.categoryRepo.GetByID(ctx, tenantID, categoryID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Calculate new level and path
	var newLevel int
	var newPath string

	if newParentID == nil {
		newLevel = 0
		newPath = category.Name
	} else {
		parent, err := s.categoryRepo.GetByID(ctx, tenantID, *newParentID)
		if err != nil {
			return fmt.Errorf("parent category not found: %w", err)
		}

		newLevel = parent.Level + 1
		if parent.Path != "" {
			newPath = parent.Path + "/" + category.Name
		} else {
			newPath = category.Name
		}
	}

	// Update category
	category.ParentID = newParentID
	category.Level = newLevel
	category.Path = newPath

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// GetCategoryWithChildren retrieves a category with all its subcategories recursively
func (s *CategoryService) GetCategoryWithChildren(ctx context.Context, tenantID uuid.UUID, categoryID uuid.UUID) (*models.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, tenantID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// Get all categories to build tree
	allCategories, err := s.categoryRepo.List(ctx, tenantID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	// Build subcategories recursively
	s.attachSubcategories(category, allCategories)

	return category, nil
}

// Helper function to build hierarchy tree
func (s *CategoryService) buildHierarchyTree(categories []*models.Category) []*models.CategoryTree {
	// Create a map for quick lookup
	catMap := make(map[uuid.UUID]*models.Category)
	for _, cat := range categories {
		catMap[cat.ID] = cat
	}

	// Find root categories and build tree
	var roots []*models.CategoryTree
	for _, cat := range categories {
		if cat.ParentID == nil {
			tree := &models.CategoryTree{
				Category: *cat,
				Path:     []string{cat.Name},
				Depth:    0,
			}
			s.buildSubtree(tree, catMap)
			roots = append(roots, tree)
		}
	}

	return roots
}

// Helper function to build subtree recursively
func (s *CategoryService) buildSubtree(parent *models.CategoryTree, catMap map[uuid.UUID]*models.Category) {
	for _, cat := range catMap {
		if cat.ParentID != nil && *cat.ParentID == parent.ID {
			child := &models.CategoryTree{
				Category: *cat,
				Path:     append(parent.Path, cat.Name),
				Depth:    parent.Depth + 1,
			}
			s.buildSubtree(child, catMap)
			parent.Subcategories = append(parent.Subcategories, &child.Category)
		}
	}
}

// Helper function to attach subcategories recursively
func (s *CategoryService) attachSubcategories(parent *models.Category, allCategories []*models.Category) {
	for _, cat := range allCategories {
		if cat.ParentID != nil && *cat.ParentID == parent.ID {
			s.attachSubcategories(cat, allCategories)
			parent.Subcategories = append(parent.Subcategories, cat)
		}
	}
}

// GetCategoryStats returns statistics about category usage
func (s *CategoryService) GetCategoryStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	categories, err := s.categoryRepo.List(ctx, tenantID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	stats := map[string]interface{}{
		"total_categories": len(categories),
		"root_categories":  0,
		"max_depth":        0,
		"avg_depth":        0.0,
	}

	var totalDepth int
	maxDepth := 0

	for _, cat := range categories {
		if cat.ParentID == nil {
			stats["root_categories"] = stats["root_categories"].(int) + 1
		}

		if cat.Level > maxDepth {
			maxDepth = cat.Level
		}
		totalDepth += cat.Level
	}

	stats["max_depth"] = maxDepth
	if len(categories) > 0 {
		stats["avg_depth"] = float64(totalDepth) / float64(len(categories))
	}

	return stats, nil
}
