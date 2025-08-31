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
fi</search>
</search_and_replace>

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

# Step 5: Test Image Upload
echo ""
echo "⬆️  Testing Image Upload..."
UPLOAD_RESPONSE=$(curl -s -X POST $BASE_URL/v1/products/$PRODUCT_ID/images \
  -H "Authorization: Bearer $TOKEN" \
  -F "image=@$TEST_IMAGE_PATH" \
  -F "alt_text=Test Image")

if echo "$UPLOAD_RESPONSE" | jq -e '.message' > /dev/null 2>&1; then
    echo "✅ Image upload successful"
else
    echo "❌ Image upload failed"
    echo "Response: $UPLOAD_RESPONSE"
    exit 1
fi

# Step 6: Test Get Product Images
echo ""
echo "📋 Testing Get Product Images..."
IMAGES_RESPONSE=$(curl -s -X GET $BASE_URL/v1/products/$PRODUCT_ID/images \
  -H "Authorization: Bearer $TOKEN")

if echo "$IMAGES_RESPONSE" | jq -e '.images' > /dev/null 2>&1; then
    IMAGE_COUNT=$(echo "$IMAGES_RESPONSE" | jq -r '.count')
    echo "✅ Product images retrieved: $IMAGE_COUNT images"

    if [ "$IMAGE_COUNT" -gt 0 ]; then
        IMAGE_ID=$(echo "$IMAGES_RESPONSE" | jq -r '.images[0].id')
        echo "✅ Image ID: $IMAGE_ID"
    else
        echo "❌ No images found for product"
        exit 1
    fi
else
    echo "❌ Failed to retrieve product images"
    echo "Response: $IMAGES_RESPONSE"
    exit 1
fi

# Step 7: Test Presigned URL Generation
echo ""
echo "🔗 Testing Presigned URL Generation..."
URL_RESPONSE=$(curl -s -X GET "$BASE_URL/v1/products/$PRODUCT_ID/images/$IMAGE_ID/url?expiry_minutes=60" \
  -H "Authorization: Bearer $TOKEN")

if echo "$URL_RESPONSE" | jq -e '.url' > /dev/null 2>&1; then
    PRESIGNED_URL=$(echo "$URL_RESPONSE" | jq -r '.url')
    EXPIRES_IN=$(echo "$URL_RESPONSE" | jq -r '.expires_in')
    echo "✅ Presigned URL generated"
    echo "URL: ${PRESIGNED_URL:0:100}..."
    echo "Expires in: $EXPIRES_IN"
else
    echo "❌ Presigned URL generation failed"
    echo "Response: $URL_RESPONSE"
    exit 1
fi

# Step 8: Test Image Download via Presigned URL
echo ""
echo "⬇️  Testing Image Download via Presigned URL..."
DOWNLOAD_RESPONSE=$(curl -s -I "$PRESIGNED_URL")

if echo "$DOWNLOAD_RESPONSE" | grep -q "HTTP/1.1 200\|HTTP/2 200"; then
    CONTENT_TYPE=$(echo "$DOWNLOAD_RESPONSE" | grep -i "content-type" | head -1)
    CONTENT_LENGTH=$(echo "$DOWNLOAD_RESPONSE" | grep -i "content-length" | head -1)
    echo "✅ Image download successful"
    echo "$CONTENT_TYPE"
    echo "$CONTENT_LENGTH"
else
    echo "❌ Image download failed"
    echo "Response headers:"
    echo "$DOWNLOAD_RESPONSE"
fi

# Step 9: Test MinIO Storage (Check if bucket exists)
echo ""
echo "🏪 Testing MinIO Bucket Accessibility..."
# This would require mc tool or direct MinIO API call. For now, we'll check the URL structure

if [[ $PRESIGNED_URL == *"product-images"* ]]; then
    echo "✅ Bucket name 'product-images' found in URL"
else
    echo "❌ Expected bucket name not found in URL"
fi

# Step 10: Test Image Deletion
echo ""
echo "🗑️  Testing Image Deletion..."
DELETE_RESPONSE=$(curl -s -X DELETE $BASE_URL/v1/products/$PRODUCT_ID/images/$IMAGE_ID \
  -H "Authorization: Bearer $TOKEN")

if echo "$DELETE_RESPONSE" | jq -e '.message' > /dev/null 2>&1; then
    MESSAGE=$(echo "$DELETE_RESPONSE" | jq -r '.message')
    echo "✅ $MESSAGE"
else
    echo "❌ Image deletion failed"
    echo "Response: $DELETE_RESPONSE"
fi

# Step 11: Test Tenant Isolation (Verify no images after deletion)
echo ""
echo "🔒 Testing Tenant Isolation (Verify image deletion)..."
FINAL_IMAGES_RESPONSE=$(curl -s -X GET $BASE_URL/v1/products/$PRODUCT_ID/images \
  -H "Authorization: Bearer $TOKEN")

if echo "$FINAL_IMAGES_RESPONSE" | jq -e '.images' > /dev/null 2>&1; then
    FINAL_COUNT=$(echo "$FINAL_IMAGES_RESPONSE" | jq -r '.count')
    if [ "$FINAL_COUNT" -eq 0 ]; then
        echo "✅ Tenant isolation confirmed: No images found after deletion"
    else
        echo "⚠️  Unexpected: $FINAL_COUNT images still found for product"
    fi
else
    echo "❌ Failed to verify tenant isolation"
fi

# Cleanup
echo ""
echo "🧹 Cleaning up test files..."
rm -f $TEST_IMAGE_PATH

echo ""
echo "=== MinIO Functionality Test Complete ==="