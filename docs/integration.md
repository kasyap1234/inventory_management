# Tally Integration Documentation

This document provides comprehensive information about the Tally export/import functionality integration for GST compliance and accounting synchronization.

## Overview

The Tally integration enables seamless data exchange between Agromart and Tally accounting software through CSV-based export/import operations. This integration supports:

- **GST Compliance**: Automatic generation of GST-compliant CSV exports for invoices
- **Order Data Export**: Exporting order information for accounting reconciliation
- **Data Import**: Importing corrected or additional data from CSV files
- **Background Processing**: Asynchronous processing using Asynq job queues
- **Multi-tenant Support**: Data isolation per tenant with proper authorization

### Current Implementation

The current implementation uses CSV format for data exchange, providing compatibility with Tally's import capabilities while maintaining data integrity and GST compliance requirements.

### Future REST API Integration

The architecture is designed to support future direct REST API integration with Tally, allowing for real-time synchronization of data without manual CSV imports/exports.

## API Endpoints

### Export Endpoint

**Endpoint**: `POST /api/tally/export`

Triggers an asynchronous export job that generates CSV files containing invoice or order data for the specified date range.

#### Request Parameters

```json
{
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "format": "csv",
  "data_type": "invoices"
}
```

#### Parameters

- `start_date`: Start date for export (YYYY-MM-DD format, defaults to 30 days ago)
- `end_date`: End date for export (YYYY-MM-DD format, defaults to today)
- `format`: Export format (currently only "csv" supported)
- `data_type`: Type of data to export ("invoices" or "orders", defaults to "invoices")

#### Successful Response (HTTP 202)

```json
{
  "message": "Export job queued successfully",
  "job_id": "asynq-job-id-here",
  "type": "tally_export"
}
```

### Import Endpoint

**Endpoint**: `POST /api/tally/import`

Processes CSV data and imports orders or invoices into Agromart synchronously.

#### Request Parameters

```json
{
  "data": "Order Type,Product ID,Warehouse ID,Quantity,Unit Price,Order Date,Supplier/Distributor ID\npurchase,550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440001,100,50.00,2024-01-01,550e8400-e29b-41d4-a716-446655440002",
  "data_type": "orders"
}
```

#### Parameters

- `data`: CSV content as string (must include header row)
- `data_type`: Type of data ("orders" or "invoices")

#### Success Response (HTTP 200)

```json
{
  "records_processed": 10,
  "records_imported": 8,
  "errors": ["Row 9: invalid product ID", "Row 10: insufficient columns"],
  "message": "Import completed with some errors"
}
```

#### Authentication

Both endpoints require JWT authentication and extract tenant ID from the request context. Multi-tenant data isolation is enforced.

## Configuration

The Tally integration requires configuration in the `config/tally.toml` file.

### Configuration File Structure

```toml
# config/tally.toml

# Tally Integration Configuration
[tally]
api_key = "your_tally_api_key_here"
api_secret = "your_tally_api_secret_here"
api_endpoint = "https://api.tallysolutions.com"
database_url = "postgresql://tallyuser:tallypass@localhost/tallydb?sslmode=disable"

# Background Queuing Configuration
[queuing]
redis_addr = "localhost:6379"
redis_password = ""
redis_db = 0
concurrency = 10
queues = ["critical", "default", "low"]
queue_priorities = { critical = 6, default = 3, low = 1 }

# Export/Import Settings
[export_import]
timeout_seconds = 300
max_retry_attempts = 3
retry_delay_seconds = 5
```

### Configuration Sections

#### Tally Section
- `api_key`: API key for Tally Solutions API (for future use)
- `api_secret`: API secret for Tally Solutions API (for future use)
- `api_endpoint`: Base URL for Tally API (for future use)
- `database_url`: Direct database connection for data export (optional)

#### Queuing Section
- `redis_addr`: Redis server address for Asynq job queue
- `redis_password`: Redis authentication password
- `redis_db`: Redis database number
- `concurrency`: Number of concurrent workers for job processing
- `queues`: List of queue names for job prioritization
- `queue_priorities`: Priority values for each queue

