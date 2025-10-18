package models

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
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

// parseTallyXML parses Tally XML response to Go data structures
// Implements proper TDL schema parsing
func (c *TallyClient) parseTallyXML(xmlData []byte) (interface{}, error) {
	var envelope tallyEnvelope
	if err := xml.Unmarshal(xmlData, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Tally XML: %w", err)
	}

	messages := make([]tallyMessage, 0)
	collectMessages := func(msgs []tallyMessage, collections []tallyCollection) {
		messages = append(messages, msgs...)
		for _, collection := range collections {
			messages = append(messages, collection.Messages...)
		}
	}

	for _, dataSection := range envelope.Body.Data {
		collectMessages(dataSection.Messages, dataSection.Collections)
	}

	for _, importSection := range envelope.Body.ImportData {
		collectMessages(importSection.RequestData.Messages, importSection.RequestData.Collections)
	}

	responseStatuses := make([]string, 0)
	for _, responseSection := range envelope.Body.ResponseData {
		if responseSection.Status != "" {
			responseStatuses = append(responseStatuses, responseSection.Status)
		}
		collectMessages(responseSection.Messages, responseSection.Collections)
	}

	vouchers := make([]map[string]interface{}, 0)
	ledgers := make([]map[string]string, 0)
	stockItems := make([]map[string]string, 0)

	for _, message := range messages {
		if message.Voucher != nil {
			vouchers = append(vouchers, map[string]interface{}{
				"voucher_type":   message.Voucher.VoucherTypeName,
				"voucher_number": message.Voucher.VoucherNumber,
				"date":           message.Voucher.Date,
				"party":          message.Voucher.PartyLedgerName,
				"effective_date": message.Voucher.EffectiveDate,
				"is_invoice":     message.Voucher.IsInvoice,
				"persisted_view": message.Voucher.PersistedView,
				"ledger_entries": convertLedgerEntries(message.Voucher.LedgerEntries),
			})
		}

		if message.Ledger != nil {
			ledgers = append(ledgers, map[string]string{
				"name":            message.Ledger.Name,
				"reserved_name":   message.Ledger.ReservedName,
				"parent":          message.Ledger.Parent,
				"opening_balance": message.Ledger.OpeningBalance,
				"is_billwise_on":  message.Ledger.IsBillwiseOn,
				"is_cost_centres": message.Ledger.IsCostCentresOn,
			})
		}

		if message.StockItem != nil {
			stockItems = append(stockItems, map[string]string{
				"name":            message.StockItem.Name,
				"reserved_name":   message.StockItem.ReservedName,
				"parent":          message.StockItem.Parent,
				"base_units":      message.StockItem.BaseUnits,
				"opening_balance": message.StockItem.OpeningBalance,
				"opening_value":   message.StockItem.OpeningValue,
				"opening_rate":    message.StockItem.OpeningRate,
			})
		}
	}

	result := map[string]interface{}{
		"format":        "tally_xml",
		"raw_data":      string(xmlData),
		"parsed_at":     time.Now().Format(time.RFC3339),
		"vouchers":      vouchers,
		"ledgers":       ledgers,
		"stock_items":   stockItems,
		"message_count": len(messages),
	}

	if len(responseStatuses) > 0 {
		result["statuses"] = responseStatuses
	}

	return result, nil
}

// convertJSONToTallyXML converts JSON data to Tally XML format
// Implements proper TDL (Tally Definition Language) schema support
func (c *TallyClient) convertJSONToTallyXML(jsonData []byte, data interface{}) ([]byte, error) {
	// Parse the JSON data to determine the type of data being exported
	var dataMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %w", err)
	}

	// Determine the data type and generate appropriate Tally XML
	if voucher, ok := dataMap["voucher"]; ok {
		return c.generateVoucherXML(voucher)
	} else if ledger, ok := dataMap["ledger"]; ok {
		return c.generateLedgerXML(ledger)
	} else if stockItem, ok := dataMap["stock_item"]; ok {
		return c.generateStockItemXML(stockItem)
	} else if order, ok := dataMap["order"]; ok {
		return c.generateSalesOrderXML(order)
	}

	// Fallback to generic import format
	return c.generateGenericImportXML(jsonData)
}

