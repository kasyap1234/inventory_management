package services

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// SubscriptionService handles subscription-related business logic
type SubscriptionService interface {
	Create(ctx context.Context, tenantID uuid.UUID, planID string, customerEmail string) (*models.Subscription, error)
	GetByID(ctx context.Context, tenantID, subscriptionID uuid.UUID) (*models.Subscription, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Subscription, error)
	Cancel(ctx context.Context, tenantID, subscriptionID uuid.UUID) error
	Pause(ctx context.Context, tenantID, subscriptionID uuid.UUID) error
	Resume(ctx context.Context, tenantID, subscriptionID uuid.UUID) error
	UpdatePlan(ctx context.Context, tenantID, subscriptionID uuid.UUID, newPlanID string) error
	GetSubscriptionByRazorpayID(ctx context.Context, tenantID uuid.UUID, razorpayID string) (*models.Subscription, error)
	ValidateBilling(ctx context.Context, tenantID uuid.UUID, subscriptionID uuid.UUID) error
	GetAvailablePlans() map[string]PlanConfig
}

type subscriptionService struct {
	subscriptionRepo repositories.SubscriptionRepository
	razorpaySvc     RazorpayService
}

// PlanConfig represents a subscription plan configuration
type PlanConfig struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Interval    string  `json:"interval"` // monthly, yearly
	Features    []string `json:"features"`
}

// Predefined plans
var availablePlans = map[string]PlanConfig{
	"basic": {
		ID:          "basic",
		Name:        "Basic Plan",
		Description: "Essential features for small businesses",
		Amount:      999.0,
		Currency:    "INR",
		Interval:    "monthly",
		Features: []string{
			"Up to 5 warehouses",
			"Basic inventory management",
			"Invoice generation",
			"Email support",
		},
	},
	"premium": {
		ID:          "premium",
		Name:        "Premium Plan",
		Description: "Advanced features for growing businesses",
		Amount:      2499.0,
		Currency:    "INR",
		Interval:    "monthly",
		Features: []string{
			"Up to 20 warehouses",
			"Advanced analytics",
			"Multi-location inventory",
			"Priority support",
			"Custom branding",
			"API access",
		},
	},
	"enterprise": {
		ID:          "enterprise",
		Name:        "Enterprise Plan",
		Description: "Complete solution for large-scale operations",
		Amount:      4999.0,
		Currency:    "INR",
		Interval:    "monthly",
		Features: []string{
			"Unlimited warehouses",
			"Real-time analytics",
			"Advanced integrations",
			"24/7 phone support",
			"Custom development",
			"Dedicated account manager",
		},
	},
}

// NewSubscriptionService creates a new SubscriptionService instance
func NewSubscriptionService(
	subscriptionRepo repositories.SubscriptionRepository,
	razorpaySvc RazorpayService,
) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		razorpaySvc:     razorpaySvc,
	}
}

// Create creates a new subscription with Razorpay integration
func (s *subscriptionService) Create(ctx context.Context, tenantID uuid.UUID, planID string, customerEmail string) (*models.Subscription, error) {
	// Validate plan
	plan, exists := availablePlans[planID]
	if !exists {
		return nil, fmt.Errorf("invalid plan: %s", planID)
	}

	// Create subscription in Razorpay
	razorpayResp, err := s.razorpaySvc.CreateSubscription(ctx, plan.ID, tenantID, customerEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to create Razorpay subscription: %v", err)
	}

	// Set start date to now, end date to 30 days later for monthly plans
	now := time.Now()
	endDate := now.AddDate(0, 1, 0) // Monthly subscription

	// Create local subscription record
	subscription := &models.Subscription{
		ID:                     uuid.New(),
		TenantID:               tenantID,
		RazorpaySubscriptionID: &razorpayResp.ID,
		PlanName:               plan.Name,
		Amount:                 plan.Amount,
		Currency:               plan.Currency,
		Status:                 razorpayResp.Status,
		StartDate:              now,
		EndDate:                &endDate,
	}

	err = s.subscriptionRepo.Create(ctx, subscription)
	if err != nil {
		// TODO: Handle rollback of Razorpay subscription if database insert fails
		return nil, fmt.Errorf("failed to create subscription: %v", err)
	}

	return subscription, nil
}

// GetByID gets a subscription by ID
func (s *subscriptionService) GetByID(ctx context.Context, tenantID, subscriptionID uuid.UUID) (*models.Subscription, error) {
	return s.subscriptionRepo.GetByID(ctx, tenantID, subscriptionID)
}