#### Export/Import Settings
- `timeout_seconds`: Maximum execution time for export/import operations
- `max_retry_attempts`: Number of retry attempts for failed operations
- `retry_delay_seconds`: Delay between retry attempts

### Environment Setup

1. Create or update `config/tally.toml` with appropriate values
2. Ensure Redis server is running and accessible
3. Configure database connectivity if using direct Tally database access
4. Set appropriate file permissions for the configuration file

## Background Job Processing

The integration uses Asynq (https://github.com/hibiken/asynq) for background job processing, ensuring non-blocking API responses and fault tolerance.

### Job Types

- `tally_export`: Asynchronous export of data to CSV format
- `tally_import`: Future support for asynchronous import operations

### Job Processing Flow

1. **API Request**: User triggers export/import operation
2. **Job Queuing**: Request is queued as an Asynq job
3. **ASAP Response**: API returns job ID immediately
4. **Background Processing**: Worker processes job asynchronously
5. **Result Handling**: Success/failure callbacks handle outcomes

### Queue Configuration

Jobs are processed based on priority:
- **Critical** (priority 6): Immediate processing requests
- **Default** (priority 3): Regular exports/imports
- **Low** (priority 1): Scheduled/batch operations

### Monitoring Jobs

Use Asynq web UI or CLI tools to monitor job status:

```bash
# Start Asynq CLI monitoring
asynqmon

# View job statistics
asynqmon --queues=default,critical

# Check failed jobs
asynqmon --failed
```

## Usage Instructions

### Exporting Data

#### 1. Basic Invoice Export

```bash
curl -X POST \
  https://api.agromart.com/api/tally/export \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "data_type": "invoices"
  }'
```

#### 2. Order Data Export

```bash
curl -X POST \
  https://api.agromart.com/api/tally/export \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "data_type": "orders"
  }'
```

#### 3. Monitor Export Progress

Use the returned `job_id` to track job progress through your job monitoring interface.

### Importing Data

#### 1. Prepare CSV Data

Create CSV file with appropriate headers:

For Orders:
```csv
Order Type,Product ID,Warehouse ID,Quantity,Unit Price,Order Date,Supplier/Distributor ID,Notes
purchase,550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440001,100,50.00,2024-01-01,550e8400-e29b-41d4-a716-446655440002,
sales,b2c3d4e5-f6g7-8901-bcdef-123456789012,x2y3z4a5-b6c7-8901-defgh-234567890123,50,30.00,2024-01-05,d2e3f4g5-h6i7-8901-jklmn-345678901234,Urgent delivery
```

For Invoices:
```csv
Invoice Date,GSTIN,HSN/SAC,Taxable Amount,GST Rate,Total Amount,Order ID
2024-01-15,22AAAAA0000A1Z5,1234,1000.00,18.0,1180.00,550e8400-e29b-41d4-a716-446655440003
```

#### 2. Convert to API Format

Convert CSV content to base64 or process as string for transmission.

#### 3. Import Request

```bash
curl -X POST \
  https://api.agromart.com/api/tally/import \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data": "Order Type,Product ID,Warehouse ID,...\npurchase,550e8400-...,550e8400-...,100,50.00,2024-01-01,550e8400-...",
    "data_type": "orders"
  }'
```

### Automated Processing

For automated exports, integrate with your scheduler:

```bash
# Cron job for monthly exports
0 2 1 * * /usr/local/bin/export-tally-data.sh
```

## Troubleshooting

### Common Issues

#### 1. Job Queue Not Processing

**Symptoms**: Export jobs remain queued indefinitely

**Possible Causes**:
- Redis server not running
- Worker processes not started
- Queue configuration mismatch

**Solutions**:
```bash
# Check Redis connectivity
redis-cli ping

# Verify Asynq workers are running
ps aux | grep asynq

# Check Redis memory usage
redis-cli info memory
```

#### 2. Invalid CSV Format

**Symptoms**: Import fails with "CSV parse error"

**Common Issues**:
- Missing header row
- Incorrect column order
- Invalid UUID format
- Wrong date format (expected YYYY-MM-DD)

**Validation Checklist**:
- [ ] CSV contains header row as first line
- [ ] All required columns present
- [ ] UUIDs are valid 36-character strings
- [ ] Dates are in YYYY-MM-DD format
- [ ] No empty values in required fields
- [ ] Proper line endings (LF/Unix format)

#### 3. Configuration Errors

**Symptoms**: "Failed to load config file" error

**Solutions**:
```bash
# Check file permissions
ls -la config/tally.toml

# Validate TOML syntax
toml-check config/tally.toml

# Verify configuration values
cat config/tally.toml
```

#### 4. Database Connection Issues

**Symptoms**: Export/import operations fail with database errors

**Troubleshooting Steps**:
1. Verify database connectivity
2. Check connection pool settings
3. Validate SSL/TLS configuration
4. Review database permissions

#### 5. Resource Exhaustion

**Symptoms**: Operations fail with timeout errors

**Possible Causes**:
- Large dataset causing memory issues
- Database query timeouts
- Network latency

**Mitigation**:
- Implement data pagination for large exports
- Increase timeout values in configuration
- Optimize database queries
- Monitor resource usage

### Logging and Debugging

Enable detailed logging for troubleshooting:

```go
// Set log level to debug
log.SetLevel(log.DebugLevel)

// Check application logs
tail -f /var/log/agromart/application.log

// Monitor Asynq job logs
asynqmon --verbose
```

### Performance Tuning

For high-volume processing:

1. **Scale worker concurrency**: Increase `concurrency` in configuration
2. **Implement data partitioning**: Process data in batches
3. **Optimize database queries**: Add appropriate indexes
4. **Use connection pooling**: Configure proper pool sizes
5. **Monitor resource usage**: Track memory and CPU utilization

## Data Flow

### Export Flow

1. **Request Submission**:
   - User submits export request via API
   - Request validated and sanitized
   - Job queued in Redis (Asynq)

2. **Background Processing**:
   - Worker picks up job from queue
   - Query database for relevant data
   - Generate CSV with proper formatting
   - Save results to storage/notification queue

3. **Result Notification**:
   - User notified via email/webhook
   - CSV file available for download
   - Audit log updated

### Import Flow

1. **Data Validation**:
   - CSV data parsed and validated
   - Business rules applied
   - Referential integrity checked

2. **Processing**:
   - Records processed in transaction batches
   - Validation errors collected
   - Successful records committed

3. **Result Reporting**:
   - Import summary returned
   - Partial failures handled gracefully
   - Audit trail maintained

### Data Mapping

#### Invoice Export Format

Field mapping from Agromart to Tally CSV:

| Agromart Field | Tally CSV Field | Description |
|----------------|-----------------|-------------|
| ID | Invoice No | Formatted as "INV-{first-8-chars}" |
| IssuedDate | Invoice Date | DD/MM/YYYY format |
| GSTIN | GSTIN/UIN | GSTIN from invoice |
| - | Party Name | Currently "Customer" |
| HSN/SAC | HSN/SAC | HSN or SAC code |
| TaxableAmount | Taxable Value | Amount before GST |
| - | CESS Rate | Currently "0" |
| CGST | CGST Amount | CGST component |
| SGST | SGST Amount | SGST component |
| IGST | IGST Amount | IGST component |
| CGST+SGST+IGST | Total GST Amount | Total GST amount |
| TotalAmount | Total Amount | Final invoice amount |
| - | Place of Supply | Currently "Maharashtra" |

#### Order Export Format

| Order Field | CSV Field | Description |
|-------------|------------|-------------|
| ID | Order No | Formatted as "ORD-{first-8-chars}" |
| OrderDate | Order Date | DD/MM/YYYY format |
| OrderType | Order Type | "purchase" or "sales" |
| ProductID | Product ID | Full UUID |
| Quantity | Quantity | Integer quantity |
| UnitPrice | Unit Price | Price per unit |
| Total Amount | Total Amount | Quantity × Unit Price |
| Status | Status | Current order status |
| Supplier/Distributor ID | Supplier/Distributor ID | Associated entity UUID |
| WarehouseID | Warehouse ID | Associated warehouse UUID |

## Testing Procedures

### Unit Testing

#### Test CSV Parsing

```go
func TestTallyImporter_TestCSV(t *testing.T) {
    csvData := `Order Type,Product ID,Warehouse ID,Quantity,Unit Price,Order Date,Supplier/Distributor ID
purchase,a1b2c3d4-e5f6-7890-abcd-ef1234567890,w1x2y3z4-a5b6-7890-cdef-f1234567890,100,25.50,2023-12-01,s1t2u3v4-w5x6-7890-yzab-cd1234567890`

    reader := csv.NewReader(strings.NewReader(csvData))
    records, err := reader.ReadAll()
    assert.NoError(t, err)
    assert.Equal(t, 3, len(records)) // header + 2 data rows
}
```

#### Test GST CSV Generation

```go
func TestGSTCSVFormat(t *testing.T) {
    csvHeader := "Invoice No,Invoice Date,GSTIN/UIN,Party Name,HSN/SAC,Taxable Value,CESS Rate,CGST Amount,SGST Amount,IGST Amount,Total GST Amount,Total Amount,Place of Supply\n"
    // Expected 13 columns in GST format
    assert.Equal(t, 13, len(strings.Split(csvHeader, ",")))
}
```

### Integration Testing

#### 1. End-to-End Export Test

```bash
# Test invoice export
curl -X POST \
  https://staging.api.agromart.com/api/tally/export \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -d '{"start_date":"2024-01-01","end_date":"2024-01-31","data_type":"invoices"}' \
  -w "\nStatus: %{http_code}\n"

# Verify job creation and queue processing
```

#### 2. Import Test with Sample Data

```bash
# Create test CSV
cat > test_import.csv << EOF
Order Type,Product ID,Warehouse ID,Quantity,Unit Price,Order Date,Supplier/Distributor ID
purchase,550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440001,100,50.00,2024-01-01,550e8400-e29b-41d4-a716-446655440002
EOF

# Import test data
csv_content=$(cat test_import.csv | base64 -w 0)
curl -X POST \
  https://staging.api.agromart.com/api/tally/import \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -d "{\"data\":\"$csv_content\",\"data_type\":\"orders\"}"
```

#### 3. Performance Testing

```bash
# Test with large dataset
time curl -X POST \
  https://staging.api.agromart.com/api/tally/export \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -d '{"start_date":"2023-01-01","end_date":"2024-01-01","data_type":"orders"}'
```

### Configuration Testing

#### Validate Configuration Loading

```bash
# Test configuration syntax
go run -c "import (\"github.com/BurntSushi/toml\"); config := &TallyConfig{};  toml.DecodeFile(\"config/tally.toml\", config)"

# Test Redis connection
redis-cli -h localhost -p 6379 ping

# Test database connection
psql $DATABASE_URL -c "SELECT 1;"
```

### Continuous Integration

Include these tests in CI/CD pipeline:

```yaml
# Example CI test stage
test_integration:
  script:
    - go test ./internal/jobs/... -v
    - go test ./internal/handlers/... -v
    - go test ./tests/integration/... -v
  services:
    - redis:alpine
    - postgres:alpine
  variables:
    REDIS_URL: "redis://redis:6379"
    DATABASE_URL: "postgres://postgres:password@postgres:5432/agromart_test"
```

## Future Enhancements

### REST API Integration

The architecture supports direct Tally REST API integration:

- **Real-time Sync**: Automatic synchronization instead of batch processing
- **Webhook Support**: Event-driven updates
- **API Token Management**: Secure credential handling
- **Error Handling**: Comprehensive error recovery
- **Rate Limiting**: Respect API quotas and limits

### Advanced Features

- **Incremental Exports**: Export only changes since last sync
- **Data Validation**: Enhanced semantic validation
- **Custom Mappings**: Configurable field mappings
- **Multi-format Support**: Excel, XML, JSON exports
- **Audit Trails**: Complete change history
- **Compliance Reports**: Automated GST compliance reporting

### Monitoring and Observability

- **Metrics Collection**: Prometheus-compatible metrics
- **Distributed Tracing**: Request tracing across services
- **Alerting**: Configurable alerts for failures
- **Dashboard Integration**: Visualization for sync status

---

For additional support or questions about the Tally integration, refer to the development team or check the codebase for implementation details.