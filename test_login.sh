#!/bin/bash

echo "=== Testing Login Flow ==="
echo ""

# Get CSRF token
echo "1. Getting CSRF token..."
CSRF_RESPONSE=$(curl -s -c cookies.txt http://localhost:8081/v1/security/csrf)
CSRF_TOKEN=$(echo "$CSRF_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('csrf_token', ''))" 2>/dev/null)
echo "CSRF Token: ${CSRF_TOKEN:0:20}..."
echo ""

# Try login
echo "2. Attempting login with kasyap3103@gmail.com..."
LOGIN_RESPONSE=$(curl -s -b cookies.txt -c cookies.txt -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -d '{"email":"kasyap3103@gmail.com","password":"password123"}')

echo "Response:"
echo "$LOGIN_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$LOGIN_RESPONSE"
echo ""

# Check if access_token exists
HAS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token"' | wc -l)
if [ "$HAS_TOKEN" -gt 0 ]; then
  echo "✅ Login successful! Access token received."
  
  # Extract token
  ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('access_token', ''))" 2>/dev/null)
  
  # Test authenticated endpoint
  echo ""
  echo "3. Testing authenticated endpoint /me..."
  USER_RESPONSE=$(curl -s -b cookies.txt -H "Authorization: Bearer $ACCESS_TOKEN" -H "X-CSRF-Token: $CSRF_TOKEN" http://localhost:8081/v1/me)
  echo "$USER_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$USER_RESPONSE"
else
  echo "❌ Login failed. Check the response above for error details."
fi

echo ""
rm -f cookies.txt