// List lists subscriptions for a tenant
func (s *subscriptionService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Subscription, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.subscriptionRepo.List(ctx, tenantID, limit, offset)
}

// Cancel cancels a subscription
func (s *subscriptionService) Cancel(ctx context.Context, tenantID, subscriptionID uuid.UUID) error {
	// Get existing subscription
	subscription, err := s.subscriptionRepo.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return err
	}

	// Cancel in Razorpay if it exists
	if subscription.RazorpaySubscriptionID != nil {
		_, err = s.razorpaySvc.CancelSubscription(ctx, *subscription.RazorpaySubscriptionID)
		if err != nil {
			return fmt.Errorf("failed to cancel Razorpay subscription: %v", err)
		}
	}

	// Update local status
	subscription.Status = "cancelled"
	return s.subscriptionRepo.Update(ctx, subscription)
}

// Pause pauses a subscription
func (s *subscriptionService) Pause(ctx context.Context, tenantID, subscriptionID uuid.UUID) error {
	// Get existing subscription
	subscription, err := s.subscriptionRepo.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return err
	}

	// Pause in Razorpay if it exists
	if subscription.RazorpaySubscriptionID != nil {
		_, err = s.razorpaySvc.PauseSubscription(ctx, *subscription.RazorpaySubscriptionID)
		if err != nil {
			return fmt.Errorf("failed to pause Razorpay subscription: %v", err)
		}
	}

	// Update local status
	subscription.Status = "paused"
	return s.subscriptionRepo.Update(ctx, subscription)
}

// Resume resumes a subscription
func (s *subscriptionService) Resume(ctx context.Context, tenantID, subscriptionID uuid.UUID) error {
	// Get existing subscription
	subscription, err := s.subscriptionRepo.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return err
	}

	// Resume in Razorpay if it exists
	if subscription.RazorpaySubscriptionID != nil {
		_, err = s.razorpaySvc.ResumeSubscription(ctx, *subscription.RazorpaySubscriptionID)
		if err != nil {
			return fmt.Errorf("failed to resume Razorpay subscription: %v", err)
		}
	}

	// Update local status
	subscription.Status = "active"
	return s.subscriptionRepo.Update(ctx, subscription)
}

// UpdatePlan updates the subscription plan
func (s *subscriptionService) UpdatePlan(ctx context.Context, tenantID, subscriptionID uuid.UUID, newPlanID string) error {
	// Validate new plan
	plan, exists := availablePlans[newPlanID]
	if !exists {
		return fmt.Errorf("invalid plan: %s", newPlanID)
	}

	// Get existing subscription
	subscription, err := s.subscriptionRepo.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return err
	}

	// Update in Razorpay if it exists (construct updates map)
	if subscription.RazorpaySubscriptionID != nil {
		updates := map[string]interface{}{
			"plan_id":   plan.ID,
			"quantity":  1,
		}

		_, err = s.razorpaySvc.UpdateSubscription(ctx, *subscription.RazorpaySubscriptionID, updates)
		if err != nil {
			return fmt.Errorf("failed to update Razorpay subscription: %v", err)
		}
	}

	// Update local record
	subscription.PlanName = plan.Name
	subscription.Amount = plan.Amount
	return s.subscriptionRepo.Update(ctx, subscription)
}

// GetSubscriptionByRazorpayID gets subscription by Razorpay subscription ID
func (s *subscriptionService) GetSubscriptionByRazorpayID(ctx context.Context, tenantID uuid.UUID, razorpayID string) (*models.Subscription, error) {
	return s.subscriptionRepo.GetByRazorpayID(ctx, tenantID, razorpayID)
}

// ValidateBilling validates billing status for subscription
func (s *subscriptionService) ValidateBilling(ctx context.Context, tenantID uuid.UUID, subscriptionID uuid.UUID) error {
	subscription, err := s.subscriptionRepo.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %v", err)
	}

	if subscription.Status == "cancelled" {
		return fmt.Errorf("subscription is cancelled")
	}

	if subscription.Status == "paused" {
		return fmt.Errorf("subscription is paused")
	}

	if subscription.Status != "active" {
		return fmt.Errorf("subscription status is not active: %s", subscription.Status)
	}

	// Check if subscription has expired
	if subscription.EndDate != nil && subscription.EndDate.Before(time.Now()) {
		return fmt.Errorf("subscription has expired")
	}

	return nil
}

// GetAvailablePlans returns all available subscription plans
func (s *subscriptionService) GetAvailablePlans() map[string]PlanConfig {
	// Return a copy to prevent external modifications
	result := make(map[string]PlanConfig)
	for k, v := range availablePlans {
		result[k] = v
	}
	return result
}