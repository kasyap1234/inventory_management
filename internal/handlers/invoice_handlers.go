package handlers

import (
	"agromart2/internal/common"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/labstack/echo/v4"
)

// InvoiceHandlers handles HTTP requests for invoices
type InvoiceHandlers struct {
	invoiceService services.InvoiceServiceInterface
	orderService   services.OrderServiceInterface
	productService services.ProductService
	minioSvc       services.MinioService
}

// NewInvoiceHandlers creates a new invoice handlers instance
func NewInvoiceHandlers(invoiceService services.InvoiceServiceInterface, orderService services.OrderServiceInterface, productService services.ProductService, minioSvc services.MinioService) *InvoiceHandlers {
	return &InvoiceHandlers{
		invoiceService: invoiceService,
		orderService:   orderService,
		productService: productService,
		minioSvc:       minioSvc,
	}
}

// CreateInvoice handles POST /invoices
// Auto-generates invoice upon order completion
func (h *InvoiceHandlers) CreateInvoice(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return common.SendUnauthorizedError(c)
	}

	var req struct {
		OrderID string  `json:"order_id"`
		GSTIN   *string `json:"gstin"`
	}

	if err := c.Bind(&req); err != nil {
		return common.SendClientError(c, "Invalid request format")
	}

	orderID, err := common.ValidateUUID(req.OrderID, "order_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	// Validate GSTIN if provided
	if req.GSTIN != nil && common.SafeString(req.GSTIN) != "" {
		if err := common.ValidateGSTIN(common.SafeString(req.GSTIN), "gstin"); err != nil {
			return common.SendValidationError(c, "gstin", err.Error())
		}
	}

	// Verify order exists and is in deliverable state
	order, err := h.orderService.GetOrderByID(ctx, tenantID, orderID)
	if err != nil {
		return common.SendServerError(c, "Failed to retrieve order: " + err.Error())
	}

	if order == nil {
		return common.SendNotFoundError(c, "order")
	}

	if order.Status != "delivered" {
		return common.SendValidationError(c, "order_status",
			"Invoice can only be generated for orders with status 'delivered', current status: " + order.Status)
	}

	// Check if invoice already exists for this order
	existingInvoices, err := h.invoiceService.GetInvoicesByOrderID(ctx, tenantID, orderID)
	if err != nil {
		return common.SendServerError(c, "Failed to check existing invoices: " + err.Error())
	}

	if len(existingInvoices) > 0 {
		return common.SendClientError(c, "Invoice already exists for this order")
	}

	invoice := &models.Invoice{
		ID:             uuid.New(),
		TenantID:       tenantID,
		OrderID:        orderID,
		GSTIN:          req.GSTIN,
		Status:         "unpaid",
		IssuedDate:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Calculate GST based on order details with null safety
	if order.Quantity <= 0 || order.UnitPrice <= 0 {
		return common.SendValidationError(c, "order_details",
			"Invalid order quantity or unit price for invoice calculation")
	}

	totalAmount := float64(order.Quantity) * order.UnitPrice
	invoice.TotalAmount = totalAmount

	// Apply GST calculation (assuming 18% GST for general goods)
	gstRate := 18.0
	invoice.GSTRate = &gstRate
	invoice.TaxableAmount = &totalAmount

	cgst := totalAmount * 0.09 // 9% CGST
	sgst := totalAmount * 0.09 // 9% SGST
	totalGST := cgst + sgst
	invoice.CGST = &cgst
	invoice.SGST = nil
	invoice.TotalAmount = totalAmount + totalGST

	// Determine IGST for inter-state transactions (simplified logic)
	// In real implementation, this would be based on shipping addresses
	invoice.IGST = nil // Assuming intra-state for now

	if err := h.invoiceService.CreateInvoice(ctx, invoice); err != nil {
		return common.SendServerError(c, "Failed to create invoice: " + err.Error())
	}

	return c.JSON(http.StatusCreated, invoice)
}

// GetInvoices handles GET /invoices
func (h *InvoiceHandlers) GetInvoices(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return common.SendUnauthorizedError(c)
	}

	limit := 10
	offset := 0

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	invoices, err := h.invoiceService.ListInvoices(ctx, tenantID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"invoices": invoices,
		"limit":    limit,
		"offset":   offset,
	})
}

// ListInvoices handles GET /invoices (alias for GetInvoices)
func (h *InvoiceHandlers) ListInvoices(c echo.Context) error {
	return h.GetInvoices(c)
}

// GetInvoiceByID handles GET /invoices/:id
func (h *InvoiceHandlers) GetInvoiceByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid invoice ID")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return common.SendUnauthorizedError(c)
	}

	invoice, err := h.invoiceService.GetInvoiceByID(ctx, tenantID, invoiceID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if invoice == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Invoice not found")
	}

	return c.JSON(http.StatusOK, invoice)
}

// GetInvoice handles GET /invoices/:id (alias for GetInvoiceByID)
func (h *InvoiceHandlers) GetInvoice(c echo.Context) error {
	return h.GetInvoiceByID(c)
}

// UpdateInvoiceStatus handles PUT /invoices/:id/status
func (h *InvoiceHandlers) UpdateInvoiceStatus(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid invoice ID")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return common.SendUnauthorizedError(c)
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Status != "unpaid" && req.Status != "paid" && req.Status != "overdue" && req.Status != "cancelled" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid status. Must be unpaid, paid, overdue, or cancelled")
	}

	if err := h.invoiceService.UpdateInvoiceStatus(ctx, tenantID, invoiceID, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Invoice status updated successfully",
	})
}

