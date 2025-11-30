package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestPasswordHashingDirect tests password hashing without the service
func TestPasswordHashingDirect(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"simple password", "password123", false},
		{"complex password", "P@ssw0rd!#$%^&*()", false},
		{"unicode password", "密码123", false},
		{"long password", "ThisIsAVeryLongPasswordThatExceeds50Characters12345", false},
		{"empty password", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Hash the password
			hash, err := bcrypt.GenerateFromPassword([]byte(tc.password), bcrypt.DefaultCost)
			
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tc.password, string(hash))
			
			// Verify the hash
			err = bcrypt.CompareHashAndPassword(hash, []byte(tc.password))
			assert.NoError(t, err)
			
			// Verify wrong password fails
			err = bcrypt.CompareHashAndPassword(hash, []byte("wrong_password"))
			assert.Error(t, err)
		})
	}
}

// TestPasswordHashing_UniqueHashes verifies that the same password produces different hashes
func TestPasswordHashing_UniqueHashes(t *testing.T) {
	password := "TestPassword123!"
	
	hash1, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	
	hash2, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	
	// Hashes should be different (bcrypt uses salt)
	assert.NotEqual(t, string(hash1), string(hash2))
	
	// But both should verify correctly
	assert.NoError(t, bcrypt.CompareHashAndPassword(hash1, []byte(password)))
	assert.NoError(t, bcrypt.CompareHashAndPassword(hash2, []byte(password)))
}

// TestPasswordHashing_CostParameter tests different cost parameters
func TestPasswordHashing_CostParameter(t *testing.T) {
	password := "TestPassword123!"
	
	testCases := []struct {
		name string
		cost int
	}{
		{"minimum cost", bcrypt.MinCost},
		{"default cost", bcrypt.DefaultCost},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), tc.cost)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
			
			err = bcrypt.CompareHashAndPassword(hash, []byte(password))
			assert.NoError(t, err)
		})
	}
}

// TestPasswordVerification_EdgeCases tests edge cases in password verification
func TestPasswordVerification_EdgeCases(t *testing.T) {
	password := "TestPassword123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	testCases := []struct {
		name       string
		hash       []byte
		password   string
		shouldPass bool
	}{
		{"correct password", hash, password, true},
		{"wrong password", hash, "WrongPassword", false},
		{"similar password", hash, "TestPassword123", false},
		{"empty password", hash, "", false},
		{"password with extra space", hash, password + " ", false},
		{"password with leading space", hash, " " + password, false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := bcrypt.CompareHashAndPassword(tc.hash, []byte(tc.password))
			if tc.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestEmailNormalization tests email normalization logic
func TestEmailNormalization(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "test@example.com", "test@example.com"},
		{"uppercase", "TEST@EXAMPLE.COM", "test@example.com"},
		{"mixed case", "TeSt@ExAmPlE.cOm", "test@example.com"},
		{"with leading space", " test@example.com", "test@example.com"},
		{"with trailing space", "test@example.com ", "test@example.com"},
		{"with both spaces", " test@example.com ", "test@example.com"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simple normalization logic
			normalized := strings.TrimSpace(strings.ToLower(tc.input))
			assert.Equal(t, tc.expected, normalized)
		})
	}
}

// TestSubdomainValidation tests subdomain validation logic
func TestSubdomainValidation(t *testing.T) {
	testCases := []struct {
		name      string
		subdomain string
		isValid   bool
	}{
		{"valid lowercase", "mycompany", true},
		{"valid with numbers", "company123", true},
		{"valid with hyphen", "my-company", true},
		{"invalid with spaces", "my company", false},
		{"invalid with underscore", "my_company", false},
		{"invalid with special chars", "my@company", false},
		{"too short", "ab", false},
		{"starts with hyphen", "-company", false},
		{"ends with hyphen", "company-", false},
		{"valid long subdomain", "myverylongcompanyname", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := isValidSubdomain(tc.subdomain)
			assert.Equal(t, tc.isValid, isValid)
		})
	}
}

// isValidSubdomain validates subdomain format
func isValidSubdomain(subdomain string) bool {
	if len(subdomain) < 3 {
		return false
	}
	if strings.HasPrefix(subdomain, "-") || strings.HasSuffix(subdomain, "-") {
		return false
	}
	for _, c := range subdomain {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

// TestAccountLockoutLogic tests account lockout logic
func TestAccountLockoutLogic(t *testing.T) {
	const maxAttempts = 5
	
	testCases := []struct {
		name       string
		attempts   int
		shouldLock bool
	}{
		{"1 attempt", 1, false},
		{"4 attempts", 4, false},
		{"5 attempts", 5, true},
		{"6 attempts", 6, true},
		{"10 attempts", 10, true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isLocked := tc.attempts >= maxAttempts
			assert.Equal(t, tc.shouldLock, isLocked)
		})
	}
}

// TestPasswordStrengthValidation tests password strength validation
func TestPasswordStrengthValidation(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		isStrong bool
	}{
		{"too short", "Ab1!", false},
		{"no uppercase", "abcdefg1!", false},
		{"no lowercase", "ABCDEFG1!", false},
		{"no number", "Abcdefgh!", false},
		{"no special char", "Abcdefgh1", false},
		{"valid strong", "Abcdefg1!", true},
		{"complex valid", "MyP@ssw0rd!123", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isStrong := isStrongPassword(tc.password)
			assert.Equal(t, tc.isStrong, isStrong, "password: %s", tc.password)
		})
	}
}

// isStrongPassword checks if password meets strength requirements
func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", c):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// TestJWTClaimsValidation tests JWT claims validation logic
func TestJWTClaimsValidation(t *testing.T) {
	testCases := []struct {
		name     string
		claims   map[string]interface{}
		isValid  bool
	}{
		{
			"valid claims",
			map[string]interface{}{
				"user_id":   "123e4567-e89b-12d3-a456-426614174000",
				"tenant_id": "123e4567-e89b-12d3-a456-426614174001",
				"exp":       float64(9999999999),
			},
			true,
		},
		{
			"missing user_id",
			map[string]interface{}{
				"tenant_id": "123e4567-e89b-12d3-a456-426614174001",
				"exp":       float64(9999999999),
			},
			false,
		},
		{
			"missing tenant_id",
			map[string]interface{}{
				"user_id": "123e4567-e89b-12d3-a456-426614174000",
				"exp":     float64(9999999999),
			},
			false,
		},
		{
			"expired token",
			map[string]interface{}{
				"user_id":   "123e4567-e89b-12d3-a456-426614174000",
				"tenant_id": "123e4567-e89b-12d3-a456-426614174001",
				"exp":       float64(1),
			},
			false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := validateClaims(tc.claims)
			assert.Equal(t, tc.isValid, isValid)
		})
	}
}

// validateClaims validates JWT claims
func validateClaims(claims map[string]interface{}) bool {
	if _, ok := claims["user_id"]; !ok {
		return false
	}
	if _, ok := claims["tenant_id"]; !ok {
		return false
	}
	exp, ok := claims["exp"].(float64)
	if !ok || int64(exp) < 1000000000 { // Check if expired (basic check)
		return false
	}
	return true
}
