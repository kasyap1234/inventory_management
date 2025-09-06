package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"agromart2/internal/config"
	"agromart2/internal/models"
)

// TallyAPIClient handles direct REST API communication with Tally
type TallyAPIClient struct {
	config     *config.TallyConfig
	httpClient *http.Client
}

// NewTallyAPIClient creates a new Tally REST API client
func NewTallyAPIClient(cfg *config.TallyConfig) *TallyAPIClient {
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.ExportImport.TimeoutSeconds) * time.Second,
	}

	return &TallyAPIClient{
		config:     cfg,
		httpClient: httpClient,
	}
}

// makeRequest performs an HTTP request to the Tally API
func (c *TallyAPIClient) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.config.Tally.APIEndpoint, endpoint)

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Tally.APIKey))

	fmt.Printf("Making API request: %s %s\n", method, url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// ExportInvoice exports an invoice to Tally via REST API
func (c *TallyAPIClient) ExportInvoice(ctx context.Context, invoice *models.Invoice, totalAmount float64) error {
	endpoint := "/api/invoices"

	payload := map[string]interface{}{
		"document_type": "Invoice",
		"party":         "Customer", // Using static value since invoice doesn't have customer name
		"date":          invoice.IssuedDate.Format("2006-01-02"),
		"total_amount":  invoice.TotalAmount,
		"invoice_number": invoice.InvoiceNumber,
	}

	resp, err := c.makeRequest(ctx, "POST", endpoint, payload)
	if err != nil {
		fmt.Printf("Failed to export invoice: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Tally API error: status %d, body: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("Tally API returned status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Successfully exported invoice: %s\n", invoice.ID)
	return nil
}

// ExportOrder exports an order to Tally via REST API
func (c *TallyAPIClient) ExportOrder(ctx context.Context, order *models.Order) error {
	endpoint := "/api/orders"

	payload := map[string]interface{}{
		"document_type":     "Order",
		"party":            "Party", // Using static value since order doesn't have customer name
		"date":             order.OrderDate.Format("2006-01-02"),
		"total_amount":     float64(order.Quantity) * order.UnitPrice,
		"quantity":         order.Quantity,
		"unit_price":       order.UnitPrice,
		"order_type":       order.OrderType,
	}

	resp, err := c.makeRequest(ctx, "POST", endpoint, payload)
	if err != nil {
		fmt.Printf("Failed to export order: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Tally API error: status %d, body: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("Tally API returned status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Successfully exported order: %s\n", order.ID)
	return nil
}

// ImportLedger imports ledger information from Tally
func (c *TallyAPIClient) ImportLedger(ctx context.Context, ledgerName string) ([]models.TallyLedger, error) {
	endpoint := fmt.Sprintf("/api/ledger/%s", ledgerName)

	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		fmt.Printf("Failed to import ledger: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Tally API error: status %d, body: %s\n", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Tally API returned status %d: %s", resp.StatusCode, string(body))
	}

	var ledgerResponse struct {
		Ledger []models.TallyLedger `json:"ledger"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &ledgerResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Printf("Successfully imported ledger: %s, entries: %d\n", ledgerName, len(ledgerResponse.Ledger))
	return ledgerResponse.Ledger, nil
}

// ImportBalances imports account balances from Tally
func (c *TallyAPIClient) ImportBalances(ctx context.Context) ([]models.TallyBalance, error) {
	endpoint := "/api/balances"

	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		fmt.Printf("Failed to import balances: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Tally API error: status %d, body: %s\n", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Tally API returned status %d: %s", resp.StatusCode, string(body))
	}

	var balanceResponse struct {
		Balances []models.TallyBalance `json:"balances"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &balanceResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Printf("Successfully imported balances, entries: %d\n", len(balanceResponse.Balances))
	return balanceResponse.Balances, nil
}