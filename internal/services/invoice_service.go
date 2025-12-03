package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"agromart2/internal/analytics"
	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jung-kurt/gofpdf"
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

	GenerateInvoicePDF(ctx context.Context, invoice *models.Invoice, order *models.Order, tenantID uuid.UUID) ([]byte, error)
	GenerateSubscriptionInvoicePDF(ctx context.Context, subscription *models.Subscription, paymentDetails map[string]interface{}) ([]byte, error)
	EmailSubscriptionInvoice(ctx context.Context, subscription *models.Subscription, paymentDetails map[string]interface{}) error
}

// InvoiceAnalytics holds invoice analytics data
type InvoiceAnalytics struct {
	TotalInvoices         int
	UnpaidInvoices        int
	PaidInvoices          int
	OverdueInvoices       int
	TotalInvoiceAmount    float64
	TotalGSTCollected     float64
	AvgInvoiceValue       float64
	PaymentCollectionRate float64
}

type invoiceService struct {
	invoiceRepo         repositories.InvoiceRepository
	orderRepo           repositories.OrderRepository
	analyticsSvc        *analytics.AnalyticsService
	db                  *pgxpool.Pool
	tenantService       TenantService
	productService      ProductService
	supplierService     SupplierService
	distributorService  DistributorService
	notificationService NotificationService
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(invoiceRepo repositories.InvoiceRepository, orderRepo repositories.OrderRepository, analyticsSvc *analytics.AnalyticsService, db *pgxpool.Pool, tenantService TenantService, productService ProductService, supplierService SupplierService, distributorService DistributorService, notificationService NotificationService) InvoiceServiceInterface {
	return &invoiceService{
		invoiceRepo:         invoiceRepo,
		orderRepo:           orderRepo,
		analyticsSvc:        analyticsSvc,
		db:                  db,
		tenantService:       tenantService,
		productService:      productService,
		supplierService:     supplierService,
		distributorService:  distributorService,
		notificationService: notificationService,
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
		"unpaid":    {"paid", "overdue", "cancelled"},
		"paid":      {}, // Cannot transition from paid
		"overdue":   {"paid", "cancelled"},
		"cancelled": {}, // Cannot transition from cancelled
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
		"unpaid":    true,
		"paid":      true,
		"overdue":   true,
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
// based on business and buyer locations
func (s *invoiceService) DetermineGSTType(ctx context.Context, tenantID, orderID uuid.UUID) (GSTType, error) {
	// Logic for GST type determination:
	// 1. Get tenant's business location state from tenant model
	// 2. Get buyer location state from distributor/supplier address
	// 3. Compare states: same state = intra-state (CGST+SGST), different state = inter-state (IGST)

	// NOTE: Full implementation requires State field in Tenant, Supplier, and Distributor models
	// When models are enhanced with State fields, implement the following logic:
	//
	// 1. Get tenant information for seller state
	//    tenant, err := tenantRepo.GetByID(ctx, tenantID)
	//    if err != nil { return GSTIntraState, nil }
	//
	// 2. Determine buyer state based on order type
	//    var buyerState string
	//    if order.OrderType == "purchase" && order.SupplierID != nil {
	//        supplier, _ := supplierRepo.GetByID(ctx, tenantID, *order.SupplierID)
	//        buyerState = supplier.State
	//    } else if order.OrderType == "sale" && order.DistributorID != nil {
	//        distributor, _ := distributorRepo.GetByID(ctx, tenantID, *order.DistributorID)
	//        buyerState = distributor.State
	//    }
	//
	// 3. Compare states to determine GST type
	//    if tenant.State != "" && buyerState != "" && tenant.State != buyerState {
	//        return GSTInterState, nil
	//    }
	//
	// For now, default to intra-state CGST+SGST for backward compatibility

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
	taxableAmount, err := common.SafeMultiplyMonetary(float64(order.Quantity), order.UnitPrice)
	if err != nil {
		return common.SecureErrorMessage("taxable amount calculation", err)
	}

	// Apply GST calculation with standard Indian GST rate (18%)
	gstRate := 18.0
	cgst, sgst, igst := s.CalculateGSTComponents(taxableAmount, gstRate, gstType)

	// Calculate total with overflow protection
	totalAmount := taxableAmount + cgst + sgst + igst
	if math.IsInf(totalAmount, 0) || math.IsNaN(totalAmount) {
		return common.SecureErrorMessage("total amount calculation", fmt.Errorf("calculated invoice total is invalid"))
	}

	// Generate invoice number
	issuedDate := time.Now()
	invoiceNumber, err := s.invoiceRepo.GenerateInvoiceNumber(ctx, tenantID, issuedDate)
	if err != nil {
		return common.SecureErrorMessage("generate invoice number", err)
	}

	// Calculate due date (30 days from issued date)
	dueDate := issuedDate.AddDate(0, 0, 30)

	// Get HSN/SAC code from product if available
	var hsnSac *string
	if order.ProductID != uuid.Nil {
		// We need to add a product repository to invoice service
		// For now, this will be a placeholder that returns nil
		// In production, inject ProductRepository into InvoiceService
		// and retrieve product.HSNSAC
		// product, err := s.productRepo.GetByID(ctx, tenantID, order.ProductID)
		// if err == nil && product != nil && product.HSNSAC != nil {
		//     hsnSac = product.HSNSAC
		// }
	}

	// Create invoice with GST details
	invoice := &models.Invoice{
		ID:            uuid.New(),
		TenantID:      tenantID,
		OrderID:       orderID,
		InvoiceNumber: invoiceNumber,
		HSNSAC:        hsnSac, // HSN/SAC code from product (currently placeholder)
		TaxableAmount: &taxableAmount,
		GSTRate:       &gstRate,
		CGST:          &cgst,
		SGST:          &sgst,
		IGST:          &igst,
		TotalAmount:   totalAmount,
		Status:        "unpaid",
		IssuedDate:    issuedDate,
		DueDate:       dueDate,
		CreatedAt:     issuedDate,
		UpdatedAt:     issuedDate,
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

// GenerateInvoicePDF creates a professional PDF invoice
func (s *invoiceService) GenerateInvoicePDF(ctx context.Context, invoice *models.Invoice, order *models.Order, tenantID uuid.UUID) ([]byte, error) {
	// Get tenant information for contact details
	tenant, err := s.tenantService.GetByID(ctx, tenantID)
	if err != nil || tenant == nil {
		// If tenant not found, create a minimal tenant object with defaults
		tenant = &models.Tenant{
			ID:   tenantID,
			Name: "Your Company",
		}
	}

	// Get product details for the order
	product, err := s.productService.GetByID(ctx, tenantID, order.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product details: %w", err)
	}

	// Create new PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set margins
	marginX := 20.0
	marginY := 20.0
	pdf.SetMargins(marginX, marginY, marginX)
	pdf.SetAutoPageBreak(true, marginY)

	// Set fonts
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(33, 37, 41) // Dark gray

	// Company header
	pdf.SetXY(marginX, marginY)
	pdf.Cell(0, 10, "AGROMART INVOICE")
	pdf.Ln(15)

	// Invoice details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, fmt.Sprintf("Invoice Number: %s", invoice.ID.String()))
	pdf.Ln(8)
	pdf.Cell(0, 8, fmt.Sprintf("Invoice Date: %s", invoice.IssuedDate.Format("02-Jan-2006")))
	pdf.Ln(8)
	pdf.Cell(0, 8, fmt.Sprintf("Order ID: %s", order.ID.String()))
	pdf.Ln(8)

	// GSTIN if provided
	if invoice.GSTIN != nil && *invoice.GSTIN != "" {
		pdf.Cell(0, 8, fmt.Sprintf("GSTIN: %s", *invoice.GSTIN))
		pdf.Ln(8)
	}

	pdf.Ln(5)

	// Billing Information section
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 8, "BILL TO:")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)

	// Get customer details from order
	customerName := "Customer"
	customerAddress := "Address not provided"
	customerContact := "Contact not provided"

	// Try to get supplier or distributor details
	if order.SupplierID != nil {
		if supplier, err := s.supplierService.GetByID(ctx, tenantID, *order.SupplierID); err == nil && supplier != nil {
			customerName = supplier.Name
			if supplier.Address != nil && *supplier.Address != "" {
				customerAddress = *supplier.Address
			}
			if supplier.ContactEmail != nil && *supplier.ContactEmail != "" {
				customerContact = "Email: " + *supplier.ContactEmail
			} else if supplier.ContactPhone != nil && *supplier.ContactPhone != "" {
				customerContact = "Phone: " + *supplier.ContactPhone
			}
		}
	} else if order.DistributorID != nil {
		if distributor, err := s.distributorService.GetByID(ctx, tenantID, *order.DistributorID); err == nil && distributor != nil {
			customerName = distributor.Name
			if distributor.Address != nil && *distributor.Address != "" {
				customerAddress = *distributor.Address
			}
			if distributor.ContactEmail != nil && *distributor.ContactEmail != "" {
				customerContact = "Email: " + *distributor.ContactEmail
			} else if distributor.ContactPhone != nil && *distributor.ContactPhone != "" {
				customerContact = "Phone: " + *distributor.ContactPhone
			}
		}
	}

	pdf.Cell(0, 6, customerName)
	pdf.Ln(6)
	pdf.Cell(0, 6, customerAddress)
	pdf.Ln(6)
	pdf.Cell(0, 6, customerContact)
	pdf.Ln(10)

	// Items table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240) // Light gray background

	// Table headers
	headers := []string{"Description", "Qty", "Rate", "Amount"}
	colWidths := []float64{80, 20, 30, 40}

	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(8)

	// Item row
	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(255, 255, 255) // White background

	description := product.Name
	if product.Description != nil && *product.Description != "" {
		description += "\n" + *product.Description
	}

	pdf.CellFormat(colWidths[0], 8, description, "1", 0, "L", false, 0, "")
	pdf.CellFormat(colWidths[1], 8, fmt.Sprintf("%d", order.Quantity), "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidths[2], 8, fmt.Sprintf("%.2f", order.UnitPrice), "1", 0, "R", false, 0, "")

	lineAmount, err := common.SafeMultiplyMonetary(float64(order.Quantity), order.UnitPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate line total: %w", err)
	}
	pdf.CellFormat(colWidths[3], 8, fmt.Sprintf("%.2f", lineAmount), "1", 0, "R", false, 0, "")
	pdf.Ln(8)

	// Empty rows for future multiple items
	for i := 0; i < 3; i++ {
		for j, width := range colWidths {
			border := "1"
			if j == len(colWidths)-1 {
				border = "1" // Last column
			}
			pdf.CellFormat(width, 8, "", border, 0, "C", false, 0, "")
		}
		pdf.Ln(8)
	}

	pdf.Ln(5)

	// GST and totals section
	pdf.SetFont("Arial", "B", 10)

	// Subtotal
	subtotal := lineAmount
	pdf.CellFormat(130, 6, "Subtotal:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 6, fmt.Sprintf("%.2f", subtotal), "", 0, "R", false, 0, "")
	pdf.Ln(6)

	// GST breakdown
	if invoice.CGST != nil && *invoice.CGST > 0 {
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(130, 5, "CGST (9%):", "", 0, "R", false, 0, "")
		pdf.CellFormat(40, 5, fmt.Sprintf("%.2f", *invoice.CGST), "", 0, "R", false, 0, "")
		pdf.Ln(5)
	}

	if invoice.SGST != nil && *invoice.SGST > 0 {
		pdf.CellFormat(130, 5, "SGST (9%):", "", 0, "R", false, 0, "")
		pdf.CellFormat(40, 5, fmt.Sprintf("%.2f", *invoice.SGST), "", 0, "R", false, 0, "")
		pdf.Ln(5)
	}

	if invoice.IGST != nil && *invoice.IGST > 0 {
		pdf.CellFormat(130, 5, "IGST (18%):", "", 0, "R", false, 0, "")
		pdf.CellFormat(40, 5, fmt.Sprintf("%.2f", *invoice.IGST), "", 0, "R", false, 0, "")
		pdf.Ln(5)
	}

	// Total
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(220, 20, 60) // Red color for total
	pdf.CellFormat(130, 8, "TOTAL:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", invoice.TotalAmount), "", 0, "R", false, 0, "")
	pdf.Ln(10)

	// Terms and conditions
	pdf.SetTextColor(33, 37, 41) // Reset to dark
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(0, 6, "Terms & Conditions:")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 8)
	terms := []string{
		"1. Payment is due within 30 days of invoice date",
		"2. Late payments may incur additional charges",
		"3. Goods once sold will not be taken back",
		"4. This is a computer generated invoice",
	}

	for _, term := range terms {
		pdf.Cell(0, 5, term)
		pdf.Ln(5)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128) // Gray
	pdf.Cell(0, 5, "Thank you for your business!")
	pdf.Ln(5)

	// Use tenant contact info if available, otherwise use placeholder
	contactInfo := "For any queries, contact our support team"
	if tenant.SupportEmail != nil && *tenant.SupportEmail != "" {
		contactInfo = fmt.Sprintf("For any queries, contact: %s", *tenant.SupportEmail)
		if tenant.SupportPhone != nil && *tenant.SupportPhone != "" {
			contactInfo += fmt.Sprintf(" | %s", *tenant.SupportPhone)
		}
	}
	pdf.Cell(0, 5, contactInfo)

	// Get PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
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

// GenerateSubscriptionInvoicePDF creates a PDF invoice for a subscription payment
func (s *invoiceService) GenerateSubscriptionInvoicePDF(ctx context.Context, subscription *models.Subscription, paymentDetails map[string]interface{}) ([]byte, error) {
	// Get tenant information
	tenant, err := s.tenantService.GetByID(ctx, subscription.TenantID)
	if err != nil || tenant == nil {
		tenant = &models.Tenant{
			ID:   subscription.TenantID,
			Name: "Valued Customer",
		}
	}

	// Create new PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set margins
	marginX := 20.0
	marginY := 20.0
	pdf.SetMargins(marginX, marginY, marginX)
	pdf.SetAutoPageBreak(true, marginY)

	// Set fonts
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(33, 37, 41) // Dark gray

	// Company header
	pdf.SetXY(marginX, marginY)
	pdf.Cell(0, 10, "AGROMART SUBSCRIPTION INVOICE")
	pdf.Ln(15)

	// Invoice details
	pdf.SetFont("Arial", "B", 12)

	// Extract payment ID if available
	paymentID := "N/A"
	if pid, ok := paymentDetails["payment_id"].(string); ok {
		paymentID = pid
	} else if pid, ok := paymentDetails["id"].(string); ok {
		paymentID = pid
	}

	pdf.Cell(0, 8, fmt.Sprintf("Invoice Number: %s", paymentID))
	pdf.Ln(8)
	pdf.Cell(0, 8, fmt.Sprintf("Date: %s", time.Now().Format("02-Jan-2006")))
	pdf.Ln(8)
	pdf.Cell(0, 8, fmt.Sprintf("Subscription ID: %s", subscription.ID.String()))
	pdf.Ln(8)

	pdf.Ln(5)

	// Billing Information section
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 8, "BILL TO:")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, tenant.Name)
	pdf.Ln(6)
	if tenant.ContactEmail != nil {
		pdf.Cell(0, 6, *tenant.ContactEmail)
		pdf.Ln(6)
	}
	pdf.Ln(10)

	// Items table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240) // Light gray background

	// Table headers
	headers := []string{"Description", "Period", "Amount"}
	colWidths := []float64{90, 50, 50}

	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(8)

	// Item row
	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(255, 255, 255) // White background

	description := fmt.Sprintf("Subscription Plan: %s", subscription.PlanName)
	period := "Monthly" // Default, could be derived from plan details

	amount := subscription.Amount
	// If payment details has amount, use that (might be different due to pro-ration or currency)
	if amt, ok := paymentDetails["amount"].(float64); ok {
		amount = amt / 100 // Razorpay amount is in paise
	}

	pdf.CellFormat(colWidths[0], 8, description, "1", 0, "L", false, 0, "")
	pdf.CellFormat(colWidths[1], 8, period, "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidths[2], 8, fmt.Sprintf("%.2f %s", amount, subscription.Currency), "1", 0, "R", false, 0, "")
	pdf.Ln(8)

	pdf.Ln(5)

	// Total
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(220, 20, 60) // Red color for total
	pdf.CellFormat(140, 8, "TOTAL PAID:", "", 0, "R", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%.2f %s", amount, subscription.Currency), "", 0, "R", false, 0, "")
	pdf.Ln(10)

	// Footer
	pdf.SetTextColor(33, 37, 41) // Reset to dark
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128) // Gray
	pdf.Cell(0, 5, "Thank you for your subscription!")
	pdf.Ln(5)
	pdf.Cell(0, 5, "This is a computer generated invoice.")

	// Get PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// EmailSubscriptionInvoice generates and emails a subscription invoice
func (s *invoiceService) EmailSubscriptionInvoice(ctx context.Context, subscription *models.Subscription, paymentDetails map[string]interface{}) error {
	// Generate PDF
	pdfBytes, err := s.GenerateSubscriptionInvoicePDF(ctx, subscription, paymentDetails)
	if err != nil {
		return fmt.Errorf("failed to generate invoice PDF: %w", err)
	}

	// Get tenant details for email
	tenant, err := s.tenantService.GetByID(ctx, subscription.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant details: %w", err)
	}
	if tenant == nil || tenant.ContactEmail == nil || *tenant.ContactEmail == "" {
		return fmt.Errorf("tenant contact email not found")
	}

	// Prepare email
	subject := fmt.Sprintf("Invoice for %s Subscription", subscription.PlanName)
	body := fmt.Sprintf(`
		<h1>Subscription Invoice</h1>
		<p>Dear %s,</p>
		<p>Thank you for your subscription payment. Please find the invoice attached.</p>
		<p>Plan: %s</p>
		<p>Amount: %.2f %s</p>
		<p>Best regards,<br>Agromart Team</p>
	`, tenant.Name, subscription.PlanName, subscription.Amount, subscription.Currency)

	attachmentName := fmt.Sprintf("invoice_%s.pdf", time.Now().Format("20060102"))

	// Send email
	if err := s.notificationService.SendEmailWithAttachment(ctx, subscription.TenantID, *tenant.ContactEmail, subject, body, attachmentName, pdfBytes); err != nil {
		return fmt.Errorf("failed to send invoice email: %w", err)
	}

	return nil
}
