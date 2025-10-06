package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TallyClient handles API interactions with Tally ERP
type TallyClient struct {
	APIKey      string
	APISecret   string
	APIEndpoint string
	DatabaseURL string
	HTTPClient  *http.Client
}

// NewTallyClient creates a new TallyClient instance
func NewTallyClient(apiKey, apiSecret, apiEndpoint, databaseURL string) *TallyClient {
	return &TallyClient{
		APIKey:      apiKey,
		APISecret:   apiSecret,
		APIEndpoint: apiEndpoint,
		DatabaseURL: databaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExportData exports data to Tally ERP via XML or JSON format
func (c *TallyClient) ExportData(data interface{}) error {
	// Check if API endpoint is configured
	if c.APIEndpoint == "" {
		return fmt.Errorf("Tally API endpoint not configured")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Tally ERP supports XML format primarily, but we'll use JSON for the API
	// Convert to Tally XML format if needed
	xmlData, err := c.convertJSONToTallyXML(jsonData, data)
	if err != nil {
		return fmt.Errorf("failed to convert to Tally XML: %w", err)
	}

	// Send data to Tally ERP
	req, err := http.NewRequest("POST", c.APIEndpoint+"/import", bytes.NewBuffer(xmlData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.APIKey != "" && c.APISecret != "" {
		req.SetBasicAuth(c.APIKey, c.APISecret)
	}
	req.Header.Set("Content-Type", "application/xml")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send export request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("export failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ImportData imports data from Tally ERP in XML or JSON format
func (c *TallyClient) ImportData() (interface{}, error) {
	// Check if API endpoint is configured
	if c.APIEndpoint == "" {
		return nil, fmt.Errorf("Tally API endpoint not configured")
	}

	// Request data from Tally ERP
	req, err := http.NewRequest("GET", c.APIEndpoint+"/export", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.APIKey != "" && c.APISecret != "" {
		req.SetBasicAuth(c.APIKey, c.APISecret)
	}
	req.Header.Set("Accept", "application/xml,application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send import request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("import failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try to detect format and parse accordingly
	contentType := resp.Header.Get("Content-Type")
	var data interface{}

	if bytes.Contains([]byte(contentType), []byte("xml")) {
		// Parse XML response from Tally
		data, err = c.parseTallyXML(body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Tally XML: %w", err)
		}
	} else {
		// Try JSON parsing
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return data, nil
}

// convertJSONToTallyXML converts JSON data to Tally XML format
func (c *TallyClient) convertJSONToTallyXML(jsonData []byte, data interface{}) ([]byte, error) {
	// Tally uses a specific XML format
	// This is a simplified implementation - in production, use proper XML marshaling
	// based on Tally's TDL (Tally Definition Language) schema
	
	// For now, wrap JSON in a basic XML envelope
	// In production, implement proper Tally XML voucher/ledger format
	xmlTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<ENVELOPE>
  <HEADER>
    <TALLYREQUEST>Import Data</TALLYREQUEST>
  </HEADER>
  <BODY>
    <IMPORTDATA>
      <REQUESTDESC>
        <REPORTNAME>All Masters</REPORTNAME>
      </REQUESTDESC>
      <REQUESTDATA>%s</REQUESTDATA>
    </IMPORTDATA>
  </BODY>
</ENVELOPE>`

	xmlData := []byte(fmt.Sprintf(xmlTemplate, string(jsonData)))
	return xmlData, nil
}

// parseTallyXML parses Tally XML response to Go data structures
func (c *TallyClient) parseTallyXML(xmlData []byte) (interface{}, error) {
	// This is a simplified implementation
	// In production, implement proper Tally XML parsing based on TDL schema
	
	// For now, return a map with the raw XML
	result := map[string]interface{}{
		"format": "tally_xml",
		"data":   string(xmlData),
	}
	
	return result, nil
}