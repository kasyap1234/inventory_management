#!/bin/bash

# API Endpoint Testing Script for Agromart2
# Set up environment variables first:
# export DATABASE_URL="..."
# export REDIS_URL="..."
# export JWT_SECRET="..."
# Then run this script

BASE_URL="http://localhost:8080"
API_VERSION="/v1"
AUTH_HEADER=""

echo "=== AGROMART2 API ENDPOINT TESTING ===\n"

# 1. Test Health Endpoints (no auth required)
echo "1. TESTING HEALTH ENDPOINTS (no auth required)"
echo "======================================"

echo "Health check:"
curl -s -X GET "$BASE_URL/health" | jq || curl -s -X GET "$BASE_URL/health"

echo -e "\nReadiness check:"
curl -s -X GET "$BASE_URL/health/ready" | jq || curl -s -X GET "$BASE_URL/health/ready"

echo -e "\nLiveness check:"
curl -s -X GET "$BASE_URL/health/live" | jq || curl -s -X GET "$BASE_URL/health/live"

echo -e "\nDetailed health check:"
curl -s -X GET "$BASE_URL/health/detailed" | jq || curl -s -X GET "$BASE_URL/health/detailed"

echo -e "\nMetrics:"
curl -s -X GET "$BASE_URL/metrics" | jq || curl -s -X GET "$BASE_URL/metrics"

# 2. Test Documentation Endpoints
echo -e "\n2. TESTING DOCUMENTATION ENDPOINTS"
echo "=================================="

echo -e "\nDocs guide:"
curl -s -X GET "$BASE_URL$API_VERSION/docs/guide" | jq || curl -s -X GET "$BASE_URL$API_VERSION/docs/guide"

echo -e "\nSwagger spec:"
curl -s -X GET "$BASE_URL$API_VERSION/docs/spec" | head -10  # Just first few lines

# 3. Test API Root
echo -e "\n3. TESTING API ROOT ENDPOINT"
echo "============================"

echo -e "\nAPI Info:"
curl -s -X GET "$BASE_URL$API_VERSION/" | jq || curl -s -X GET "$BASE_URL$API_VERSION/"

# 4. Test Authentication Endpoints
echo -e "\n4. TESTING AUTHENTICATION ENDPOINTS"
echo "==================================="

# Sign up a new user (using /v1/auth/* endpoints - no JWT middleware)
echo -e "\nUser Signup:"
SIGNUP_RESPONSE=$(curl -s -X POST "$BASE_URL$API_VERSION/auth/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!",
    "first_name": "Test",
    "last_name": "User"
  }' | jq)

echo "$SIGNUP_RESPONSE"

# Login to get JWT token (using /v1/auth/* endpoints - no JWT middleware)
echo -e "\nUser Login:"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL$API_VERSION/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!"
  }' | jq)

echo "$LOGIN_RESPONSE"

# Extract JWT token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
    AUTH_HEADER="Authorization: Bearer $TOKEN"
    echo "✅ Successfully obtained JWT token"
else
    echo "❌ Failed to obtain JWT token from login response"
    AUTH_HEADER=""
fi

# Test logout (use v1 endpoint since it needs JWT middleware)
if [ "$AUTH_HEADER" != "" ]; then
    echo -e "\nUser Logout:"
    curl -s -X POST "$BASE_URL$API_VERSION/auth/logout" \
      -H "$AUTH_HEADER" \
      -H "Content-Type: application/json" | jq
fi

# 5. Test Tenant Endpoints (require auth)
echo -e "\n5. TESTING TENANT ENDPOINTS (auth required)"
echo "========================================"

if [ "$AUTH_HEADER" != "" ]; then
    echo -e "\nList Tenants:"
    curl -s -X GET "$BASE_URL$API_VERSION/tenants" \
      -H "$AUTH_HEADER" | jq

    echo -e "\nGet Tenant by Subdomain:"
    curl -s -X GET "$BASE_URL$API_VERSION/tenants/by-subdomain/agromart" \
      -H "$AUTH_HEADER" | jq
else
    echo "❌ Skipping tenant tests - no auth token"
fi

# 6. Test User Endpoints
echo -e "\n6. TESTING USER ENDPOINTS (auth required)"
echo "======================================"

