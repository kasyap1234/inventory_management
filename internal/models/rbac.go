package models

import "github.com/google/uuid"

// RBACPermission represents a user permission used by RBAC checks (non-DB model)
type RBACPermission struct {
	Name       string                 `json:"name"`
	Resource   string                 `json:"resource"`
	Action     string                 `json:"action"`
	Conditions map[string]interface{} `json:"conditions"`
	RoleName   string                 `json:"role_name"`
}

// PermissionContext represents the context for permission checking
type PermissionContext struct {
	UserID     uuid.UUID              `json:"user_id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	Resource   string                 `json:"resource"`
	Action     string                 `json:"action"`
	ResourceID *uuid.UUID             `json:"resource_id,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
}

// PermissionInfo represents permission information for API responses
type PermissionInfo struct {
	Permissions []RBACPermission       `json:"permissions"`
	DataScope   map[string]interface{} `json:"data_scope"`
	Roles       []string               `json:"roles"`
}
