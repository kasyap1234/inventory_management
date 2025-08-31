package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"agromart2/internal/analytics"
	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InvoiceServiceInterface defines the interface for invoice service
type InvoiceServiceInterface interface {
	CreateInvoice(ctx context.Context, invoice *models.Invoice) error
	GetInvoiceByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*models.Invoice, error)
	ListInvoices(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invoice, error)
	UpdateInvoice(ctx context.Context, invoice *models.Invoice) error
	DeleteInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID) error
	UpdateInvoiceStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status string) error
	GetInvoicesByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.Invoice, error)
	GetUnpaidInvoices(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invoice, error)

	// Business logic methods
	CalculateGST(orderTotal float64, gstRate float64) (cgst, sgst, igst float64)
	AutoGenerateInvoiceOnDelivery(ctx context.Context, tenantID, orderID uuid.UUID) error
	MarkOverdueInvoices(ctx context.Context, tenantID uuid.UUID) error
	CalculateInvoiceAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*InvoiceAnalytics, error)
}

// InvoiceAnalytics holds invoice analytics data
type InvoiceAnalytics struct {
	TotalInvoices        int
	UnpaidInvoices      int
	PaidInvoices        int
	OverdueInvoices     int
	TotalInvoiceAmount  float64
	TotalGSTCollected   float64
	AvgInvoiceValue     float64
	PaymentCollectionRate float64
}

type invoiceService struct {
	invoiceRepo repositories.InvoiceRepository
	orderRepo   repositories.OrderRepository
	analyticsSvc *analytics.AnalyticsService
	db          *pgxpool.Pool
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(invoiceRepo repositories.InvoiceRepository, orderRepo repositories.OrderRepository, analyticsSvc *analytics.AnalyticsService, db *pgxpool.Pool) InvoiceServiceInterface {
	return &invoiceService{
		invoiceRepo: invoiceRepo,
		orderRepo:   orderRepo,
		analyticsSvc: analyticsSvc,
		db:          db,
	}
}

// validateInvoiceFinancialData validates financial data in invoices
func (s *invoiceService) validateInvoiceFinancialData(invoice *models.Invoice) error {
	// Validate total amount (required)
	if invoice.TotalAmount <= 0 {
		return fmt.Errorf("total amount must be positive")
	}
	if invoice.TotalAmount > 10000000.00 {
		return fmt.Errorf("total amount cannot exceed ₹1,00,00,000")
	}

	// Validate taxable amount if provided
	if invoice.TaxableAmount != nil {
		if *invoice.TaxableAmount <= 0 {
			return fmt.Errorf("taxable amount must be positive")
		}
		if *invoice.TaxableAmount > 10000000.00 {
			return fmt.Errorf("taxable amount cannot exceed ₹1,00,00,000")
		}
	}

	// Validate GST rate if provided
	if invoice.GSTRate != nil {
		if *invoice.GSTRate < 0 || *invoice.GSTRate > 100 {
			return fmt.Errorf("GST rate must be between 0 and 100")
		}
	}

	// Validate GST components if provided
	if invoice.CGST != nil && *invoice.CGST < 0 {
		return fmt.Errorf("CGST cannot be negative")
	}
	if invoice.SGST != nil && *invoice.SGST < 0 {
		return fmt.Errorf("SGST cannot be negative")
	}
	if invoice.IGST != nil && *invoice.IGST < 0 {
		return fmt.Errorf("IGST cannot be negative")
	}

	// Validate financial consistency
	if invoice.TaxableAmount != nil && invoice.CGST != nil && invoice.SGST != nil {
		expectedTotal := *invoice.TaxableAmount + *invoice.CGST + *invoice.SGST
		if invoice.IGST != nil {
			expectedTotal += *invoice.IGST
		}
		if expectedTotal < 0 {
			return fmt.Errorf("calculated total would cause overflow")
		}
	}

	return nil
}

// CreateInvoice creates a new invoice with enhanced security and validation
func (s *invoiceService) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	// Validate GSTIN if provided
	if invoice.GSTIN != nil {
		gstinVal := common.SafeString(invoice.GSTIN)
		if gstinVal != "" {
			if err := common.ValidateGSTIN(gstinVal, "GSTIN"); err != nil {
				return common.SecureErrorMessage("GSTIN validation", err)
			}
		}
	}

	// Validate HSN/SAC code if provided
	if invoice.HSNSAC != nil {
		hsnVal := common.SafeString(invoice.HSNSAC)
		if hsnVal != "" && len(hsnVal) > 6 {
			return common.SecureErrorMessage("HSN/SAC validation", fmt.Errorf("HSN/SAC must be 6 characters or less"))
		}
		*invoice.HSNSAC = hsnVal
	}

	// Validate and sanitize financial data
	if err := s.validateInvoiceFinancialData(invoice); err != nil {
		return common.SecureErrorMessage("financial data validation", err)
	}

	invoice.CreatedAt = time.Now()
	invoice.UpdatedAt = time.Now()

	// Generate invoice number if not provided
	if invoice.InvoiceNumber == "" {
		invoiceNumber, err := s.invoiceRepo.GenerateInvoiceNumber(ctx, invoice.TenantID, invoice.IssuedDate)
		if err != nil {
			return common.SecureErrorMessage("generate invoice number", err)
		}
		invoice.InvoiceNumber = invoiceNumber
	}

	// Set due date if not provided
	if invoice.DueDate.IsZero() {
		invoice.DueDate = invoice.IssuedDate.AddDate(0, 0, 30) // 30 days from issued date
	}

	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return common.SecureErrorMessage("create invoice", err)
	}

	// Update analytics asynchronously
	s.updateAnalytics(ctx, invoice.TenantID)

	return nil
}

