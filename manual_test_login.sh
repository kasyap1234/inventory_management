#!/bin/bash

echo "=== Manual Login Test ==="
echo ""

# Step 1: Get CSRF token
echo "1. Getting CSRF token..."
CSRF_JSON=$(curl -s http://localhost:8081/v1/security/csrf)
echo "Response: $CSRF_JSON"

CSRF_TOKEN=$(echo "$CSRF_JSON" | python3 -c "import sys, json; data = json.load(sys.stdin); print(data.get('token', ''))" 2>/dev/null)
echo "Extracted Token: ${CSRF_TOKEN:0:40}..."
echo ""

# Step 2: Test login
echo "2. Testing login..."
echo "Trying with password: password123"
LOGIN_RESULT=$(curl -s -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -d '{"email":"kasyap3103@gmail.com","password":"password123"}')

echo "Login response:"
echo "$LOGIN_RESULT" | python3 -m json.tool 2>/dev/null || echo "$LOGIN_RESULT"
echo ""

# Check success
if echo "$LOGIN_RESULT" | grep -q '"access_token"'; then
  echo "✅ Login successful!"
else
  echo "❌ Login failed"
  echo ""
  echo "Common passwords to try:"
  echo "  - password123"
  echo "  - Password123"
  echo "  - Password123!"
  echo "  - admin123"
  echo ""
  echo "You can reset the password with:"
  echo "  psql \"postgresql://testuser:testpass@localhost:5440/testdb\" -c \\"
  echo "    \"UPDATE users SET password_hash = '\$2a\$10\$...' WHERE email = 'kasyap3103@gmail.com';\""
fi

rm -f cookies.txt
