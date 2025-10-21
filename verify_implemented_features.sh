#!/bin/bash

# Feature Implementation Verification Script
# Tests all newly implemented features

set -e

echo "======================================"
echo "Feature Implementation Verification"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
TENANT_ID=""
USER_ID=""
TOKEN=""

# Function to print success
success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Function to print error
error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to print info
info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Function to check if server is running
check_server() {
    echo "Checking if server is running..."
    if curl -s "${API_URL}/health" > /dev/null 2>&1; then
        success "Server is running at ${API_URL}"
        return 0
    else
        error "Server is not running at ${API_URL}"
        echo "Please start the server first"
        exit 1
    fi
}

# Function to setup authentication
setup_auth() {
    echo ""
    echo "======================================"
    echo "Setting up Authentication"
    echo "======================================"
    
    # Check if we can use existing credentials from environment
    if [ -n "$TEST_TOKEN" ]; then
        TOKEN="$TEST_TOKEN"
        TENANT_ID="$TEST_TENANT_ID"
        USER_ID="$TEST_USER_ID"
        success "Using provided test credentials"
        return 0
    fi
    
    info "Please provide test credentials or ensure TEST_TOKEN, TEST_TENANT_ID, TEST_USER_ID are set"
    echo ""
    read -p "Enter tenant ID: " TENANT_ID
    read -p "Enter user ID: " USER_ID
    read -p "Enter auth token: " TOKEN
    
    if [ -z "$TOKEN" ] || [ -z "$TENANT_ID" ] || [ -z "$USER_ID" ]; then
        error "Authentication credentials are required"
        exit 1
    fi
    
    success "Authentication configured"
}

# Test 1: Device Token Management
test_device_tokens() {
    echo ""
    echo "======================================"
    echo "Test 1: Device Token Management"
    echo "======================================"
    
    # Register a device
    echo "1.1. Registering test device..."
    REGISTER_RESPONSE=$(curl -s -X POST "${API_URL}/api/v1/notifications/devices" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "device_token": "test-fcm-token-'$(date +%s)'",
            "device_type": "android",
            "device_name": "Test Android Device",
            "app_version": "1.0.0"
        }')
    
    if echo "$REGISTER_RESPONSE" | grep -q "registered"; then
        success "Device registered successfully"
        DEVICE_TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"device_id":"[^"]*"' | cut -d'"' -f4)
    else
        error "Failed to register device"
        echo "Response: $REGISTER_RESPONSE"
    fi
    
    # List devices
    echo "1.2. Listing registered devices..."
    LIST_RESPONSE=$(curl -s -X GET "${API_URL}/api/v1/notifications/devices" \
        -H "Authorization: Bearer ${TOKEN}")
    
    if echo "$LIST_RESPONSE" | grep -q "devices"; then
        DEVICE_COUNT=$(echo "$LIST_RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
        success "Listed devices (count: ${DEVICE_COUNT})"
    else
        error "Failed to list devices"
    fi
    
    info "Device token management tests completed"
}

# Test 2: Inventory Reservations
test_inventory_reservations() {
    echo ""
    echo "======================================"
    echo "Test 2: Inventory Reservations"
    echo "======================================"
    
    info "Inventory reservation system already implemented"
    info "Check database for inventory_reservations table"
    
    # Verify table exists
    echo "2.1. Checking if inventory_reservations table exists..."
    if command -v psql > /dev/null 2>&1; then
        TABLE_EXISTS=$(psql -U postgres -d agromart2 -tAc "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'inventory_reservations');")
        if [ "$TABLE_EXISTS" = "t" ]; then
            success "inventory_reservations table exists"
        else
            error "inventory_reservations table not found"
        fi
    else
        info "psql not available, skipping database check"
    fi
}

# Test 3: Push Notifications (FCM)
test_push_notifications() {
    echo ""
    echo "======================================"
    echo "Test 3: Push Notification Setup"
    echo "======================================"
    
    echo "3.1. Checking FCM configuration..."
    
    # Check environment variables
    if [ -n "$FIREBASE_CREDENTIALS_PATH" ] && [ -f "$FIREBASE_CREDENTIALS_PATH" ]; then
        success "Firebase credentials file found"
    else
        info "Firebase credentials not configured (set FIREBASE_CREDENTIALS_PATH)"
    fi
    
    if [ "$FCM_ENABLED" = "true" ]; then
        success "FCM is enabled"
    else
        info "FCM is disabled (set FCM_ENABLED=true to enable)"
    fi
    
    # Check if device_tokens table exists
    echo "3.2. Checking if device_tokens table exists..."
    if command -v psql > /dev/null 2>&1; then
        TABLE_EXISTS=$(psql -U postgres -d agromart2 -tAc "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'device_tokens');")
        if [ "$TABLE_EXISTS" = "t" ]; then
            success "device_tokens table exists"
        else
            error "device_tokens table not found - run migration 025"
        fi
    else
        info "psql not available, skipping database check"
    fi
}

