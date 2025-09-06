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

type TallyImporter struct {
	orderRepo   repositories.OrderRepository
	invoiceRepo repositories.InvoiceRepository
	config      *config.TallyConfig
	mode        string
	apiClient   interface{} // *internal.TallyAPIClient - using interface{} to avoid import cycle
}

// Setter for API client to avoid circular imports
func (i *TallyImporter) SetAPIClient(client interface{}) {
	i.apiClient = client
}

type ImportRequest struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Data     string    `json:"data"`      // CSV content
	DataType string    `json:"data_type"` // "orders" or "invoices"
}

type ImportResult struct {
	RecordsProcessed int
	RecordsImported  int
	Errors           []string
}

func NewTallyImporter(orderRepo repositories.OrderRepository, invoiceRepo repositories.InvoiceRepository, cfg *config.TallyConfig) *TallyImporter {
	return &TallyImporter{
		orderRepo:   orderRepo,
		invoiceRepo: invoiceRepo,
		config:      cfg,
		mode:        cfg.ExportImport.Mode,
	}
}

func (i *TallyImporter) isRestMode() bool {
	return i.mode == "rest"
}

func (i *TallyImporter) ImportData(ctx context.Context, req ImportRequest) (*ImportResult, error) {
	fmt.Printf("Starting import in %s mode\n", i.mode)

	if i.isRestMode() {
		return i.importViaAPI(ctx, req)
	}

	return i.importViaCSV(ctx, req)
}

func (i *TallyImporter) importViaCSV(ctx context.Context, req ImportRequest) (*ImportResult, error) {
	result := &ImportResult{
		RecordsProcessed: 0,
		RecordsImported:  0,
		Errors:           []string{},
	}

	reader := csv.NewReader(strings.NewReader(req.Data))
	records, err := reader.ReadAll()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse CSV: %v", err))
		return result, nil
	}

	if len(records) < 2 {
		result.Errors = append(result.Errors, "CSV must have at least a header row and one data row")
		return result, nil
	}

	// Skip header row
	dataRows := records[1:]

	switch req.DataType {
	case "orders":
		err = i.importOrders(ctx, req.TenantID, dataRows, result)
	case "invoices":
		err = i.importInvoices(ctx, req.TenantID, dataRows, result)
	default:
		result.Errors = append(result.Errors, "Invalid data_type: must be 'orders' or 'invoices'")
		return result, nil
	}

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Import failed: %v", err))
	}

	return result, nil
}

func (i *TallyImporter) importViaAPI(ctx context.Context, req ImportRequest) (*ImportResult, error) {
	result := &ImportResult{
		RecordsProcessed: 0,
		RecordsImported:  0,
		Errors:           []string{},
	}

	switch req.DataType {
	case "ledger":
		return i.importLedgerViaAPI(ctx, req)
	case "balances":
		return i.importBalancesViaAPI(ctx, req)
	default:
		result.Errors = append(result.Errors, "Invalid data_type for REST mode: must be 'ledger' or 'balances'")
		return result, nil
	}
}

func (i *TallyImporter) importLedgerViaAPI(ctx context.Context, req ImportRequest) (*ImportResult, error) {
	result := &ImportResult{
		RecordsProcessed: 0,
		RecordsImported:  0,
		Errors:           []string{},
	}

	if i.apiClient == nil {
		result.Errors = append(result.Errors, "API client not configured")
		return result, nil
	}

	// Import ledger data from Tally
	ledgerName := req.Data // Use data field as ledger name for REST mode
	ledgerData, err := i.callImportLedgerMethod(ctx, ledgerName)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to import ledger: %v", err))
		return result, nil
	}

	// Process ledger entries (in a real scenario, you might store this data or use it for reconciliation)
	result.RecordsProcessed = len(ledgerData)
	for _, entry := range ledgerData {
		// Log the entry for demonstration
		log.Printf("Imported ledger entry: %s, Balance: %.2f", entry.Name, entry.Balance)
		result.RecordsImported++
	}

	return result, nil
}

