package services

import "strings"

// matchesWildcard checks if a permission matches a wildcard pattern
// Supports patterns like:
// - "*" (matches all)
// - "product.*" (matches all product permissions)
// - "*.read" (matches all read permissions)
func (s *rbacService) matchesWildcard(pattern, permission string) bool {
	// Full wildcard
	if pattern == "*" {
		return true
	}

	// No wildcard in pattern
	if !strings.Contains(pattern, "*") {
		return false
	}

	// Split both pattern and permission
	patternParts := strings.Split(pattern, ".")
	permParts := strings.Split(permission, ".")

	// Must have same number of parts
	if len(patternParts) != len(permParts) {
		return false
	}

	// Check each part
	for i := range patternParts {
		if patternParts[i] == "*" {
			continue // Wildcard matches anything
		}
		if patternParts[i] != permParts[i] {
			return false
		}
	}

	return true
}