// UpdateInvoice handles PUT /invoices/:id
func (h *InvoiceHandlers) UpdateInvoice(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid invoice ID")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return common.SendUnauthorizedError(c)
	}

	var req struct {
		Status string  `json:"status"`
		GSTIN  *string `json:"gstin"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Update status if provided
	if req.Status != "" {
		if req.Status != "unpaid" && req.Status != "paid" && req.Status != "overdue" && req.Status != "cancelled" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid status. Must be unpaid, paid, overdue, or cancelled")
		}
		if err := h.invoiceService.UpdateInvoiceStatus(ctx, tenantID, invoiceID, req.Status); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	// Get the current invoice to update GSTIN
	invoice, err := h.invoiceService.GetInvoiceByID(ctx, tenantID, invoiceID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if invoice == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Invoice not found")
	}

	if req.GSTIN != nil {
		invoice.GSTIN = req.GSTIN
		invoice.UpdatedAt = time.Now()
		if err := h.invoiceService.UpdateInvoice(ctx, invoice); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Invoice updated successfully",
	})
}

// DeleteInvoice handles DELETE /invoices/:id
func (h *InvoiceHandlers) DeleteInvoice(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid invoice ID")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return common.SendUnauthorizedError(c)
	}

	// Check if invoice exists and can be deleted
	invoice, err := h.invoiceService.GetInvoiceByID(ctx, tenantID, invoiceID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if invoice == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Invoice not found")
	}

	// Only allow deletion of unpaid invoices
	if invoice.Status != "unpaid" {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot delete invoice with status: "+invoice.Status)
	}

	if err := h.invoiceService.DeleteInvoice(ctx, tenantID, invoiceID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Invoice deleted successfully",
	})
}

// GetUnpaidInvoices handles GET /invoices/unpaid
func (h *InvoiceHandlers) GetUnpaidInvoices(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return common.SendUnauthorizedError(c)
	}

	limit := 10
	offset := 0

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	invoices, err := h.invoiceService.GetUnpaidInvoices(ctx, tenantID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"invoices": invoices,
		"limit":    limit,
		"offset":   offset,
	})
}

// generateInvoicePDF creates a professional PDF invoice
func (h *InvoiceHandlers) generateInvoicePDF(ctx context.Context, invoice *models.Invoice, order *models.Order, tenantID uuid.UUID) ([]byte, error) {
	// Get product details for the order
	product, err := h.productService.GetByID(ctx, tenantID, order.ProductID)
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
	pdf.Cell(0, 6, "Agromart Customer")
	pdf.Ln(6)
	pdf.Cell(0, 6, "Address: To be configured")
	pdf.Ln(6)
	pdf.Cell(0, 6, "Contact: support@agromart.com")
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
	pdf.CellFormat(colWidths[3], 8, fmt.Sprintf("%.2f", float64(order.Quantity)*order.UnitPrice), "1", 0, "R", false, 0, "")
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
	subtotal := float64(order.Quantity) * order.UnitPrice
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
	pdf.Cell(0, 5, "For any queries, contact: support@agromart.com | +91-XXXXXXXXXX")

	// Get PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateInvoicePDF handles POST /invoices/:id/generate-pdf
// Generates and stores PDF invoice using MinIO
func (h *InvoiceHandlers) GenerateInvoicePDF(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid invoice ID")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return common.SendUnauthorizedError(c)
	}

	invoice, err := h.invoiceService.GetInvoiceByID(ctx, tenantID, invoiceID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if invoice == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Invoice not found")
	}

	// Get the associated order details
	order, err := h.orderService.GetOrderByID(ctx, tenantID, invoice.OrderID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if order == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Order not found for this invoice")
	}

	// Generate PDF bytes with comprehensive error handling
	pdfBytes, err := h.generateInvoicePDF(ctx, invoice, order, tenantID)
	if err != nil {
		return common.SendServerError(c, fmt.Sprintf("Failed to generate PDF: %v", err))
	}

	// Validate PDF was generated successfully
	if len(pdfBytes) == 0 {
		return common.SendServerError(c, "Generated PDF is empty")
	}

	// Store PDF in MinIO with retry logic consideration
	bucketName := "invoices"
	objectName := fmt.Sprintf("%s-%s.pdf", tenantID.String(), invoiceID.String())

	err = h.minioSvc.UploadImage(ctx, bucketName, objectName, bytes.NewReader(pdfBytes), int64(len(pdfBytes)))
	if err != nil {
		return common.SendServerError(c, "Failed to upload PDF to storage: " + err.Error())
	}

	// Generate presigned URL for download with error handling
	pdfURL, err := h.minioSvc.GetPresignedURL(bucketName, objectName, 24*time.Hour)
	if err != nil {
		return common.SendServerError(c, "Failed to generate download URL: " + err.Error())
	}

	// Validate the URL was generated
	if pdfURL == "" {
		return common.SendServerError(c, "Generated download URL is empty")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "PDF generated and uploaded successfully",
		"pdf_url":    pdfURL,
		"expires_in": "24 hours",
	})
}