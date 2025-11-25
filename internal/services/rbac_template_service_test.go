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
