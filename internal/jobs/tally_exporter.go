package jobs

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	
	"strconv"
	"strings"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type TallyExporter struct {
	invoiceRepo repositories.InvoiceRepository
	orderRepo   repositories.OrderRepository
	productRepo repositories.ProductRepository
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

func NewTallyExporter(invoiceRepo repositories.InvoiceRepository, orderRepo repositories.OrderRepository, productRepo repositories.ProductRepository) *TallyExporter {
	return &TallyExporter{
		invoiceRepo: invoiceRepo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (e *TallyExporter) ExportInvoicesForTenant(ctx context.Context, req ExportRequest) (*ExportResult, error) {
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
	log.Println("Starting daily Tally export job")

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