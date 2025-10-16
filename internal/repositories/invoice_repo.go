package repositories

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GSTReportRow represents a row in GST reporting
type GSTReportRow struct {
	InvoiceID       uuid.UUID `json:"invoice_id"`
	OrderID         uuid.UUID `json:"order_id"`
	HSNSAC          *string   `json:"hsn_sac"`
	TaxableAmount   *float64  `json:"taxable_amount"`
	GSTRate         *float64  `json:"gst_rate"`
	CGST            *float64  `json:"cgst"`
	SGST            *float64  `json:"sgst"`
	IGST            *float64  `json:"igst"`
	TotalAmount     float64   `json:"total_amount"`
	Status          string    `json:"status"`
	IssuedDate      time.Time `json:"issued_date"`
	GSTIN           *string   `json:"gstin"`
}

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *models.Invoice) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Invoice, error)
	Update(ctx context.Context, invoice *models.Invoice) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invoice, error)
	GetInvoicesByTenantAndDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.Invoice, error)
	GetInvoicesByStatus(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*models.Invoice, error)
	GetInvoicesByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.Invoice, error)
	GetUnpaidInvoices(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invoice, error)
	GetGSTReportData(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]GSTReportRow, error)
	UpdateInvoiceStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status string) error
	GenerateInvoiceNumber(ctx context.Context, tenantID uuid.UUID, issuedDate time.Time) (string, error)
}

type invoiceRepo struct {
	db *pgxpool.Pool
}

func NewInvoiceRepo(db *pgxpool.Pool) InvoiceRepository {
	return &invoiceRepo{db: db}
}

func (r *invoiceRepo) Create(ctx context.Context, invoice *models.Invoice) error {
	query := `
		INSERT INTO invoices (id, tenant_id, order_id, invoice_number, gstin, hsn_sac, taxable_amount, gst_rate, cgst, sgst, igst, total_amount, status, issued_date, paid_date, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, invoice.ID, invoice.TenantID, invoice.OrderID, invoice.InvoiceNumber, invoice.GSTIN, invoice.HSNSAC, invoice.TaxableAmount, invoice.GSTRate, invoice.CGST, invoice.SGST, invoice.IGST, invoice.TotalAmount, invoice.Status, invoice.IssuedDate, invoice.PaidDate, invoice.DueDate)
	return err
}

func (r *invoiceRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Invoice, error) {
	invoice := &models.Invoice{}
	query := `
		SELECT id, tenant_id, order_id, invoice_number, gstin, hsn_sac, taxable_amount, gst_rate, cgst, sgst, igst, total_amount, status, issued_date, paid_date, due_date, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&invoice.ID, &invoice.TenantID, &invoice.OrderID, &invoice.InvoiceNumber, &invoice.GSTIN, &invoice.HSNSAC, &invoice.TaxableAmount, &invoice.GSTRate, &invoice.CGST, &invoice.SGST, &invoice.IGST, &invoice.TotalAmount, &invoice.Status, &invoice.IssuedDate, &invoice.PaidDate, &invoice.DueDate, &invoice.CreatedAt, &invoice.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return invoice, nil
}

func (r *invoiceRepo) Update(ctx context.Context, invoice *models.Invoice) error {
	query := `
		UPDATE invoices
		SET gstin = $1, hsn_sac = $2, taxable_amount = $3, gst_rate = $4, cgst = $5, sgst = $6, igst = $7, total_amount = $8, status = $9, issued_date = $10, paid_date = $11, due_date = $12, updated_at = NOW()
		WHERE tenant_id = $13 AND id = $14
	`
	_, err := r.db.Exec(ctx, query, invoice.GSTIN, invoice.HSNSAC, invoice.TaxableAmount, invoice.GSTRate, invoice.CGST, invoice.SGST, invoice.IGST, invoice.TotalAmount, invoice.Status, invoice.IssuedDate, invoice.PaidDate, invoice.DueDate, invoice.TenantID, invoice.ID)
	return err
}