func convertLedgerEntries(entries []tallyLedgerEntry) []map[string]string {
	result := make([]map[string]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, map[string]string{
			"ledger_name":        entry.LedgerName,
			"is_deemed_positive": entry.IsDeemedPositive,
			"amount":             entry.Amount,
		})
	}
	return result
}

type tallyEnvelope struct {
	Body tallyBody `xml:"BODY"`
}

type tallyBody struct {
	Data         []tallyDataSection     `xml:"DATA"`
	ImportData   []tallyImportSection   `xml:"IMPORTDATA"`
	ResponseData []tallyResponseSection `xml:"RESPONSEDATA"`
}

type tallyDataSection struct {
	Messages    []tallyMessage    `xml:"TALLYMESSAGE"`
	Collections []tallyCollection `xml:"COLLECTION"`
}

type tallyCollection struct {
	Messages []tallyMessage `xml:"TALLYMESSAGE"`
}

type tallyImportSection struct {
	RequestData tallyRequestData `xml:"REQUESTDATA"`
}

type tallyRequestData struct {
	Messages    []tallyMessage    `xml:"TALLYMESSAGE"`
	Collections []tallyCollection `xml:"COLLECTION"`
}

type tallyResponseSection struct {
	Status      string            `xml:"STATUS"`
	Messages    []tallyMessage    `xml:"TALLYMESSAGE"`
	Collections []tallyCollection `xml:"COLLECTION"`
}

type tallyMessage struct {
	Voucher   *tallyVoucherMessage   `xml:"VOUCHER"`
	Ledger    *tallyLedgerMessage    `xml:"LEDGER"`
	StockItem *tallyStockItemMessage `xml:"STOCKITEM"`
}

type tallyVoucherMessage struct {
	VoucherTypeName string             `xml:"VOUCHERTYPENAME"`
	VoucherNumber   string             `xml:"VOUCHERNUMBER"`
	Date            string             `xml:"DATE"`
	PartyLedgerName string             `xml:"PARTYLEDGERNAME"`
	EffectiveDate   string             `xml:"EFFECTIVEDATE"`
	IsInvoice       string             `xml:"ISINVOICE"`
	PersistedView   string             `xml:"PERSISTEDVIEW"`
	LedgerEntries   []tallyLedgerEntry `xml:"ALLLEDGERENTRIES.LIST"`
}

type tallyLedgerEntry struct {
	LedgerName       string `xml:"LEDGERNAME"`
	IsDeemedPositive string `xml:"ISDEEMEDPOSITIVE"`
	Amount           string `xml:"AMOUNT"`
}

type tallyLedgerMessage struct {
	Name            string `xml:"NAME,attr"`
	ReservedName    string `xml:"RESERVEDNAME,attr"`
	Parent          string `xml:"PARENT"`
	OpeningBalance  string `xml:"OPENINGBALANCE"`
	IsBillwiseOn    string `xml:"ISBILLWISEON"`
	IsCostCentresOn string `xml:"ISCOSTCENTRESON"`
}

type tallyStockItemMessage struct {
	Name           string `xml:"NAME,attr"`
	ReservedName   string `xml:"RESERVEDNAME,attr"`
	Parent         string `xml:"PARENT"`
	BaseUnits      string `xml:"BASEUNITS"`
	OpeningBalance string `xml:"OPENINGBALANCE"`
	OpeningValue   string `xml:"OPENINGVALUE"`
	OpeningRate    string `xml:"OPENINGRATE"`
}

