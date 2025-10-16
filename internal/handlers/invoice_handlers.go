package handlers

import (
	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// InvoiceHandlers handles HTTP requests for invoices
type InvoiceHandlers struct {
	invoiceService     services.InvoiceServiceInterface
	orderService       services.OrderServiceInterface
	productService     services.ProductService
	minioSvc           services.MinioService
	rbacMiddleware     *middleware.RBACMiddleware
	supplierService    services.SupplierService
	distributorService services.DistributorService
	tenantService      services.TenantService
}

// NewInvoiceHandlers creates a new invoice handlers instance
func NewInvoiceHandlers(
	invoiceService services.InvoiceServiceInterface,
	orderService services.OrderServiceInterface,
	productService services.ProductService,
	minioSvc services.MinioService,
	rbacMiddleware *middleware.RBACMiddleware,
	supplierService services.SupplierService,
	distributorService services.DistributorService,
	tenantService services.TenantService,
) *InvoiceHandlers {
	return &InvoiceHandlers{
		invoiceService:     invoiceService,
		orderService:       orderService,
		productService:     productService,
		minioSvc:           minioSvc,
		rbacMiddleware:     rbacMiddleware,
		supplierService:    supplierService,
		distributorService: distributorService,
		tenantService:      tenantService,
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
		return common.SendServerError(c, "Failed to retrieve order: "+err.Error())
	}

	if order == nil {
		return common.SendNotFoundError(c, "order")
	}

	if order.Status != "delivered" {
		return common.SendValidationError(c, "order_status",
			"Invoice can only be generated for orders with status 'delivered', current status: "+order.Status)
	}

	// Check if invoice already exists for this order
	existingInvoices, err := h.invoiceService.GetInvoicesByOrderID(ctx, tenantID, orderID)
	if err != nil {
		return common.SendServerError(c, "Failed to check existing invoices: "+err.Error())
	}

	if len(existingInvoices) > 0 {
		return common.SendClientError(c, "Invoice already exists for this order")
	}

	invoice := &models.Invoice{
		ID:         uuid.New(),
		TenantID:   tenantID,
		OrderID:    orderID,
		GSTIN:      req.GSTIN,
		Status:     "unpaid",
		IssuedDate: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Validate quantity and unit price for overflow protection
	if err := common.ValidateQuantityPrice(order.Quantity, order.UnitPrice); err != nil {
		return common.SendValidationError(c, "order_details", err.Error())
	}

	// Calculate total amount with overflow protection
	totalAmount, err := common.SafeMultiplyMonetary(float64(order.Quantity), order.UnitPrice)
	if err != nil {
		return common.SendValidationError(c, "total_amount", fmt.Sprintf("Amount calculation failed: %v", err))
	}

	// Apply GST calculation (assuming 18% GST for general goods)
	gstRate := 18.0
	invoice.GSTRate = &gstRate
	invoice.TaxableAmount = &totalAmount

	// Calculate CGST and SGST with overflow protection
	cgst, err := common.CalculateGST(totalAmount, 9.0) // 9% CGST
	if err != nil {
		return common.SendServerError(c, fmt.Sprintf("CGST calculation failed: %v", err))
	}

	sgst, err := common.CalculateGST(totalAmount, 9.0) // 9% SGST
	if err != nil {
		return common.SendServerError(c, fmt.Sprintf("SGST calculation failed: %v", err))
	}

	totalGST := cgst + sgst
	invoice.CGST = &cgst
	invoice.SGST = &sgst // FIX: Was nil, now correctly set to sgst
	invoice.TotalAmount = totalAmount + totalGST

	// Determine IGST for inter-state transactions (simplified logic)
	// In real implementation, this would be based on shipping addresses
	invoice.IGST = nil // Assuming intra-state for now

	if err := h.invoiceService.CreateInvoice(ctx, invoice); err != nil {
		return common.SendServerError(c, "Failed to create invoice: "+err.Error())
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

// BulkCreateInvoices handles POST /invoices/bulk-create
func (h *InvoiceHandlers) BulkCreateInvoices(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return common.SendUnauthorizedError(c)
	}

	var req struct {
		Invoices []struct {
			OrderID       string  `json:"order_id"`
			GSTIN         string  `json:"gstin"`
			HSNSAC        string  `json:"hsn_sac"`
			TaxableAmount float64 `json:"taxable_amount"`
			GSTRate       float64 `json:"gst_rate"`
			CGST          float64 `json:"cgst"`
			SGST          float64 `json:"sgst"`
			IGST          float64 `json:"igst"`
			TotalAmount   float64 `json:"total_amount"`
		} `json:"invoices"`
	}

	if err := c.Bind(&req); err != nil {
		return common.SendClientError(c, "Invalid request format")
	}

	if len(req.Invoices) == 0 {
		return common.SendClientError(c, "At least one invoice is required")
	}

	createdInvoices := []models.Invoice{}
	failedOrders := []string{}

	for _, invReq := range req.Invoices {
		orderID, err := common.ValidateUUID(invReq.OrderID, "order_id")
		if err != nil {
			failedOrders = append(failedOrders, invReq.OrderID)
			continue
		}

		// Check if invoice already exists for this order
		existingInvoices, err := h.invoiceService.GetInvoicesByOrderID(ctx, tenantID, orderID)
		if err == nil && len(existingInvoices) > 0 {
			failedOrders = append(failedOrders, invReq.OrderID)
			continue
		}

		// Create invoice
		gstin := invReq.GSTIN
		hsnSac := invReq.HSNSAC
		gstRate := invReq.GSTRate
		taxableAmount := invReq.TaxableAmount
		cgst := invReq.CGST
		sgst := invReq.SGST
		igst := invReq.IGST

		invoice := &models.Invoice{
			ID:            uuid.New(),
			TenantID:      tenantID,
			OrderID:       orderID,
			GSTIN:         &gstin,
			HSNSAC:        &hsnSac,
			TaxableAmount: &taxableAmount,
			GSTRate:       &gstRate,
			CGST:          &cgst,
			SGST:          &sgst,
			IGST:          &igst,
			TotalAmount:   invReq.TotalAmount,
			Status:        "unpaid",
			IssuedDate:    time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := h.invoiceService.CreateInvoice(ctx, invoice); err != nil {
			failedOrders = append(failedOrders, invReq.OrderID)
			continue
		}

		createdInvoices = append(createdInvoices, *invoice)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":          "Bulk invoice creation completed",
		"created_count":    len(createdInvoices),
		"failed_count":     len(failedOrders),
		"total_count":      len(req.Invoices),
		"created_invoices": createdInvoices,
		"failed_orders":    failedOrders,
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
	pdfBytes, err := h.invoiceService.GenerateInvoicePDF(ctx, invoice, order, tenantID)
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
		return common.SendServerError(c, "Failed to upload PDF to storage: "+err.Error())
	}

	// Generate presigned URL for download with error handling
	pdfURL, err := h.minioSvc.GetPresignedURL(bucketName, objectName, 24*time.Hour)
	if err != nil {
		return common.SendServerError(c, "Failed to generate download URL: "+err.Error())
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
