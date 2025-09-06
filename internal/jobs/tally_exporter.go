package jobs

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"

	"strconv"
	"strings"
	"time"

	"agromart2/internal/config"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type TallyExporter struct {
	invoiceRepo repositories.InvoiceRepository
	orderRepo   repositories.OrderRepository
	productRepo repositories.ProductRepository
	config      *config.TallyConfig
	mode        string
	apiClient   interface{} // *internal.TallyAPIClient - using interface{} to avoid import cycle
}

type ExportRequest struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Format    string    `json:"format"` // "csv" or "excel" (csv only for now)
}

type ExportResult struct {
	FileName        string
	FileContent     string
	RecordsExported int
}

// Setter for API client to avoid circular imports
func (e *TallyExporter) SetAPIClient(client interface{}) {
	e.apiClient = client
}

func NewTallyExporter(invoiceRepo repositories.InvoiceRepository, orderRepo repositories.OrderRepository, productRepo repositories.ProductRepository, cfg *config.TallyConfig) *TallyExporter {
	return &TallyExporter{
		invoiceRepo: invoiceRepo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		config:      cfg,
		mode:        cfg.ExportImport.Mode,
	}
}

func (e *TallyExporter) isRestMode() bool {
	return e.mode == "rest"
}

func (e *TallyExporter) ExportInvoicesForTenant(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	fmt.Printf("Starting invoice export in %s mode\n", e.mode)

	if e.isRestMode() {
		return e.exportInvoicesViaAPI(ctx, req)
	}

	return e.exportInvoicesViaCSV(ctx, req)
}

