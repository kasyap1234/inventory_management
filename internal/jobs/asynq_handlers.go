package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Task type definitions
const (
	TypeTallyExport = "tally_export"
	TypeTallyImport = "tally_import"
)

// TallyExportPayload defines the payload for tally export tasks
type TallyExportPayload struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Format    string    `json:"format"`
	DataType  string    `json:"data_type"`
}

// TallyImportPayload defines the payload for tally import tasks
type TallyImportPayload struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Data     string    `json:"data"`
	DataType string    `json:"data_type"`
}

// NewTallyExportTask creates a new tally export task
func NewTallyExportTask(tenantID uuid.UUID, startDate, endDate, format, dataType string) (*asynq.Task, error) {
	payload := TallyExportPayload{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
		Format:    format,
		DataType:  dataType,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTallyExport, data), nil
}

// NewTallyImportTask creates a new tally import task
func NewTallyImportTask(tenantID uuid.UUID, data, dataType string) (*asynq.Task, error) {
	payload := TallyImportPayload{
		TenantID: tenantID,
		Data:     data,
		DataType: dataType,
	}
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTallyImport, dataBytes), nil
}

// TallyExportHandler handles tally export tasks
func (e *TallyExporter) TallyExportHandler(ctx context.Context, t *asynq.Task) error {
	var payload TallyExportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal export payload: %w", err)
	}

	log.Printf("Starting tally export for tenant %s, data type %s", payload.TenantID, payload.DataType)

	req := ExportRequest{
		TenantID:  payload.TenantID,
		StartDate: payload.StartDate,
		EndDate:   payload.EndDate,
		Format:    payload.Format,
	}

	var result *ExportResult
	var err error

	if payload.DataType == "invoices" || payload.DataType == "" {
		result, err = e.ExportInvoicesForTenant(ctx, req)
	} else if payload.DataType == "orders" {
		result, err = e.ExportOrdersForTenant(ctx, req)
	} else {
		return fmt.Errorf("invalid data type: %s", payload.DataType)
	}

	if err != nil {
		log.Printf("Tally export failed for tenant %s: %v", payload.TenantID, err)
		return err
	}

	log.Printf("Tally export completed for tenant %s: %d records exported", payload.TenantID, result.RecordsExported)

	// Trigger success callback (e.g., send notification or trigger another task)
	if err := e.handleExportSuccess(ctx, payload.TenantID, result); err != nil {
		log.Printf("Failed to handle export success callback: %v", err)
	}

	return nil
}

// TallyImportHandler handles tally import tasks
func (i *TallyImporter) TallyImportHandler(ctx context.Context, t *asynq.Task) error {
	var payload TallyImportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal import payload: %w", err)
	}

	log.Printf("Starting tally import for tenant %s, data type %s", payload.TenantID, payload.DataType)

	req := ImportRequest{
		TenantID: payload.TenantID,
		Data:     payload.Data,
		DataType: payload.DataType,
	}

	result, err := i.ImportData(ctx, req)
	if err != nil {
		log.Printf("Tally import failed for tenant %s: %v", payload.TenantID, err)
		return err
	}

	log.Printf("Tally import completed for tenant %s: %d/%d records processed/imported", payload.TenantID, result.RecordsProcessed, result.RecordsImported)

	// Trigger success callback
	if err := i.handleImportSuccess(ctx, payload.TenantID, result); err != nil {
		log.Printf("Failed to handle import success callback: %v", err)
	}

	return nil
}

// handleExportSuccess is a success callback after successful export
func (e *TallyExporter) handleExportSuccess(ctx context.Context, tenantID uuid.UUID, result *ExportResult) error {
	// TODO: Implement success actions like sending notification, storing result in DB, etc.
	log.Printf("Export success callback triggered for tenant %s", tenantID)
	return nil
}

// handleImportSuccess is a success callback after successful import
func (i *TallyImporter) handleImportSuccess(ctx context.Context, tenantID uuid.UUID, result *ImportResult) error {
	// TODO: Implement success actions like sending notification, storing result in DB, etc.
	log.Printf("Import success callback triggered for tenant %s", tenantID)
	return nil
}