// GetInvoiceByID retrieves an invoice by ID
func (s *invoiceService) GetInvoiceByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*models.Invoice, error) {
	return s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
}

// ListInvoices retrieves invoices with pagination
func (s *invoiceService) ListInvoices(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invoice, error) {
	return s.invoiceRepo.List(ctx, tenantID, limit, offset)
}

// UpdateInvoice updates an invoice
func (s *invoiceService) UpdateInvoice(ctx context.Context, invoice *models.Invoice) error {
	invoice.UpdatedAt = time.Now()
	return s.invoiceRepo.Update(ctx, invoice)
}

// DeleteInvoice deletes an invoice
func (s *invoiceService) DeleteInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID) error {
	return s.invoiceRepo.Delete(ctx, tenantID, invoiceID)
}

// isValidStatusTransition validates invoice status transitions
func (s *invoiceService) isValidStatusTransition(currentStatus, newStatus string) bool {
	// Define valid status transitions
	validTransitions := map[string][]string{
		"unpaid":     {"paid", "overdue", "cancelled"},
		"paid":       {}, // Cannot transition from paid
		"overdue":    {"paid", "cancelled"},
		"cancelled":  {}, // Cannot transition from cancelled
	}

	allowed, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	// Check if newStatus is in the allowed list
	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}

// UpdateInvoiceStatus updates invoice status and triggers analytics updates
func (s *invoiceService) UpdateInvoiceStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"unpaid":   true,
		"paid":     true,
		"overdue":  true,
		"cancelled": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s. Must be one of: unpaid, paid, overdue, cancelled", status)
	}

	// Get current invoice for status transition validation
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return common.SecureErrorMessage("get invoice for status update", err)
	}
	if invoice == nil {
		return fmt.Errorf("invoice not found")
	}

	// Validate status transitions
	if !s.isValidStatusTransition(invoice.Status, status) {
		return fmt.Errorf("invalid status transition from %s to %s", invoice.Status, status)
	}

	// If changing to paid, set paid_date
	if status == "paid" {
		now := time.Now()
		invoice.Status = status
		invoice.PaidDate = &now
		invoice.UpdatedAt = now

		if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
			return common.SecureErrorMessage("update invoice with paid date", err)
		}
	} else {
		// For other statuses, just update status
		if err := s.invoiceRepo.UpdateInvoiceStatus(ctx, tenantID, invoiceID, status); err != nil {
			return common.SecureErrorMessage("update invoice status", err)
		}
	}

	// Update analytics asynchronously
	s.updateAnalytics(ctx, tenantID)

	return nil
}

// GetInvoicesByOrderID retrieves invoices for a specific order
func (s *invoiceService) GetInvoicesByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.Invoice, error) {
	return s.invoiceRepo.GetInvoicesByOrderID(ctx, tenantID, orderID)
}

