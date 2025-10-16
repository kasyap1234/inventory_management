#!/bin/bash

# Debug script for MinIO functionality hotspot
BASE_URL="http://localhost:8081"
# Get fresh token
echo "Getting fresh JWT token..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!"
  }')

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
if [ "$ACCESS_TOKEN" == "null" ] || [ "$ACCESS_TOKEN" == "" ]; then
    echo "❌ Failed to get fresh token"
    exit 1
fi
echo "✅ Fresh token obtained"

echo "=== DEBUGGING MinIO IMAGE RETRIEVAL ==="
echo "Using access token: ${ACCESS_TOKEN:0:20}..."
echo ""

# Test 1: Direct product lookup
echo "Test 1: Getting existing products to verify authentication..."
PRODUCTS_RESPONSE=$(curl -s -X GET $BASE_URL/v1/products \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.')

echo "Products response:"
echo "$PRODUCTS_RESPONSE | jq '.'"
echo ""

# Test 2: Check if there are any existing products
if echo "$PRODUCTS_RESPONSE" | jq -e '.products' > /dev/null 2>&1; then
    PRODUCT_COUNT=$(echo "$PRODUCTS_RESPONSE" | jq '.products | length')
    if [ "$PRODUCT_COUNT" -gt 0 ]; then
        PRODUCT_ID=$(echo "$PRODUCTS_RESPONSE" | jq -r '.products[0].id')
        echo "Found existing product ID: $PRODUCT_ID"

        echo ""
        echo "Test 3: Checking images for product $PRODUCT_ID..."
        echo "Request URL: $BASE_URL/v1/products/$PRODUCT_ID/images"
        IMAGES_RESPONSE=$(curl -s -X GET $BASE_URL/v1/products/$PRODUCT_ID/images \
          -H "Authorization: Bearer $ACCESS_TOKEN" -w "\nHTTP Status: %{http_code}\n" \
          -v)

        echo "Images response:"
        # Check if response is JSON before parsing
        if [[ $IMAGES_RESPONSE == "{"* ]]; then
            echo "$IMAGES_RESPONSE" | jq '.'
        else
            echo "Raw response: $IMAGES_RESPONSE"
        fi

        # Try the image upload endpoint with the product ID
        echo ""
        echo "Test 4: Attempting to upload image to confirm product ID works..."
        # Create minimal test image
        echo -e "\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x01\x00H\x00H\x00\x00\xff\xc0\x00\x11\x08\x00\x10\x00\x10\x03\x01\x11\x00\x02\x11\x01\x03\x11\x01\xff\xc4\x00\x1f\x00\x01\x05\x01\x01\x01\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\xff\xc4\x00\x1f\x10\x01\x05\x01\x01\x01\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\xff\xda\x00\x0c\x03\x01\x00\x02\x11\x03\x11\x00\x3f\x00\xab\xba\xcf\x00\xff\xd9" > test_upload.jpg

        UPLOAD_RESPONSE=$(curl -s -X POST $BASE_URL/v1/products/$PRODUCT_ID/images \
          -H "Authorization: Bearer $ACCESS_TOKEN" \
          -F "image=@test_upload.jpg" \
          -F "alt_text=Debug Test Image" \
          -w "\nHTTP Status: %{http_code}\n")

        echo "Upload response:"
        if [[ $UPLOAD_RESPONSE == "{"* ]]; then
            echo "$UPLOAD_RESPONSE" | jq '.'
        else
            echo "Raw response: $UPLOAD_RESPONSE"
        fi
    else
        echo "No existing products found"
    fi
else
    echo "Could not retrieve products"
fi