func (e *TallyExporter) exportInvoicesViaAPI(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Get invoices
	invoices, err := e.invoiceRepo.GetInvoicesByTenantAndDateRange(ctx, req.TenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	if len(invoices) == 0 {
		return &ExportResult{
			FileName:        "empty_export_api.txt",
			FileContent:     "No invoices found for the given period",
			RecordsExported: 0,
		}, nil
	}

	// Export each invoice via REST API
	exported := 0
	for _, invoice := range invoices {
		if e.apiClient != nil {
			// Call the API client export method using reflection to avoid circular imports
			if err := e.callExportInvoiceMethod(ctx, invoice); err != nil {
				log.Printf("Failed to export invoice %s: %v", invoice.ID, err)
				continue
			}
		}
		exported++
	}

	fileName := fmt.Sprintf("tally_api_export_%s_%s_%s.txt", req.TenantID.String(), req.StartDate, req.EndDate)

	return &ExportResult{
		FileName:        fileName,
		FileContent:     fmt.Sprintf("Successfully exported %d out of %d invoices via REST API", exported, len(invoices)),
		RecordsExported: exported,
	}, nil
}

// Helper method to call API client export invoice using interface{}
func (e *TallyExporter) callExportInvoiceMethod(ctx context.Context, invoice *models.Invoice) error {
	if client, ok := e.apiClient.(apiCaller); ok {
		return client.ExportInvoice(ctx, invoice, invoice.TotalAmount)
	}
	return fmt.Errorf("API client not properly configured")
}

// API caller interface to avoid reflection
type apiCaller interface {
	ExportInvoice(ctx context.Context, invoice *models.Invoice, totalAmount float64) error
}

func (e *TallyExporter) exportInvoicesViaCSV(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Get invoices
	invoices, err := e.invoiceRepo.GetInvoicesByTenantAndDateRange(ctx, req.TenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	if len(invoices) == 0 {
		return &ExportResult{
			FileName:        "empty_export.csv",
			FileContent:     "No invoices found for the given period",
			RecordsExported: 0,
		}, nil
	}

	// Generate CSV content
	csvContent, err := e.generateGSTComplianceCSV(invoices)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSV: %w", err)
	}

	fileName := fmt.Sprintf("tally_export_%s_%s_%s.csv", req.TenantID.String(), req.StartDate, req.EndDate)

	return &ExportResult{
		FileName:        fileName,
		FileContent:     csvContent,
		RecordsExported: len(invoices),
	}, nil
}

func (e *TallyExporter) generateGSTComplianceCSV(invoices []*models.Invoice) (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Write header
	header := []string{
		"Invoice No",
		"Invoice Date",
		"GSTIN/UIN",
		"Party Name",
		"HSN/SAC",
		"Taxable Value",
		"CESS Rate",
		"CGST Amount",
		"SGST Amount",
		"IGST Amount",
		"Total GST Amount",
		"Total Amount",
		"Place of Supply",
	}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	// Write records
	for _, invoice := range invoices {
		record := []string{
			fmt.Sprintf("INV-%s", invoice.ID.String()[:8]),
			invoice.IssuedDate.Format("02/01/2006"),
			nullToEmpty(invoice.GSTIN),
			"Customer", // Could be populated from order
			nullToEmpty(invoice.HSNSAC),
			nullFloatPointerToString(invoice.TaxableAmount),
			"0", // CESS Rate
			nullFloatPointerToString(invoice.CGST),
			nullFloatPointerToString(invoice.SGST),
			nullFloatPointerToString(invoice.IGST),
			fmt.Sprintf("%.2f", nullPointerSum(invoice.CGST, invoice.SGST, invoice.IGST)),
			fmt.Sprintf("%.2f", invoice.TotalAmount),
			"Maharashtra", // Fixed for example
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func (e *TallyExporter) ExportOrdersForTenant(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	fmt.Printf("Starting order export in %s mode\n", e.mode)

	if e.isRestMode() {
		return e.exportOrdersViaAPI(ctx, req)
	}

	return e.exportOrdersViaCSV(ctx, req)
}

func (e *TallyExporter) exportOrdersViaAPI(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Get orders
	orders, err := e.orderRepo.GetOrdersByTenantAndDateRange(ctx, req.TenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	if len(orders) == 0 {
		return &ExportResult{
			FileName:        "empty_orders_export_api.txt",
			FileContent:     "No orders found for the given period",
			RecordsExported: 0,
		}, nil
	}

	// Export each order via REST API
	exported := 0
	for _, order := range orders {
		if e.apiClient != nil {
			if err := e.callExportOrderMethod(ctx, order); err != nil {
				log.Printf("Failed to export order %s: %v", order.ID, err)
				continue
			}
		}
		exported++
	}

	fileName := fmt.Sprintf("tally_order_api_export_%s_%s_%s.txt", req.TenantID.String(), req.StartDate, req.EndDate)

	return &ExportResult{
		FileName:        fileName,
		FileContent:     fmt.Sprintf("Successfully exported %d out of %d orders via REST API", exported, len(orders)),
		RecordsExported: exported,
	}, nil
}

// Helper method to call API client export order using interface{}
func (e *TallyExporter) callExportOrderMethod(ctx context.Context, order *models.Order) error {
	if client, ok := e.apiClient.(orderCaller); ok {
		return client.ExportOrder(ctx, order)
	}
	return fmt.Errorf("API client not properly configured")
}

// Order caller interface to avoid reflection
type orderCaller interface {
	ExportOrder(ctx context.Context, order *models.Order) error
}

func (e *TallyExporter) exportOrdersViaCSV(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Get orders
	orders, err := e.orderRepo.GetOrdersByTenantAndDateRange(ctx, req.TenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	if len(orders) == 0 {
		return &ExportResult{
			FileName:        "empty_orders_export.csv",
			FileContent:     "No orders found for the given period",
			RecordsExported: 0,
		}, nil
	}

	// Generate CSV content for orders
	csvContent, err := e.generateOrderCSV(orders)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSV: %w", err)
	}

	fileName := fmt.Sprintf("tally_order_export_%s_%s_%s.csv", req.TenantID.String(), req.StartDate, req.EndDate)

	return &ExportResult{
		FileName:        fileName,
		FileContent:     csvContent,
		RecordsExported: len(orders),
	}, nil
}

func (e *TallyExporter) generateOrderCSV(orders []*models.Order) (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Write header
	header := []string{
		"Order No",
		"Order Date",
		"Order Type",
		"Product ID",
		"Quantity",
		"Unit Price",
		"Total Amount",
		"Status",
		"Supplier/Distributor ID",
		"Warehouse ID",
	}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	// Write records
	for _, order := range orders {
		record := []string{
			fmt.Sprintf("ORD-%s", order.ID.String()[:8]),
			order.OrderDate.Format("02/01/2006"),
			order.OrderType,
			order.ProductID.String(),
			strconv.Itoa(order.Quantity),
			fmt.Sprintf("%.2f", order.UnitPrice),
			fmt.Sprintf("%.2f", float64(order.Quantity)*order.UnitPrice),
			order.Status,
			nullUUIDPointerToString(order.SupplierID, order.DistributorID),
			order.WarehouseID.String(),
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// Scheduled export job
func (e *TallyExporter) DailyExportJob(ctx context.Context) error {
	fmt.Printf("Starting daily Tally export job in %s mode\n", e.mode)

	// Get current date -1 day for yesterday's data
	yesterday := time.Now().AddDate(0, 0, -1)
	startDate := yesterday.Format("2006-01-02")
	endDate := yesterday.Format("2006-01-02")

	// This would need tenant management to pump all tenants
	// For now, just log
	log.Printf("Daily export would export data from %s to %s", startDate, endDate)

	return nil
}

// Helper functions
func nullToEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nullFloatPointerToString(f *float64) string {
	if f == nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", *f)
}

func nullPointerSum(values ...*float64) float64 {
	sum := 0.0
	for _, v := range values {
		if v != nil {
			sum += *v
		}
	}
	return sum
}

func nullUUIDPointerToString(uuids ...*uuid.UUID) string {
	for _, u := range uuids {
		if u != nil {
			return u.String()
		}
	}
	return ""
}