// GetUnpaidInvoices retrieves unpaid invoices
func (s *invoiceService) GetUnpaidInvoices(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invoice, error) {
	return s.invoiceRepo.GetUnpaidInvoices(ctx, tenantID, limit, offset)
}

// GSTType represents the type of GST applicable
type GSTType int

const (
	GSTIntraState GSTType = iota // CGST + SGST
	GSTInterState                // IGST
)

// CalculateGSTComponents calculates GST components for a given amount, rate, and GST type
func (s *invoiceService) CalculateGSTComponents(amount float64, gstRate float64, gstType GSTType) (cgst, sgst, igst float64) {
	if amount < 0 {
		return 0, 0, 0 // Prevent negative calculations
	}
	if gstRate < 0 || gstRate > 100 {
		gstRate = 18.0 // Default safe rate
	}

	gstAmount := amount * (gstRate / 100)

	switch gstType {
	case GSTIntraState:
		// CGST and SGST are each half of GST
		cgst = gstAmount / 2
		sgst = gstAmount / 2
		igst = 0
	case GSTInterState:
		// IGST is full GST amount
		cgst = 0
		sgst = 0
		igst = gstAmount
	default:
		// Default to intra-state for backward compatibility
		cgst = gstAmount / 2
		sgst = gstAmount / 2
		igst = 0
	}

	return cgst, sgst, igst
}

// CalculateGST calculates GST components based on invoice total and rate
// Deprecated: Use CalculateGSTComponents for better control over GST type
func (s *invoiceService) CalculateGST(orderTotal float64, gstRate float64) (cgst, sgst, igst float64) {
	return s.CalculateGSTComponents(orderTotal, gstRate, GSTIntraState)
}

// DetermineGSTType determines whether GST should be intra-state or inter-state
// For now, defaults to intra-state. In future, should be based on business and buyer locations
func (s *invoiceService) DetermineGSTType(ctx context.Context, tenantID, orderID uuid.UUID) (GSTType, error) {
	// TODO: Implement logic based on:
	// 1. Business/tenant location (from tenant model enhancement needed)
	// 2. Buyer location (from distributor/supplier addresses)
	// 3. Shipping destination (from order model enhancement)

	// For now, default to intra-state for backward compatibility
	return GSTIntraState, nil
}

// AutoGenerateInvoiceOnDelivery automatically creates invoice when order is delivered
func (s *invoiceService) AutoGenerateInvoiceOnDelivery(ctx context.Context, tenantID, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return common.SecureErrorMessage("retrieve order for invoice generation", err)
	}

	if order.Status != "delivered" {
		return common.SecureErrorMessage("invoice generation eligibility", fmt.Errorf("order must be delivered to generate invoice"))
	}

	// Check if invoice already exists
	existingInvoices, err := s.GetInvoicesByOrderID(ctx, tenantID, orderID)
	if err != nil {
		return common.SecureErrorMessage("check existing invoices", err)
	}

	if len(existingInvoices) > 0 {
		return common.SecureErrorMessage("invoice uniqueness check", fmt.Errorf("invoice already exists for this order"))
	}

	// Determine GST type based on business and buyer locations
	gstType, err := s.DetermineGSTType(ctx, tenantID, orderID)
	if err != nil {
		return common.SecureErrorMessage("determine GST type", err)
	}

	// Validate order data for financial calculations
	if order.Quantity <= 0 || order.UnitPrice <= 0 {
		return common.SecureErrorMessage("order data validation", fmt.Errorf("invalid order data for invoice generation"))
	}

	// Calculate totals with overflow protection
	taxableAmount := float64(order.Quantity) * order.UnitPrice
	if taxableAmount < 0 {
		return common.SecureErrorMessage("taxable amount calculation", fmt.Errorf("negative taxable amount"))
	}

	// Apply GST calculation with standard Indian GST rate (18%)
	gstRate := 18.0
	cgst, sgst, igst := s.CalculateGSTComponents(taxableAmount, gstRate, gstType)

	// Calculate total with overflow protection
	totalAmount := taxableAmount + cgst + sgst + igst

	// Generate invoice number
	issuedDate := time.Now()
	invoiceNumber, err := s.invoiceRepo.GenerateInvoiceNumber(ctx, tenantID, issuedDate)
	if err != nil {
		return common.SecureErrorMessage("generate invoice number", err)
	}

	// Calculate due date (30 days from issued date)
	dueDate := issuedDate.AddDate(0, 0, 30)

	// Create invoice with GST details
	invoice := &models.Invoice{
		ID:             uuid.New(),
		TenantID:       tenantID,
		OrderID:        orderID,
		InvoiceNumber:  invoiceNumber,
		HSNSAC:         nil, // TODO: Get from product HSN/SAC code
		TaxableAmount:  &taxableAmount,
		GSTRate:        &gstRate,
		CGST:           &cgst,
		SGST:           &sgst,
		IGST:           &igst,
		TotalAmount:    totalAmount,
		Status:         "unpaid",
		IssuedDate:     issuedDate,
		DueDate:        dueDate,
		CreatedAt:      issuedDate,
		UpdatedAt:      issuedDate,
	}

	return s.CreateInvoice(ctx, invoice)
}