func (i *TallyImporter) importBalancesViaAPI(ctx context.Context, req ImportRequest) (*ImportResult, error) {
	result := &ImportResult{
		RecordsProcessed: 0,
		RecordsImported:  0,
		Errors:           []string{},
	}

	if i.apiClient == nil {
		result.Errors = append(result.Errors, "API client not configured")
		return result, nil
	}

	// Import balances from Tally
	balances, err := i.callImportBalancesMethod(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to import balances: %v", err))
		return result, nil
	}

	// Process balance entries
	result.RecordsProcessed = len(balances)
	for _, balance := range balances {
		// Log the balance for demonstration
		log.Printf("Imported balance: %s, Amount: %.2f", balance.AccountName, balance.Balance)
		result.RecordsImported++
	}

	return result, nil
}

// Helper method to call API client import ledger using interface{}
func (i *TallyImporter) callImportLedgerMethod(ctx context.Context, ledgerName string) ([]models.TallyLedger, error) {
	if client, ok := i.apiClient.(ledgerCaller); ok {
		return client.ImportLedger(ctx, ledgerName)
	}
	return nil, fmt.Errorf("API client not properly configured")
}

// Ledger caller interface to avoid reflection
type ledgerCaller interface {
	ImportLedger(ctx context.Context, ledgerName string) ([]models.TallyLedger, error)
}

// Helper method to call API client import balances using interface{}
func (i *TallyImporter) callImportBalancesMethod(ctx context.Context) ([]models.TallyBalance, error) {
	if client, ok := i.apiClient.(balanceCaller); ok {
		return client.ImportBalances(ctx)
	}
	return nil, fmt.Errorf("API client not properly configured")
}

// Balance caller interface to avoid reflection
type balanceCaller interface {
	ImportBalances(ctx context.Context) ([]models.TallyBalance, error)
}

func (i *TallyImporter) importOrders(ctx context.Context, tenantID uuid.UUID, rows [][]string, result *ImportResult) error {
	for _, row := range rows {
		result.RecordsProcessed++
		if len(row) < 7 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: insufficient columns, expected at least 7", result.RecordsProcessed))
			continue
		}

		order, err := i.parseOrderRow(tenantID, row)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", result.RecordsProcessed, err))
			continue
		}

		if err := i.orderRepo.Create(ctx, order); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: failed to save order: %v", result.RecordsProcessed, err))
			continue
		}

		result.RecordsImported++
	}

	return nil
}

func (i *TallyImporter) parseOrderRow(tenantID uuid.UUID, row []string) (*models.Order, error) {
	order := &models.Order{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   "pending",
	}

	// Expected format: Order Type, Product ID, Warehouse ID, Quantity, Unit Price, Order Date, Supplier/Distributor ID, Notes
	if len(row) >= 7 {
		order.OrderType = strings.TrimSpace(row[0])
		if order.OrderType == "" {
			return nil, fmt.Errorf("order type is required")
		}
		if order.OrderType != "purchase" && order.OrderType != "sales" {
			return nil, fmt.Errorf("order type must be 'purchase' or 'sales'")
		}

		productIDStr := strings.TrimSpace(row[1])
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %v", err)
		}
		order.ProductID = productID

		warehouseIDStr := strings.TrimSpace(row[2])
		warehouseID, err := uuid.Parse(warehouseIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid warehouse ID: %v", err)
		}
		order.WarehouseID = warehouseID

		quantityStr := strings.TrimSpace(row[3])
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			return nil, fmt.Errorf("invalid quantity: %v", err)
		}
		order.Quantity = quantity

		unitPriceStr := strings.TrimSpace(row[4])
		unitPrice, err := strconv.ParseFloat(unitPriceStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid unit price: %v", err)
		}
		order.UnitPrice = unitPrice

		orderDateStr := strings.TrimSpace(row[5])
		orderDate, err := time.Parse("2006-01-02", orderDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid order date (expected YYYY-MM-DD): %v", err)
		}
		order.OrderDate = orderDate

		// Optional fields
		if len(row) >= 7 {
			suppDistIDStr := strings.TrimSpace(row[6])
			if suppDistIDStr != "" {
				suppDistID, err := uuid.Parse(suppDistIDStr)
				if err != nil {
					return nil, fmt.Errorf("invalid supplier/distributor ID: %v", err)
				}
				if order.OrderType == "purchase" {
					order.SupplierID = &suppDistID
				} else {
					order.DistributorID = &suppDistID
				}
			}
		}

		if len(row) >= 8 {
			notes := strings.TrimSpace(row[7])
			if notes != "" {
				order.Notes = &notes
			}
		}
	}

	return order, nil
}

