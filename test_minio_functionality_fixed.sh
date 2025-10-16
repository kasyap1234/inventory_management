#!/bin/bash

# Test MinIO File Upload/Download Functionality for Product Images
# This script tests the complete MinIO integration workflow

echo "=== MinIO Functionality Test Script ==="
echo "Testing product image upload, storage, and retrieval"
echo ""

BASE_URL="http://localhost:8081"
TEST_IMAGE_PATH="test_image.jpg"

# Step 1: Check if required tools are available
if ! command -v curl &> /dev/null; then
    echo "❌ curl command not found. Please install curl."
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "❌ jq command not found. Please install jq for JSON parsing."
    exit 1
fi

# Step 2: Create test image file
echo "📸 Creating test image file..."
echo -e "\xFF\xD8\xFF\xE0\x00\x10JFIF\x00\x01\x01\x01\x00H\x00H\x00\x00\xFF\xC0\x00\x11\x08\x00\x10\x00\x10\x03\x01\x11\x00\x02\x11\x01\x03\x11\x01\xFF\xC4\x00\x14\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x08\xFF\xDB\x00C\x00\x08\x06\x06\x07\x06\x05\x08\x07\x07\x07\t\t\x08\n\x0C\x14\r\x0C\x0B\x0B\x0C\x19\x12\x13\x0F\x14\x1D\x1A\x1F\x1E\x1D\x1A\x1C\x1C $.' ',#\x26)O\x1C\x1C (;7\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\x3C\xFF\xDA\x00\x0C\x03\x01\x00\x02\x11\x03\x11\x00\x3F\x00\x00\x00\x00\xff" >> $TEST_IMAGE_PATH
echo "✅ Test image created"

# Step 3: Authentication
echo ""
echo "🔐 Testing Authentication..."
# Create test user with timestamped email
EMAIL="test$(date +%s)@example.com"
USERNAME="testuser$(date +%s)"

SIGNUP_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'"$EMAIL"'",
    "password": "password123",
    "username": "'"$USERNAME"'",
    "first_name": "Test",
    "last_name": "User"
  }')

if echo "$SIGNUP_RESPONSE" | jq -e '.token' > /dev/null 2>&1; then
    TOKEN=$(echo "$SIGNUP_RESPONSE" | jq -r '.token')
    echo "✅ Signup successful"
else
    echo "❌ Signup failed"
    echo "Signup response: $SIGNUP_RESPONSE"
    exit 1
fi

echo "Token obtained: ${TOKEN:0:20}..."

# Step 4: Test Product Creation
echo ""
echo "📦 Testing Product Creation..."
PRODUCT_RESPONSE=$(curl -s -X POST $BASE_URL/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product for MinIO",
    "unit_price": 99.99,
    "quantity": 10,
    "description": "Product for testing MinIO image upload functionality"
  }')

if echo "$PRODUCT_RESPONSE" | jq -e '.product.id' > /dev/null 2>&1; then
    PRODUCT_ID=$(echo "$PRODUCT_RESPONSE" | jq -r '.product.id')
    echo "✅ Product created with ID: $PRODUCT_ID"
else
    echo "❌ Product creation failed"
    echo "Response: $PRODUCT_RESPONSE"
    exit 1
fi

echo ""
echo "=== MinIO Test Complete (Partial - Basic Setup Working) ==="
echo "Product created: $PRODUCT_ID"
echo "Token: ${TOKEN:0:20}..."
echo "Next steps: Image upload and MinIO functionality tests"