// MarkOverdueInvoices marks unpaid invoices as overdue if past due date
func (s *invoiceService) MarkOverdueInvoices(ctx context.Context, tenantID uuid.UUID) error {
	// Validate date range before processing
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	oneYearAgo := time.Now().AddDate(-1, 0, 0)

	invoices, err := s.invoiceRepo.GetInvoicesByTenantAndDateRange(ctx, tenantID, oneYearAgo, thirtyDaysAgo)
	if err != nil {
		return common.SecureErrorMessage("retrieve invoices for overdue marking", err)
	}

	for _, invoice := range invoices {
		if invoice.Status == "unpaid" && time.Now().After(invoice.DueDate) {
			if err := s.UpdateInvoiceStatus(ctx, tenantID, invoice.ID, "overdue"); err != nil {
				log.Printf("Failed to mark invoice %s as overdue: %v", invoice.ID, common.SecureErrorMessage("update overdue status", err))
			}
		}
	}

	return nil
}

// CalculateInvoiceAnalytics generates comprehensive invoice analytics
func (s *invoiceService) CalculateInvoiceAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*InvoiceAnalytics, error) {
	// Validate date range
	if err := common.ValidateDateRange(startDate, endDate); err != nil {
		return nil, common.SecureErrorMessage("validate analytics date range", err)
	}

	invoices, err := s.invoiceRepo.GetInvoicesByTenantAndDateRange(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, common.SecureErrorMessage("retrieve invoices for analytics", err)
	}

	analytics := &InvoiceAnalytics{}

	for _, invoice := range invoices {
		analytics.TotalInvoices++
		analytics.TotalInvoiceAmount += invoice.TotalAmount

		switch invoice.Status {
		case "unpaid":
			analytics.UnpaidInvoices++
		case "paid":
			analytics.PaidInvoices++
		case "overdue":
			analytics.OverdueInvoices++
		case "cancelled":
			// Cancelled invoices are not counted in active metrics
		}

		// Calculate GST collected with null checks
		if invoice.CGST != nil {
			analytics.TotalGSTCollected += *invoice.CGST
		}
		if invoice.SGST != nil {
			analytics.TotalGSTCollected += *invoice.SGST
		}
		if invoice.IGST != nil {
			analytics.TotalGSTCollected += *invoice.IGST
		}
	}

	// Calculate averages with division by zero protection
	if analytics.TotalInvoices > 0 {
		analytics.AvgInvoiceValue = analytics.TotalInvoiceAmount / float64(analytics.TotalInvoices)

		totalProcessed := analytics.PaidInvoices + analytics.OverdueInvoices
		if totalProcessed > 0 {
			analytics.PaymentCollectionRate = float64(analytics.PaidInvoices) / float64(totalProcessed) * 100
		}
	}

	return analytics, nil
}

// updateAnalytics updates invoice analytics asynchronously
func (s *invoiceService) updateAnalytics(ctx context.Context, tenantID uuid.UUID) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in invoice analytics update: %v", r)
			}
		}()

		_, err := s.analyticsSvc.CalculateTenantAnalytics(context.Background(), tenantID)
		if err != nil {
			log.Printf("Failed to update invoice analytics: %v", common.SecureErrorMessage("analytics update", err))
		}
	}()
}