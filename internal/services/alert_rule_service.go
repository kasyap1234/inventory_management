package services

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// AlertRuleService interface for alert rule operations
type AlertRuleService interface {
	CreateAlertRule(ctx context.Context, tenantID uuid.UUID, rule *models.AlertRule) error
	UpdateAlertRule(ctx context.Context, tenantID uuid.UUID, rule *models.AlertRule) error
	GetAlertRule(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.AlertRule, error)
	ListAlertRules(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.AlertRule, error)
	DeleteAlertRule(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
	EvaluateAlertRules(ctx context.Context, tenantID uuid.UUID, eventType string, data map[string]interface{}) error
	TestAlertRule(ctx context.Context, tenantID uuid.UUID, ruleID uuid.UUID, testData map[string]interface{}) (*AlertRuleTestResult, error)
}

// AlertRuleTestResult represents the result of alert rule testing
type AlertRuleTestResult struct {
	Success    bool                   `json:"success"`
	Triggered  bool                   `json:"triggered"`
	Conditions map[string]interface{} `json:"conditions"`
	Actions    []models.AlertAction   `json:"actions"`
	Errors     []string               `json:"errors,omitempty"`
	Evaluation map[string]bool        `json:"evaluation"`
}

// Condition operators
const (
	OperatorEquals             = "eq"
	OperatorNotEquals          = "ne"
	OperatorGreaterThan        = "gt"
	OperatorGreaterThanOrEqual = "gte"
	OperatorLessThan           = "lt"
	OperatorLessThanOrEqual    = "lte"
	OperatorContains           = "contains"
	OperatorNotContains        = "not_contains"
	OperatorIn                 = "in"
	OperatorNotIn              = "not_in"
	OperatorExists             = "exists"
	OperatorNotExists          = "not_exists"
)

// alertRuleService implements AlertRuleService
type alertRuleService struct {
	repository repositories.AlertRuleRepository
	logger     *common.StructuredLogger
}

// NewAlertRuleService creates a new alert rule service
func NewAlertRuleService(
	repository repositories.AlertRuleRepository,
	logger *common.StructuredLogger,
) AlertRuleService {
	return &alertRuleService{
		repository: repository,
		logger:     logger,
	}
}

// CreateAlertRule creates a new alert rule
func (s *alertRuleService) CreateAlertRule(ctx context.Context, tenantID uuid.UUID, rule *models.AlertRule) error {
	rule.TenantID = tenantID
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// Validate alert rule
	if err := s.validateAlertRule(rule); err != nil {
		return common.CreateValidationError("create_alert_rule", map[string]interface{}{
			"validation": err.Error(),
		})
	}

	// Create alert rule
	if err := s.repository.Create(ctx, rule); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to create alert rule", err, map[string]interface{}{
			"tenant_id":  tenantID,
			"rule_name":  rule.Name,
			"event_type": rule.EventType,
		})
		return common.CreateDatabaseError("create_alert_rule", err)
	}

	s.logger.InfoWithContext(ctx, "Alert rule created", map[string]interface{}{
		"rule_id":    rule.ID,
		"tenant_id":  tenantID,
		"rule_name":  rule.Name,
		"event_type": rule.EventType,
		"is_active":  rule.IsActive,
	})

	// Audit log
	common.AuditCreate(ctx, "alert_rule", rule.ID.String(), map[string]interface{}{
		"name":       rule.Name,
		"event_type": rule.EventType,
		"conditions": rule.Conditions,
		"actions":    rule.Actions,
		"is_active":  rule.IsActive,
	})

	return nil
}

// UpdateAlertRule updates an existing alert rule
func (s *alertRuleService) UpdateAlertRule(ctx context.Context, tenantID uuid.UUID, rule *models.AlertRule) error {
	// Get existing rule for audit logging
	existing, err := s.repository.GetByID(ctx, tenantID, rule.ID)
	if err != nil {
		return common.CreateDatabaseError("update_alert_rule", err)
	}

	rule.TenantID = tenantID
	rule.UpdatedAt = time.Now()

	// Validate alert rule
	if err := s.validateAlertRule(rule); err != nil {
		return common.CreateValidationError("update_alert_rule", map[string]interface{}{
			"validation": err.Error(),
		})
	}

	// Update alert rule
	if err := s.repository.Update(ctx, rule); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to update alert rule", err, map[string]interface{}{
			"rule_id":   rule.ID,
			"tenant_id": tenantID,
		})
		return common.CreateDatabaseError("update_alert_rule", err)
	}

	s.logger.InfoWithContext(ctx, "Alert rule updated", map[string]interface{}{
		"rule_id":   rule.ID,
		"tenant_id": tenantID,
		"rule_name": rule.Name,
	})

	// Audit log
	oldValues := map[string]interface{}{
		"name":       existing.Name,
		"event_type": existing.EventType,
		"conditions": existing.Conditions,
		"actions":    existing.Actions,
		"is_active":  existing.IsActive,
	}
	newValues := map[string]interface{}{
		"name":       rule.Name,
		"event_type": rule.EventType,
		"conditions": rule.Conditions,
		"actions":    rule.Actions,
		"is_active":  rule.IsActive,
	}
	common.AuditUpdate(ctx, "alert_rule", rule.ID.String(), oldValues, newValues)

	return nil
}

