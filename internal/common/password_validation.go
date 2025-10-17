package common

import (
	"agromart2/internal/validation"
)

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	return validation.ValidatePasswordSimple(password)
}

// ValidatePasswordWithDetails validates password and returns detailed results
func ValidatePasswordWithDetails(password string) validation.PasswordValidationResult {
	requirements := validation.DefaultPasswordRequirements()
	return validation.ValidatePassword(password, requirements)
}