func (i *TallyImporter) importInvoices(ctx context.Context, tenantID uuid.UUID, rows [][]string, result *ImportResult) error {
	for _, row := range rows {
		result.RecordsProcessed++
		if len(row) < 7 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: insufficient columns, expected at least 7", result.RecordsProcessed))
			continue
		}

		invoice, err := i.parseInvoiceRow(tenantID, row)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", result.RecordsProcessed, err))
			continue
		}

		if err := i.invoiceRepo.Create(ctx, invoice); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: failed to save invoice: %v", result.RecordsProcessed, err))
			continue
		}

		result.RecordsImported++
	}

	return nil
}

func (i *TallyImporter) parseInvoiceRow(tenantID uuid.UUID, row []string) (*models.Invoice, error) {
	invoice := &models.Invoice{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   "unpaid",
	}

	// Expected format: Invoice Date, GSTIN, HSN/SAC, Taxable Amount, GST Rate, Total Amount, Order ID (optional)
	if len(row) >= 7 {
		invoiceDateStr := strings.TrimSpace(row[0])
		invoiceDate, err := time.Parse("2006-01-02", invoiceDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid invoice date (expected YYYY-MM-DD): %v", err)
		}
		invoice.IssuedDate = invoiceDate

		gstin := strings.TrimSpace(row[1])
		if gstin != "" {
			invoice.GSTIN = &gstin
		}

		hsnSac := strings.TrimSpace(row[2])
		if hsnSac != "" {
			invoice.HSNSAC = &hsnSac
		}

		taxableAmountStr := strings.TrimSpace(row[3])
		if taxableAmountStr != "" {
			taxableAmount, err := strconv.ParseFloat(taxableAmountStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid taxable amount: %v", err)
			}
			invoice.TaxableAmount = &taxableAmount
		}

		gstRateStr := strings.TrimSpace(row[4])
		if gstRateStr != "" {
			gstRate, err := strconv.ParseFloat(gstRateStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid GST rate: %v", err)
			}
			invoice.GSTRate = &gstRate

			// Calculate GST components if we have taxable amount
			if invoice.TaxableAmount != nil {
				cgst := gstRate / 2.0
				sgst := gstRate / 2.0
				cgstAmount := *invoice.TaxableAmount * cgst / 100.0
				sgstAmount := *invoice.TaxableAmount * sgst / 100.0
				invoice.CGST = &cgstAmount
				invoice.SGST = &sgstAmount
			}
		}

		totalAmountStr := strings.TrimSpace(row[5])
		totalAmount, err := strconv.ParseFloat(totalAmountStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid total amount: %v", err)
		}
		invoice.TotalAmount = totalAmount

		// Optional Order ID
		orderIDStr := strings.TrimSpace(row[6])
		if orderIDStr != "" {
			orderID, err := uuid.Parse(orderIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid order ID: %v", err)
			}
			invoice.OrderID = orderID
		}
	}

	return invoice, nil
}

// Scheduled import job (for scanning import directory, not implemented)
func (i *TallyImporter) ScheduledImportJob(ctx context.Context) error {
	fmt.Printf("Starting scheduled Tally import job in %s mode\n", i.mode)

	// In a real implementation, this would scan a directory for CSV files
	// and process them automatically

	log.Println("Scheduled import job completed (no files to process)")
	return nil
}