// GetAlertRule retrieves an alert rule by ID
func (s *alertRuleService) GetAlertRule(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.AlertRule, error) {
	rule, err := s.repository.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, common.CreateDatabaseError("get_alert_rule", err)
	}

	return rule, nil
}

// ListAlertRules retrieves alert rules
func (s *alertRuleService) ListAlertRules(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.AlertRule, error) {
	rules, err := s.repository.List(ctx, tenantID, eventType)
	if err != nil {
		return nil, common.CreateDatabaseError("list_alert_rules", err)
	}

	return rules, nil
}

// DeleteAlertRule deletes an alert rule
func (s *alertRuleService) DeleteAlertRule(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	// Get existing rule for audit logging
	existing, err := s.repository.GetByID(ctx, tenantID, id)
	if err != nil {
		return common.CreateDatabaseError("delete_alert_rule", err)
	}

	// Delete alert rule
	if err := s.repository.Delete(ctx, tenantID, id); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to delete alert rule", err, map[string]interface{}{
			"rule_id":   id,
			"tenant_id": tenantID,
		})
		return common.CreateDatabaseError("delete_alert_rule", err)
	}

	s.logger.InfoWithContext(ctx, "Alert rule deleted", map[string]interface{}{
		"rule_id":   id,
		"tenant_id": tenantID,
		"rule_name": existing.Name,
	})

	// Audit log
	common.AuditDelete(ctx, "alert_rule", id.String(), map[string]interface{}{
		"name":       existing.Name,
		"event_type": existing.EventType,
		"conditions": existing.Conditions,
		"actions":    existing.Actions,
	})

	return nil
}

// EvaluateAlertRules evaluates all active alert rules for an event
func (s *alertRuleService) EvaluateAlertRules(ctx context.Context, tenantID uuid.UUID, eventType string, data map[string]interface{}) error {
	// Get active rules for this event type
	rules, err := s.repository.GetActiveRulesForEvent(ctx, tenantID, eventType)
	if err != nil {
		return common.CreateDatabaseError("evaluate_alert_rules", err)
	}

	if len(rules) == 0 {
		return nil // No rules to evaluate
	}

	s.logger.DebugWithContext(ctx, "Evaluating alert rules", map[string]interface{}{
		"tenant_id":  tenantID,
		"event_type": eventType,
		"rule_count": len(rules),
	})

	// Evaluate each rule
	for _, rule := range rules {
		triggered, err := s.evaluateRule(rule, data)
		if err != nil {
			s.logger.ErrorWithContext(ctx, "Failed to evaluate alert rule", err, map[string]interface{}{
				"rule_id":   rule.ID,
				"rule_name": rule.Name,
				"tenant_id": tenantID,
			})
			continue
		}

		if triggered {
			// Update trigger status
			if err := s.repository.UpdateTriggerStatus(ctx, tenantID, rule.ID); err != nil {
				s.logger.ErrorWithContext(ctx, "Failed to update alert rule trigger status", err, map[string]interface{}{
					"rule_id": rule.ID,
				})
			}

			// Execute actions (this would integrate with notification delivery service)
			s.executeAlertActions(ctx, rule, data)

			s.logger.InfoWithContext(ctx, "Alert rule triggered", map[string]interface{}{
				"rule_id":      rule.ID,
				"rule_name":    rule.Name,
				"tenant_id":    tenantID,
				"event_type":   eventType,
				"action_count": len(rule.Actions),
			})
		}
	}

	return nil
}