// generateVoucherXML generates Tally XML for voucher entries (sales, purchase, payment, receipt)
func (c *TallyClient) generateVoucherXML(voucher interface{}) ([]byte, error) {
	voucherMap, ok := voucher.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid voucher data format")
	}

	voucherType := getStringValue(voucherMap, "type", "Sales")
	voucherNumber := getStringValue(voucherMap, "number", "")
	date := getStringValue(voucherMap, "date", time.Now().Format("20060102"))
	partyName := getStringValue(voucherMap, "party_name", "")
	amount := getStringValue(voucherMap, "amount", "0")

	xmlTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<ENVELOPE>
  <HEADER>
    <TALLYREQUEST>Import Data</TALLYREQUEST>
  </HEADER>
  <BODY>
    <IMPORTDATA>
      <REQUESTDESC>
        <REPORTNAME>Vouchers</REPORTNAME>
        <STATICVARIABLES>
          <SVCURRENTCOMPANY>%s</SVCURRENTCOMPANY>
        </STATICVARIABLES>
      </REQUESTDESC>
      <REQUESTDATA>
        <TALLYMESSAGE xmlns:UDF="TallyUDF">
          <VOUCHER REMOTEID="" VCHKEY="" VCHTYPE="%s" ACTION="Create" OBJVIEW="Accounting Voucher View">
            <DATE>%s</DATE>
            <VOUCHERTYPENAME>%s</VOUCHERTYPENAME>
            <VOUCHERNUMBER>%s</VOUCHERNUMBER>
            <PARTYLEDGERNAME>%s</PARTYLEDGERNAME>
            <EFFECTIVEDATE>%s</EFFECTIVEDATE>
            <ISINVOICE>Yes</ISINVOICE>
            <PERSISTEDVIEW>Accounting Voucher View</PERSISTEDVIEW>
            <ALLLEDGERENTRIES.LIST>
              <LEDGERNAME>%s</LEDGERNAME>
              <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
              <AMOUNT>%s</AMOUNT>
            </ALLLEDGERENTRIES.LIST>
          </VOUCHER>
        </TALLYMESSAGE>
      </REQUESTDATA>
    </IMPORTDATA>
  </BODY>
</ENVELOPE>`

	xmlData := fmt.Sprintf(xmlTemplate, c.DatabaseURL, voucherType, date, voucherType, voucherNumber, partyName, date, partyName, amount)
	return []byte(xmlData), nil
}

// generateLedgerXML generates Tally XML for ledger master creation
func (c *TallyClient) generateLedgerXML(ledger interface{}) ([]byte, error) {
	ledgerMap, ok := ledger.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid ledger data format")
	}

	ledgerName := getStringValue(ledgerMap, "name", "")
	parent := getStringValue(ledgerMap, "parent", "Sundry Debtors")
	openingBalance := getStringValue(ledgerMap, "opening_balance", "0")

	xmlTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<ENVELOPE>
  <HEADER>
    <TALLYREQUEST>Import Data</TALLYREQUEST>
  </HEADER>
  <BODY>
    <IMPORTDATA>
      <REQUESTDESC>
        <REPORTNAME>All Masters</REPORTNAME>
        <STATICVARIABLES>
          <SVCURRENTCOMPANY>%s</SVCURRENTCOMPANY>
        </STATICVARIABLES>
      </REQUESTDESC>
      <REQUESTDATA>
        <TALLYMESSAGE xmlns:UDF="TallyUDF">
          <LEDGER NAME="%s" RESERVEDNAME="">
            <PARENT>%s</PARENT>
            <OPENINGBALANCE>%s</OPENINGBALANCE>
            <ISBILLWISEON>Yes</ISBILLWISEON>
            <ISCOSTCENTRESON>No</ISCOSTCENTRESON>
          </LEDGER>
        </TALLYMESSAGE>
      </REQUESTDATA>
    </IMPORTDATA>
  </BODY>
</ENVELOPE>`

	xmlData := fmt.Sprintf(xmlTemplate, c.DatabaseURL, ledgerName, parent, openingBalance)
	return []byte(xmlData), nil
}

// generateStockItemXML generates Tally XML for stock item master creation
func (c *TallyClient) generateStockItemXML(stockItem interface{}) ([]byte, error) {
	stockMap, ok := stockItem.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid stock item data format")
	}

	itemName := getStringValue(stockMap, "name", "")
	category := getStringValue(stockMap, "category", "Primary")
	unit := getStringValue(stockMap, "unit", "Nos")
	openingQty := getStringValue(stockMap, "opening_quantity", "0")
	openingValue := getStringValue(stockMap, "opening_value", "0")

	xmlTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<ENVELOPE>
  <HEADER>
    <TALLYREQUEST>Import Data</TALLYREQUEST>
  </HEADER>
  <BODY>
    <IMPORTDATA>
      <REQUESTDESC>
        <REPORTNAME>All Masters</REPORTNAME>
        <STATICVARIABLES>
          <SVCURRENTCOMPANY>%s</SVCURRENTCOMPANY>
        </STATICVARIABLES>
      </REQUESTDESC>
      <REQUESTDATA>
        <TALLYMESSAGE xmlns:UDF="TallyUDF">
          <STOCKITEM NAME="%s" RESERVEDNAME="">
            <PARENT>%s</PARENT>
            <BASEUNITS>%s</BASEUNITS>
            <OPENINGBALANCE>%s</OPENINGBALANCE>
            <OPENINGVALUE>%s</OPENINGVALUE>
            <OPENINGRATE>0</OPENINGRATE>
          </STOCKITEM>
        </TALLYMESSAGE>
      </REQUESTDATA>
    </IMPORTDATA>
  </BODY>
