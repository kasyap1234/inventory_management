package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// PaymentService handles one-time payments (orders) backed by Razorpay.
type PaymentService interface {
	CreateOneTimePayment(ctx context.Context, tenantID uuid.UUID, amount float64, currency, receipt string, orderID *uuid.UUID, notes map[string]string) (*models.Payment, *CreateOrderResponse, error)
	ConfirmPayment(ctx context.Context, tenantID uuid.UUID, razorpayOrderID, paymentID, signature string) (*models.Payment, error)
	MarkPaymentStatus(ctx context.Context, tenantID uuid.UUID, razorpayOrderID, status string, paymentID *string, signature *string, paidAt *time.Time) error
}

type paymentService struct {
	paymentRepo repositories.PaymentRepository
	orderRepo   repositories.OrderRepository
	razorpaySvc RazorpayService
}

func NewPaymentService(paymentRepo repositories.PaymentRepository, orderRepo repositories.OrderRepository, razorpaySvc RazorpayService) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		razorpaySvc: razorpaySvc,
	}
}

// CreateOneTimePayment sets up a Razorpay order and persists the payment intent.
func (s *paymentService) CreateOneTimePayment(ctx context.Context, tenantID uuid.UUID, amount float64, currency, receipt string, orderID *uuid.UUID, notes map[string]string) (*models.Payment, *CreateOrderResponse, error) {
	if amount <= 0 {
		return nil, nil, fmt.Errorf("amount must be greater than zero")
	}
	if currency == "" {
		currency = "INR"
	}
	amountPaise := int64(math.Round(amount * 100))
	if amountPaise <= 0 {
		return nil, nil, fmt.Errorf("amount is too small after conversion to paise")
	}
	if receipt == "" {
		receipt = fmt.Sprintf("order_%s", uuid.New().String())
	}
	if notes == nil {
		notes = map[string]string{}
	}
	notes["tenant_id"] = tenantID.String()

	razorpayOrder, err := s.razorpaySvc.CreateOrder(ctx, amountPaise, currency, receipt, notes, true)
	if err != nil {
		return nil, nil, err
	}

	payment := &models.Payment{
		ID:              uuid.New(),
		TenantID:        tenantID,
		OrderID:         orderID,
		RazorpayOrderID: razorpayOrder.ID,
		Currency:        currency,
		AmountPaise:     amountPaise,
		Status:          razorpayOrder.Status,
		Receipt:         &receipt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, nil, fmt.Errorf("failed to persist payment: %w", err)
	}

	if orderID != nil {
		// Track payment state on the order for downstream consumers.
		_ = s.orderRepo.UpdatePaymentStatus(ctx, tenantID, *orderID, "pending")
	}

	return payment, razorpayOrder, nil
}

// ConfirmPayment verifies checkout signature and marks the payment as paid.
func (s *paymentService) ConfirmPayment(ctx context.Context, tenantID uuid.UUID, razorpayOrderID, paymentID, signature string) (*models.Payment, error) {
	if err := s.razorpaySvc.VerifyPaymentSignature(razorpayOrderID, paymentID, signature); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	now := time.Now()
	if err := s.paymentRepo.MarkPaid(ctx, tenantID, razorpayOrderID, &paymentID, &signature, now); err != nil {
		return nil, fmt.Errorf("failed to update payment state: %w", err)
	}

	p, err := s.paymentRepo.GetByRazorpayOrderID(ctx, tenantID, razorpayOrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload payment: %w", err)
	}

	if p.OrderID != nil {
		_ = s.orderRepo.UpdatePaymentStatus(ctx, tenantID, *p.OrderID, "paid")
	}

	return p, nil
}

// MarkPaymentStatus updates payment and associated order based on webhook signals.
func (s *paymentService) MarkPaymentStatus(ctx context.Context, tenantID uuid.UUID, razorpayOrderID, status string, paymentID *string, signature *string, paidAt *time.Time) error {
	if err := s.paymentRepo.UpdateStatus(ctx, tenantID, razorpayOrderID, status, paymentID, signature, paidAt); err != nil {
		return err
	}

	p, err := s.paymentRepo.GetByRazorpayOrderID(ctx, tenantID, razorpayOrderID)
	if err != nil {
		return nil // Already updated payment; skip propagating to order if lookup fails
	}

	if p.OrderID != nil {
		var orderStatus string
		switch status {
		case "paid", "captured":
			orderStatus = "paid"
		case "failed":
			orderStatus = "payment_failed"
		case "authorized":
			orderStatus = "authorized"
		default:
			orderStatus = status
		}
		_ = s.orderRepo.UpdatePaymentStatus(ctx, tenantID, *p.OrderID, orderStatus)
	}

	return nil
}