// TestAlertRule tests an alert rule with sample data
func (s *alertRuleService) TestAlertRule(ctx context.Context, tenantID uuid.UUID, ruleID uuid.UUID, testData map[string]interface{}) (*AlertRuleTestResult, error) {
	// Get rule
	rule, err := s.repository.GetByID(ctx, tenantID, ruleID)
	if err != nil {
		return nil, common.CreateDatabaseError("test_alert_rule", err)
	}

	result := &AlertRuleTestResult{
		Conditions: rule.Conditions,
		Actions:    rule.Actions,
		Evaluation: make(map[string]bool),
	}

	// Evaluate rule
	triggered, err := s.evaluateRule(rule, testData)
	if err != nil {
		result.Success = false
		result.Errors = []string{err.Error()}
		return result, nil
	}

	result.Success = true
	result.Triggered = triggered

	// Add detailed evaluation results
	for conditionName, condition := range rule.Conditions {
		if conditionMap, ok := condition.(map[string]interface{}); ok {
			conditionResult, _ := s.evaluateCondition(conditionMap, testData)
			result.Evaluation[conditionName] = conditionResult
		}
	}

	return result, nil
}

// Helper methods

// validateAlertRule validates an alert rule
func (s *alertRuleService) validateAlertRule(rule *models.AlertRule) error {
	// Basic validation
	if err := rule.ValidateAlertRule(); err != nil {
		return err
	}

	// Validate conditions
	if err := s.validateConditions(rule.Conditions); err != nil {
		return fmt.Errorf("invalid conditions: %w", err)
	}

	// Validate actions
	if err := s.validateActions(rule.Actions); err != nil {
		return fmt.Errorf("invalid actions: %w", err)
	}

	return nil
}

// validateConditions validates alert rule conditions
func (s *alertRuleService) validateConditions(conditions map[string]interface{}) error {
	if len(conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}

	for name, condition := range conditions {
		conditionMap, ok := condition.(map[string]interface{})
		if !ok {
			return fmt.Errorf("condition '%s' must be an object", name)
		}

		// Validate required fields
		if _, exists := conditionMap["field"]; !exists {
			return fmt.Errorf("condition '%s' must have a 'field' property", name)
		}

		if _, exists := conditionMap["operator"]; !exists {
			return fmt.Errorf("condition '%s' must have an 'operator' property", name)
		}

		// Validate operator
		operator, ok := conditionMap["operator"].(string)
		if !ok {
			return fmt.Errorf("condition '%s' operator must be a string", name)
		}

		if !s.isValidOperator(operator) {
			return fmt.Errorf("condition '%s' has invalid operator: %s", name, operator)
		}

		// Validate value is present for operators that require it
		if s.operatorRequiresValue(operator) {
			if _, exists := conditionMap["value"]; !exists {
				return fmt.Errorf("condition '%s' with operator '%s' must have a 'value' property", name, operator)
			}
		}
	}

	return nil
}

// validateActions validates alert rule actions
func (s *alertRuleService) validateActions(actions []models.AlertAction) error {
	if len(actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	for i, action := range actions {
		if action.Type == "" {
			return fmt.Errorf("action %d must have a type", i)
		}

		if action.Target == "" {
			return fmt.Errorf("action %d must have a target", i)
		}

		// Validate action type
		validTypes := map[string]bool{
			"email":   true,
			"sms":     true,
			"webhook": true,
			"push":    true,
		}

		if !validTypes[action.Type] {
			return fmt.Errorf("action %d has invalid type: %s", i, action.Type)
		}
	}

	return nil
}

// evaluateRule evaluates a single alert rule
func (s *alertRuleService) evaluateRule(rule *models.AlertRule, data map[string]interface{}) (bool, error) {
	// Evaluate all conditions (AND logic)
	for _, condition := range rule.Conditions {
		conditionMap, ok := condition.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("invalid condition format")
		}

		result, err := s.evaluateCondition(conditionMap, data)
		if err != nil {
			return false, err
		}

		if !result {
			return false, nil // One condition failed, rule doesn't trigger
		}
	}

	return true, nil // All conditions passed
}

