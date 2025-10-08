package common

import (
	"fmt"
	"math"
)

// MaxMonetaryValue is the maximum allowed monetary value (10 billion)
const MaxMonetaryValue = 10000000000.0

// MaxQuantity is the maximum allowed quantity
const MaxQuantity = 1000000

// ValidateMonetaryAmount validates monetary amounts for overflow protection
func ValidateMonetaryAmount(amount float64, fieldName string) error {
	if amount < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}
	if math.IsInf(amount, 0) || math.IsNaN(amount) {
		return fmt.Errorf("%s contains invalid value", fieldName)
	}
	if amount > MaxMonetaryValue {
		return fmt.Errorf("%s exceeds maximum allowed value (₹10,000,000,000)", fieldName)
	}
	return nil
}

// SafeMultiplyMonetary safely multiplies monetary values with overflow check
func SafeMultiplyMonetary(a, b float64) (float64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}

	// Check for potential overflow before multiplication
	if a > 0 && b > 0 {
		if a > MaxMonetaryValue/b {
			return 0, fmt.Errorf("monetary calculation would overflow maximum value")
		}
	}

	result := a * b

	if math.IsInf(result, 0) || math.IsNaN(result) {
		return 0, fmt.Errorf("monetary calculation overflow")
	}

	if result > MaxMonetaryValue {
		return 0, fmt.Errorf("result exceeds maximum monetary value (₹10,000,000,000)")
	}

	return result, nil
}

// ValidateQuantityPrice validates quantity and price before calculations
func ValidateQuantityPrice(quantity int, unitPrice float64) error {
	if quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	if quantity == 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	if quantity > MaxQuantity {
		return fmt.Errorf("quantity exceeds maximum allowed (1,000,000)")
	}

	if err := ValidateMonetaryAmount(unitPrice, "unit_price"); err != nil {
		return err
	}

	// Validate result won't overflow
	_, err := SafeMultiplyMonetary(float64(quantity), unitPrice)
	if err != nil {
		return fmt.Errorf("calculation would result in overflow: %w", err)
	}

	return nil
}

// ValidateGSTRate validates GST percentage rates
func ValidateGSTRate(rate float64, fieldName string) error {
	if rate < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}
	if rate > 100 {
		return fmt.Errorf("%s cannot exceed 100%%", fieldName)
	}
	if math.IsInf(rate, 0) || math.IsNaN(rate) {
		return fmt.Errorf("%s contains invalid value", fieldName)
	}
	return nil
}

// CalculateGST calculates GST amount with overflow protection
func CalculateGST(amount float64, gstRate float64) (float64, error) {
	if err := ValidateMonetaryAmount(amount, "amount"); err != nil {
		return 0, err
	}
	if err := ValidateGSTRate(gstRate, "gst_rate"); err != nil {
		return 0, err
	}

	// Calculate GST: amount * (gstRate / 100)
	gstAmount, err := SafeMultiplyMonetary(amount, gstRate/100)
	if err != nil {
		return 0, fmt.Errorf("GST calculation overflow: %w", err)
	}

	return gstAmount, nil
}

// ValidateBulkUpdateAmount validates amounts in bulk operations
func ValidateBulkUpdateAmount(amount *float64, fieldName string) error {
	if amount == nil {
		return nil // Optional field
	}
	return ValidateMonetaryAmount(*amount, fieldName)
}

// ValidateBulkUpdateQuantity validates quantities in bulk operations
func ValidateBulkUpdateQuantity(quantity *int, fieldName string) error {
	if quantity == nil {
		return nil // Optional field
	}
	if *quantity < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}
	if *quantity > MaxQuantity {
		return fmt.Errorf("%s exceeds maximum allowed (1,000,000)", fieldName)
	}
	return nil
}