if [ "$AUTH_HEADER" != "" ]; then
    echo -e "\nList Users:"
    curl -s -X GET "$BASE_URL$API_VERSION/users" \
      -H "$AUTH_HEADER" | jq

    echo -e "\nCreate User:"
    curl -s -X POST "$BASE_URL$API_VERSION/users" \
      -H "$AUTH_HEADER" \
      -H "Content-Type: application/json" \
      -d '{
        "email": "admin@example.com",
        "first_name": "Admin",
        "last_name": "User",
        "status": "active"
      }' | jq
else
    echo "❌ Skipping user tests - no auth token"
fi
# 6. Test Category Endpoints
echo -e "\n6. TESTING CATEGORY ENDPOINTS (auth required)"
echo "========================================"

if [ "$AUTH_HEADER" != "" ]; then
    echo -e "\nList Categories:"
    curl -s -X GET "$BASE_URL$API_VERSION/categories" \
      -H "$AUTH_HEADER" | jq

    echo -e "\nCreate Category:"
    curl -s -X POST "$BASE_URL$API_VERSION/categories" \
      -H "$AUTH_HEADER" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "Fruits",
        "description": "Fresh fruits category"
      }' | jq
else
    echo "❌ Skipping category tests - no auth token"
fi

# 7. Test User Endpoints

# 8. Test Product Endpoints
echo -e "\n7. TESTING PRODUCT ENDPOINTS (auth required)"
echo "========================================="

if [ "$AUTH_HEADER" != "" ]; then
    echo -e "\nList Products:"
    curl -s -X GET "$BASE_URL$API_VERSION/products" \
      -H "$AUTH_HEADER" | jq

    echo -e "\nCreate Product:"
    curl -s -X POST "$BASE_URL$API_VERSION/products" \
      -H "$AUTH_HEADER" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "Test Product",
        "category_id": "00000000-0000-0000-0000-000000000001",
        "quantity": 100,
        "unit_price": 25.50,
        "description": "Test product for API validation"
      }' | jq

    echo -e "\nSearch Products:"
    curl -s -X GET "$BASE_URL$API_VERSION/products/search?q=test" \
      -H "$AUTH_HEADER" | jq
else
    echo "❌ Skipping product tests - no auth token"
fi

# 9. Test Protected Legacy Route
echo -e "\n8. TESTING PROTECTED LEGACY ROUTE"
echo "================================="

if [ "$AUTH_HEADER" != "" ]; then
    echo -e "\nLegacy Protected Route:"
    curl -s -X GET "$BASE_URL/protected" \
      -H "$AUTH_HEADER" | jq
else
    echo "❌ Skipping legacy route test - no auth token"
fi

# 10. Test Bulk Operations (not implemented yet)
echo -e "\n9. TESTING BULK OPERATIONS (not fully implemented)"
echo "================================================="

if [ "$AUTH_HEADER" != "" ]; then
    echo -e "\nBulk Update:"
    curl -s -X POST "$BASE_URL$API_VERSION/products/bulk/update" \
      -H "$AUTH_HEADER" \
      -H "Content-Type: application/json" | jq

    echo -e "\nBulk Create:"
    curl -s -X POST "$BASE_URL$API_VERSION/products/bulk/create" \
      -H "$AUTH_HEADER" \
      -H "Content-Type: application/json" | jq
else
    echo "❌ Skipping bulk operations tests - no auth token"
fi

# 11. Test Tenant Lookup (no auth required)
echo -e "\n10. TESTING TENANT LOOKUP (no auth required)"
echo "============================================="

echo -e "\nTenant by Subdomain (no auth):"
curl -s -X GET "$BASE_URL$API_VERSION/tenants/by-subdomain/agromart" | jq

echo -e "\n=== TESTING COMPLETE ==="
echo "======================"
echo "Summary:"
echo "- Health endpoints: Tested ✅"
echo "- Documentation endpoints: Tested ✅"
echo "- API root: Tested ✅"
echo "- Authentication endpoints: Tested ✅ (signup/login)"
echo "- Protected endpoints: $([ "$AUTH_HEADER" != "" ] && echo "Tested ✅" || echo "Skipped ❌ (no token)")"
echo ""
echo "To run this script:"
echo "1. Set up your database and Redis"
echo "2. Start the server: ./main"
echo "3. Run: chmod +x test_api_endpoints.sh && ./test_api_endpoints.sh"
echo ""
echo "Note: Some endpoints may return database errors if not properly set up."