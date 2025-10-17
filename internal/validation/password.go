package validation

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// PasswordStrength represents the strength level of a password
type PasswordStrength int

const (
	PasswordWeak PasswordStrength = iota
	PasswordMedium
	PasswordStrong
	PasswordVeryStrong
)

// PasswordValidationResult contains the result of password validation
type PasswordValidationResult struct {
	Valid    bool
	Strength PasswordStrength
	Score    int
	Errors   []string
	Warnings []string
}

// PasswordRequirements defines password requirements
type PasswordRequirements struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
	RequireSpecial   bool
	DisallowCommon   bool
}

// DefaultPasswordRequirements returns default password requirements
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:        8,
		MaxLength:        128,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   true,
		DisallowCommon:   true,
	}
}

// Common weak passwords to check against
var commonPasswords = map[string]bool{
	"password":   true,
	"12345678":   true,
	"123456789":  true,
	"qwerty":     true,
	"abc123":     true,
	"password1":  true,
	"password123": true,
	"admin":      true,
	"admin123":   true,
	"welcome":    true,
	"welcome123": true,
	"letmein":    true,
	"monkey":     true,
	"dragon":     true,
	"master":     true,
	"sunshine":   true,
	"princess":   true,
	"football":   true,
	"iloveyou":   true,
	"!@#$%^&*":   true,
}

// ValidatePassword validates a password against the given requirements
func ValidatePassword(password string, requirements PasswordRequirements) PasswordValidationResult {
	result := PasswordValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		Score:    0,
	}

	// Check length
	if len(password) < requirements.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must be at least "+string(rune(requirements.MinLength+'0'))+" characters long")
	} else if len(password) >= requirements.MinLength {
		result.Score += 1
	}

	if len(password) > requirements.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must not exceed "+string(rune(requirements.MaxLength+'0'))+" characters")
	}

	// Check character requirements
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}

	if requirements.RequireUppercase && !hasUpper {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one uppercase letter")
	} else if hasUpper {
		result.Score += 1
	}

	if requirements.RequireLowercase && !hasLower {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one lowercase letter")
	} else if hasLower {
		result.Score += 1
	}

	if requirements.RequireDigit && !hasDigit {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one digit")
	} else if hasDigit {
		result.Score += 1
	}

	if requirements.RequireSpecial && !hasSpecial {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one special character")
	} else if hasSpecial {
		result.Score += 1
	}

	// Check for common passwords
	if requirements.DisallowCommon {
		lowerPassword := strings.ToLower(password)
		if commonPasswords[lowerPassword] {
			result.Valid = false
			result.Errors = append(result.Errors, "Password is too common and easily guessable")
		}
	}

	// Check for sequential characters
	if hasSequentialCharacters(password) {
		result.Warnings = append(result.Warnings, "Password contains sequential characters")
		result.Score -= 1
	}

	// Check for repeated characters
	if hasRepeatedCharacters(password) {
		result.Warnings = append(result.Warnings, "Password contains repeated characters")
		result.Score -= 1
	}

	// Bonus points for length
	if len(password) >= 12 {
		result.Score += 1
	}
	if len(password) >= 16 {
		result.Score += 1
	}

	// Determine strength
	if result.Score >= 7 {
		result.Strength = PasswordVeryStrong
	} else if result.Score >= 5 {
		result.Strength = PasswordStrong
	} else if result.Score >= 3 {
		result.Strength = PasswordMedium
	} else {
		result.Strength = PasswordWeak
	}

	return result
}

// ValidatePasswordSimple performs basic password validation
func ValidatePasswordSimple(password string) error {
	requirements := DefaultPasswordRequirements()
	result := ValidatePassword(password, requirements)

	if !result.Valid {
		return errors.New(strings.Join(result.Errors, "; "))
	}

	return nil
}

// hasSequentialCharacters checks if password contains sequential characters
func hasSequentialCharacters(password string) bool {
	sequences := []string{
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij", "ijk", "jkl",
		"klm", "lmn", "mno", "nop", "opq", "pqr", "qrs", "rst", "stu", "tuv",
		"uvw", "vwx", "wxy", "xyz",
		"012", "123", "234", "345", "456", "567", "678", "789",
	}

	lowerPassword := strings.ToLower(password)
	for _, seq := range sequences {
		if strings.Contains(lowerPassword, seq) {
			return true
		}
	}

	return false
}

// hasRepeatedCharacters checks if password contains repeated characters
func hasRepeatedCharacters(password string) bool {
	repeated := regexp.MustCompile(`(.)\1{2,}`)
	return repeated.MatchString(password)
}

// GetPasswordStrengthDescription returns a human-readable description of password strength
func GetPasswordStrengthDescription(strength PasswordStrength) string {
	switch strength {
	case PasswordWeak:
		return "Weak"
	case PasswordMedium:
		return "Medium"
	case PasswordStrong:
		return "Strong"
	case PasswordVeryStrong:
		return "Very Strong"
	default:
		return "Unknown"
	}
}