// evaluateCondition evaluates a single condition
func (s *alertRuleService) evaluateCondition(condition map[string]interface{}, data map[string]interface{}) (bool, error) {
	field, ok := condition["field"].(string)
	if !ok {
		return false, fmt.Errorf("condition field must be a string")
	}

	operator, ok := condition["operator"].(string)
	if !ok {
		return false, fmt.Errorf("condition operator must be a string")
	}

	// Get field value from data
	fieldValue := s.getFieldValue(data, field)

	// Handle existence operators
	if operator == OperatorExists {
		return fieldValue != nil, nil
	}
	if operator == OperatorNotExists {
		return fieldValue == nil, nil
	}

	// For other operators, we need a comparison value
	expectedValue, exists := condition["value"]
	if !exists && s.operatorRequiresValue(operator) {
		return false, fmt.Errorf("operator %s requires a value", operator)
	}

	// Evaluate based on operator
	switch operator {
	case OperatorEquals:
		return s.compareValues(fieldValue, expectedValue, "eq"), nil
	case OperatorNotEquals:
		return !s.compareValues(fieldValue, expectedValue, "eq"), nil
	case OperatorGreaterThan:
		return s.compareValues(fieldValue, expectedValue, "gt"), nil
	case OperatorGreaterThanOrEqual:
		return s.compareValues(fieldValue, expectedValue, "gte"), nil
	case OperatorLessThan:
		return s.compareValues(fieldValue, expectedValue, "lt"), nil
	case OperatorLessThanOrEqual:
		return s.compareValues(fieldValue, expectedValue, "lte"), nil
	case OperatorContains:
		return s.stringContains(fieldValue, expectedValue), nil
	case OperatorNotContains:
		return !s.stringContains(fieldValue, expectedValue), nil
	case OperatorIn:
		return s.valueInArray(fieldValue, expectedValue), nil
	case OperatorNotIn:
		return !s.valueInArray(fieldValue, expectedValue), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// getFieldValue extracts a field value from data using dot notation
func (s *alertRuleService) getFieldValue(data map[string]interface{}, field string) interface{} {
	parts := strings.Split(field, ".")
	current := data

	for i, part := range parts {
		if current == nil {
			return nil
		}

		if i == len(parts)-1 {
			// Last part, return the value
			return current[part]
		}

		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

// compareValues compares two values based on the comparison type
func (s *alertRuleService) compareValues(a, b interface{}, comparison string) bool {
	if a == nil || b == nil {
		return false
	}

	// Try numeric comparison first
	if numA, okA := s.toFloat64(a); okA {
		if numB, okB := s.toFloat64(b); okB {
			switch comparison {
			case "eq":
				return numA == numB
			case "gt":
				return numA > numB
			case "gte":
				return numA >= numB
			case "lt":
				return numA < numB
			case "lte":
				return numA <= numB
			}
		}
	}

	// Fall back to string comparison
	strA := fmt.Sprintf("%v", a)
	strB := fmt.Sprintf("%v", b)

	switch comparison {
	case "eq":
		return strA == strB
	case "gt":
		return strA > strB
	case "gte":
		return strA >= strB
	case "lt":
		return strA < strB
	case "lte":
		return strA <= strB
	}

	return false
}

// stringContains checks if a string contains another string
func (s *alertRuleService) stringContains(haystack, needle interface{}) bool {
	haystackStr := fmt.Sprintf("%v", haystack)
	needleStr := fmt.Sprintf("%v", needle)
	return strings.Contains(strings.ToLower(haystackStr), strings.ToLower(needleStr))
}

// valueInArray checks if a value is in an array
func (s *alertRuleService) valueInArray(value, array interface{}) bool {
	arrayValue := reflect.ValueOf(array)
	if arrayValue.Kind() != reflect.Slice && arrayValue.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < arrayValue.Len(); i++ {
		if s.compareValues(value, arrayValue.Index(i).Interface(), "eq") {
			return true
		}
	}

	return false
}

// toFloat64 converts a value to float64 if possible
func (s *alertRuleService) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// isValidOperator checks if an operator is valid
func (s *alertRuleService) isValidOperator(operator string) bool {
	validOperators := map[string]bool{
		OperatorEquals:             true,
		OperatorNotEquals:          true,
		OperatorGreaterThan:        true,
		OperatorGreaterThanOrEqual: true,
		OperatorLessThan:           true,
		OperatorLessThanOrEqual:    true,
		OperatorContains:           true,
		OperatorNotContains:        true,
		OperatorIn:                 true,
		OperatorNotIn:              true,
		OperatorExists:             true,
		OperatorNotExists:          true,
	}
	return validOperators[operator]
}

// operatorRequiresValue checks if an operator requires a comparison value
func (s *alertRuleService) operatorRequiresValue(operator string) bool {
	noValueOperators := map[string]bool{
		OperatorExists:    true,
		OperatorNotExists: true,
	}
	return !noValueOperators[operator]
}

// executeAlertActions executes the actions for a triggered alert
func (s *alertRuleService) executeAlertActions(ctx context.Context, rule *models.AlertRule, data map[string]interface{}) {
	for _, action := range rule.Actions {
		s.logger.InfoWithContext(ctx, "Executing alert action", map[string]interface{}{
			"rule_id":       rule.ID,
			"action_type":   action.Type,
			"action_target": action.Target,
		})

		// Publish alert action event for notification system to handle
		common.PublishEvent(ctx, "alert.action_triggered", map[string]interface{}{
			"tenant_id":     rule.TenantID,
			"alert_rule_id": rule.ID,
			"alert_name":    rule.Name,
			"action_type":   action.Type,
			"target":        action.Target,
			"template_id":   action.TemplateID,
			"custom_data":   action.CustomData,
			"event_data":    data,
		})
	}
}