func (r *invoiceRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM invoices WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *invoiceRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invoice, error) {
	query := `
		SELECT id, tenant_id, order_id, invoice_number, gstin, hsn_sac, taxable_amount, gst_rate, cgst, sgst, igst, total_amount, status, issued_date, paid_date, due_date, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1
		ORDER BY issued_date DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		if err := rows.Scan(&invoice.ID, &invoice.TenantID, &invoice.OrderID, &invoice.InvoiceNumber, &invoice.GSTIN, &invoice.HSNSAC, &invoice.TaxableAmount, &invoice.GSTRate, &invoice.CGST, &invoice.SGST, &invoice.IGST, &invoice.TotalAmount, &invoice.Status, &invoice.IssuedDate, &invoice.PaidDate, &invoice.DueDate, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

func (r *invoiceRepo) GetInvoicesByTenantAndDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.Invoice, error) {
	query := `
		SELECT id, tenant_id, order_id, invoice_number, gstin, hsn_sac, taxable_amount, gst_rate, cgst, sgst, igst, total_amount, status, issued_date, paid_date, due_date, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1 AND issued_date BETWEEN $2 AND $3
		ORDER BY issued_date DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		if err := rows.Scan(&invoice.ID, &invoice.TenantID, &invoice.OrderID, &invoice.InvoiceNumber, &invoice.GSTIN, &invoice.HSNSAC, &invoice.TaxableAmount, &invoice.GSTRate, &invoice.CGST, &invoice.SGST, &invoice.IGST, &invoice.TotalAmount, &invoice.Status, &invoice.IssuedDate, &invoice.PaidDate, &invoice.DueDate, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

// GetInvoicesByStatus retrieves invoices by status
func (r *invoiceRepo) GetInvoicesByStatus(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*models.Invoice, error) {
	query := `
		SELECT id, tenant_id, order_id, invoice_number, gstin, hsn_sac, taxable_amount, gst_rate, cgst, sgst, igst, total_amount, status, issued_date, paid_date, due_date, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1 AND status = $2
		ORDER BY issued_date DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		if err := rows.Scan(&invoice.ID, &invoice.TenantID, &invoice.OrderID, &invoice.InvoiceNumber, &invoice.GSTIN, &invoice.HSNSAC, &invoice.TaxableAmount, &invoice.GSTRate, &invoice.CGST, &invoice.SGST, &invoice.IGST, &invoice.TotalAmount, &invoice.Status, &invoice.IssuedDate, &invoice.PaidDate, &invoice.DueDate, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

// GetInvoicesByOrderID retrieves invoices for a specific order
func (r *invoiceRepo) GetInvoicesByOrderID(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.Invoice, error) {
	query := `
		SELECT id, tenant_id, order_id, invoice_number, gstin, hsn_sac, taxable_amount, gst_rate, cgst, sgst, igst, total_amount, status, issued_date, paid_date, due_date, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1 AND order_id = $2
		ORDER BY issued_date DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		if err := rows.Scan(&invoice.ID, &invoice.TenantID, &invoice.OrderID, &invoice.InvoiceNumber, &invoice.GSTIN, &invoice.HSNSAC, &invoice.TaxableAmount, &invoice.GSTRate, &invoice.CGST, &invoice.SGST, &invoice.IGST, &invoice.TotalAmount, &invoice.Status, &invoice.IssuedDate, &invoice.PaidDate, &invoice.DueDate, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

// GetUnpaidInvoices retrieves unpaid invoices
func (r *invoiceRepo) GetUnpaidInvoices(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invoice, error) {
	query := `
		SELECT id, tenant_id, order_id, invoice_number, gstin, hsn_sac, taxable_amount, gst_rate, cgst, sgst, igst, total_amount, status, issued_date, paid_date, due_date, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1 AND status NOT IN ('paid', 'cancelled')
		ORDER BY issued_date DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		if err := rows.Scan(&invoice.ID, &invoice.TenantID, &invoice.OrderID, &invoice.InvoiceNumber, &invoice.GSTIN, &invoice.HSNSAC, &invoice.TaxableAmount, &invoice.GSTRate, &invoice.CGST, &invoice.SGST, &invoice.IGST, &invoice.TotalAmount, &invoice.Status, &invoice.IssuedDate, &invoice.PaidDate, &invoice.DueDate, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

// GetGSTReportData retrieves GST report data
func (r *invoiceRepo) GetGSTReportData(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]GSTReportRow, error) {
	query := `
		SELECT id, order_id, hsn_sac, taxable_amount, gst_rate, cgst, sgst, igst, total_amount, status, issued_date, gstin
		FROM invoices
		WHERE tenant_id = $1 AND issued_date BETWEEN $2 AND $3
		ORDER BY issued_date ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reportRows []GSTReportRow
	for rows.Next() {
		row := GSTReportRow{}
		if err := rows.Scan(&row.InvoiceID, &row.OrderID, &row.HSNSAC, &row.TaxableAmount, &row.GSTRate, &row.CGST, &row.SGST, &row.IGST, &row.TotalAmount, &row.Status, &row.IssuedDate, &row.GSTIN); err != nil {
			return nil, err
		}
		reportRows = append(reportRows, row)
	}
	return reportRows, nil
}

// UpdateInvoiceStatus updates invoice status
func (r *invoiceRepo) UpdateInvoiceStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status string) error {
	query := `
		UPDATE invoices
		SET status = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND id = $3
	`
	_, err := r.db.Exec(ctx, query, status, tenantID, invoiceID)
	return err
}

// GenerateInvoiceNumber generates a unique invoice number for a tenant
func (r *invoiceRepo) GenerateInvoiceNumber(ctx context.Context, tenantID uuid.UUID, issuedDate time.Time) (string, error) {
	yearMonth := issuedDate.Format("2006-01")

	// Get the next sequence number for this tenant and month
	query := `
		WITH upsert AS (
			INSERT INTO invoice_sequences (tenant_id, year_month, last_number)
			VALUES ($1, $2, 1)
			ON CONFLICT (tenant_id, year_month)
			DO UPDATE SET
				last_number = invoice_sequences.last_number + 1,
				updated_at = NOW()
			RETURNING last_number
		)
		SELECT last_number FROM upsert;
	`

	var sequenceNum int
	err := r.db.QueryRow(ctx, query, tenantID, yearMonth).Scan(&sequenceNum)
	if err != nil {
		return "", fmt.Errorf("failed to generate invoice sequence: %w", err)
	}

	// Format invoice number: INV-TENANTSHORTID-YYYY-MM-XXXXXX
	// For simplicity, use tenant UUID suffix for brevity
	tenantSuffix := tenantID.String()[len(tenantID.String())-8:]
	invoiceNumber := fmt.Sprintf("INV-%s-%s-%06d", tenantSuffix, yearMonth, sequenceNum)

	return invoiceNumber, nil
}