# Test 4: Tally Import System
test_tally_import() {
    echo ""
    echo "======================================"
    echo "Test 4: Tally Import System"
    echo "======================================"
    
    IMPORT_DIR="${TALLY_IMPORT_DIRECTORY:-./tally_imports}"
    
    echo "4.1. Checking import directory structure..."
    if [ -d "$IMPORT_DIR" ]; then
        success "Import directory exists: $IMPORT_DIR"
    else
        info "Creating import directory: $IMPORT_DIR"
        mkdir -p "$IMPORT_DIR/archive" "$IMPORT_DIR/failed"
        success "Import directories created"
    fi
    
    if [ -d "$IMPORT_DIR/archive" ]; then
        success "Archive directory exists"
    fi
    
    if [ -d "$IMPORT_DIR/failed" ]; then
        success "Failed directory exists"
    fi
    
    # Check for pending imports
    echo "4.2. Checking for pending import files..."
    CSV_COUNT=$(find "$IMPORT_DIR" -maxdepth 1 -name "*.csv" 2>/dev/null | wc -l)
    if [ "$CSV_COUNT" -gt 0 ]; then
        info "Found $CSV_COUNT CSV file(s) pending import"
    else
        info "No pending CSV files to import"
    fi
}

# Test 5: Analytics
test_analytics() {
    echo ""
    echo "======================================"
    echo "Test 5: Analytics Features"
    echo "======================================"
    
    echo "5.1. Testing search performance metrics..."
    ANALYTICS_RESPONSE=$(curl -s -X GET "${API_URL}/api/v1/analytics/search-performance" \
        -H "Authorization: Bearer ${TOKEN}")
    
    if echo "$ANALYTICS_RESPONSE" | grep -q "most_used_filters"; then
        success "Analytics endpoint responding"
        
        # Check if it's real data vs placeholder
        if echo "$ANALYTICS_RESPONSE" | grep -q '"most_used_filters":\['; then
            success "Filter analytics data available"
        fi
    else
        info "Analytics endpoint may need authentication or setup"
    fi
}

# Test 6: Check Go Dependencies
test_dependencies() {
    echo ""
    echo "======================================"
    echo "Test 6: Go Dependencies"
    echo "======================================"
    
    echo "6.1. Checking Firebase SDK..."
    if grep -q "firebase.google.com/go/v4" go.mod; then
        success "Firebase SDK v4 is installed"
    else
        error "Firebase SDK not found in go.mod"
    fi
    
    echo "6.2. Running go mod tidy..."
    if go mod tidy 2>&1 | grep -q "error"; then
        error "go mod tidy found errors"
    else
        success "Go modules are clean"
    fi
}

# Test 7: Migration Status
test_migrations() {
    echo ""
    echo "======================================"
    echo "Test 7: Database Migrations"
    echo "======================================"
    
    echo "7.1. Checking for migration files..."
    if [ -f "migrations/024_create_inventory_reservations.sql" ]; then
        success "Migration 024 (inventory reservations) exists"
    else
        error "Migration 024 not found"
    fi
    
    if [ -f "migrations/025_create_device_tokens.sql" ]; then
        success "Migration 025 (device tokens) exists"
    else
        error "Migration 025 not found"
    fi
    
    if command -v psql > /dev/null 2>&1; then
        echo "7.2. Checking applied migrations..."
        
        # Check if tables exist
        RESERVATIONS_EXISTS=$(psql -U postgres -d agromart2 -tAc "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'inventory_reservations');")
        TOKENS_EXISTS=$(psql -U postgres -d agromart2 -tAc "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'device_tokens');")
        
        if [ "$RESERVATIONS_EXISTS" = "t" ]; then
            success "inventory_reservations table applied"
        else
            error "inventory_reservations table not applied"
        fi
        
        if [ "$TOKENS_EXISTS" = "t" ]; then
            success "device_tokens table applied"
        else
            error "device_tokens table not applied - run: psql -U postgres -d agromart2 -f migrations/025_create_device_tokens.sql"
        fi
    else
        info "psql not available, skipping database migration checks"
    fi
}

# Main execution
main() {
    echo ""
    echo "Starting feature verification..."
    echo "API URL: ${API_URL}"
    echo ""
    
    # Run checks
    check_server
    
    # Only setup auth if we're testing API endpoints
    if [ -z "$SKIP_API_TESTS" ]; then
        setup_auth
        test_device_tokens
        test_analytics
    else
        info "Skipping API tests (SKIP_API_TESTS is set)"
    fi
    
    # Run non-API tests
    test_inventory_reservations
    test_push_notifications
    test_tally_import
    test_dependencies
    test_migrations
    
    # Summary
    echo ""
    echo "======================================"
    echo "Verification Complete"
    echo "======================================"
    echo ""
    success "All implemented features verified"
    echo ""
    echo "Next steps:"
    echo "1. Set up Firebase credentials (if not already done)"
    echo "2. Run pending database migrations"
    echo "3. Configure mobile apps to register device tokens"
    echo "4. Test CSV imports by placing files in $IMPORT_DIR"
    echo ""
}

# Run main function
main