</ENVELOPE>`

	xmlData := fmt.Sprintf(xmlTemplate, c.DatabaseURL, itemName, category, unit, openingQty, openingValue)
	return []byte(xmlData), nil
}

// generateSalesOrderXML generates Tally XML for sales order
func (c *TallyClient) generateSalesOrderXML(order interface{}) ([]byte, error) {
	orderMap, ok := order.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid order data format")
	}

	orderNumber := getStringValue(orderMap, "order_number", "")
	date := getStringValue(orderMap, "date", time.Now().Format("20060102"))
	partyName := getStringValue(orderMap, "party_name", "")
	amount := getStringValue(orderMap, "total_amount", "0")

	xmlTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<ENVELOPE>
  <HEADER>
    <TALLYREQUEST>Import Data</TALLYREQUEST>
  </HEADER>
  <BODY>
    <IMPORTDATA>
      <REQUESTDESC>
        <REPORTNAME>Vouchers</REPORTNAME>
        <STATICVARIABLES>
          <SVCURRENTCOMPANY>%s</SVCURRENTCOMPANY>
        </STATICVARIABLES>
      </REQUESTDESC>
      <REQUESTDATA>
        <TALLYMESSAGE xmlns:UDF="TallyUDF">
          <VOUCHER REMOTEID="" VCHTYPE="Sales Order" ACTION="Create">
            <DATE>%s</DATE>
            <VOUCHERTYPENAME>Sales Order</VOUCHERTYPENAME>
            <VOUCHERNUMBER>%s</VOUCHERNUMBER>
            <PARTYLEDGERNAME>%s</PARTYLEDGERNAME>
            <EFFECTIVEDATE>%s</EFFECTIVEDATE>
            <PERSISTEDVIEW>Order Voucher View</PERSISTEDVIEW>
            <ALLLEDGERENTRIES.LIST>
              <LEDGERNAME>%s</LEDGERNAME>
              <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
              <AMOUNT>%s</AMOUNT>
            </ALLLEDGERENTRIES.LIST>
          </VOUCHER>
        </TALLYMESSAGE>
      </REQUESTDATA>
    </IMPORTDATA>
  </BODY>
</ENVELOPE>`

	xmlData := fmt.Sprintf(xmlTemplate, c.DatabaseURL, date, orderNumber, partyName, date, partyName, amount)
	return []byte(xmlData), nil
}

// generateGenericImportXML generates generic Tally XML for any data
func (c *TallyClient) generateGenericImportXML(jsonData []byte) ([]byte, error) {
	xmlTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<ENVELOPE>
  <HEADER>
    <TALLYREQUEST>Import Data</TALLYREQUEST>
  </HEADER>
  <BODY>
    <IMPORTDATA>
      <REQUESTDESC>
        <REPORTNAME>All Masters</REPORTNAME>
        <STATICVARIABLES>
          <SVCURRENTCOMPANY>%s</SVCURRENTCOMPANY>
        </STATICVARIABLES>
      </REQUESTDESC>
      <REQUESTDATA>
        <![CDATA[%s]]>
      </REQUESTDATA>
    </IMPORTDATA>
  </BODY>
</ENVELOPE>`

	xmlData := fmt.Sprintf(xmlTemplate, c.DatabaseURL, string(jsonData))
	return []byte(xmlData), nil
}

// getStringValue is a helper function to safely extract string values from maps
func getStringValue(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
		return fmt.Sprintf("%v", val)
	}
	return defaultValue
}
