package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRBACTemplateService_LoadTemplates(t *testing.T) {
	// This test requires the actual repos to be mocked
	// For now, we'll test the YAML loading logic independently

	service := NewRBACTemplateService(nil, nil, nil)

	err := service.LoadTemplates("../../config/role_templates.yaml")
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	// Verify templates were loaded
	names := service.GetTemplateNames()
	assert.NotEmpty(t, names, "Should have loaded templates")
	assert.Contains(t, names, "admin", "Should contain admin template")
	assert.Contains(t, names, "manager", "Should contain manager template")
	assert.Contains(t, names, "user", "Should contain user template")
	assert.Contains(t, names, "viewer", "Should contain viewer template")
}

func TestRBACTemplateService_GetTemplate(t *testing.T) {
	service := NewRBACTemplateService(nil, nil, nil)
	err := service.LoadTemplates("../../config/role_templates.yaml")
	assert.NoError(t, err)

	// Test getting admin template
	adminTemplate, err := service.GetTemplate("admin")
	assert.NoError(t, err)
	assert.NotNil(t, adminTemplate)
	assert.Contains(t, adminTemplate.Permissions, "*")

	// Test getting manager template
	managerTemplate, err := service.GetTemplate("manager")
	assert.NoError(t, err)
	assert.NotNil(t, managerTemplate)
	assert.Contains(t, managerTemplate.Permissions, "product.*")

	// Test non-existent template
	_, err = service.GetTemplate("nonexistent")
	assert.Error(t, err)
}

func TestRBACTemplateService_GetPermissionGroup(t *testing.T) {
	service := NewRBACTemplateService(nil, nil, nil)
	err := service.LoadTemplates("../../config/role_templates.yaml")
	assert.NoError(t, err)

	// Test getting permission group
	group, err := service.GetPermissionGroup("inventory_management")
	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Contains(t, group.Permissions, "inventory.*")

	// Test non-existent group
	_, err = service.GetPermissionGroup("nonexistent")
	assert.Error(t, err)
}

s := &rbacService{}

	tests := []struct {
		pattern    string
		permission string
		expected   bool
	}{
		// Full wildcard
		{"*", "product.list", true},
		{"*", "user.create", true},

		// Resource wildcard
		{"product.*", "product.list", true},
		{"product.*", "product.create", true},
		{"product.*", "user.list", false},

		// Action wildcard
		{"*.read", "product.read", true},
		{"*.read", "user.read", true},
		{"*.read", "product.list", false},

		// No wildcard (exact match required)
		{"user.list", "user.list", false}, // Should be exact match, not wildcard
		{"user.list", "user.create", false},

		// Invalid patterns
		{"product", "product.list", false},
		{"product.*.*", "product.list", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_vs_"+tt.permission, func(t *testing.T) {
			result := service.matchesWildcard(tt.pattern, tt.permission)
			assert.Equal(t, tt.expected, result, "Pattern: %s, Permission: %s", tt.pattern, tt.permission)
		})
	}
}
