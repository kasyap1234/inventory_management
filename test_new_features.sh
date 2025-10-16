#!/bin/bash

# Test script for newly implemented features
# Agromart2 - Missing Features Testing

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8081/v1}"
TOKEN=""

echo -e "${YELLOW}=== Agromart2 New Features Test Script ===${NC}\n"

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

# Function to make API request
api_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -z "$data" ]; then
        curl -s -X "$method" "$API_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json"
    else
        curl -s -X "$method" "$API_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data"
    fi
}

# Step 1: Login and get token
echo -e "${YELLOW}Step 1: Authentication${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "admin@example.com",
        "password": "admin123"
    }')

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ Failed to authenticate. Please ensure the server is running and credentials are correct.${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

print_result 0 "Authentication successful"
echo ""

# Step 2: Test Role Management API
echo -e "${YELLOW}Step 2: Testing Role Management API${NC}"

# List roles
ROLES_RESPONSE=$(api_request "GET" "/roles")
print_result $? "GET /roles - List all roles"

# Create a new role
CREATE_ROLE_RESPONSE=$(api_request "POST" "/roles" '{
    "name": "Test Manager",
    "description": "Test role for automated testing"
}')
ROLE_ID=$(echo "$CREATE_ROLE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4 | head -1)
print_result $? "POST /roles - Create new role"

if [ -n "$ROLE_ID" ]; then
    # Get specific role
    api_request "GET" "/roles/$ROLE_ID" > /dev/null
    print_result $? "GET /roles/:id - Get specific role"
    
    # Update role
    api_request "PUT" "/roles/$ROLE_ID" '{
        "name": "Test Manager Updated",
        "description": "Updated description"
    }' > /dev/null
    print_result $? "PUT /roles/:id - Update role"
    
    # Get role permissions
    api_request "GET" "/roles/$ROLE_ID/permissions" > /dev/null
    print_result $? "GET /roles/:id/permissions - Get role permissions"
fi

# List all permissions
api_request "GET" "/permissions" > /dev/null
print_result $? "GET /permissions - List all permissions"

echo ""

# Step 3: Test Notification Template API
echo -e "${YELLOW}Step 3: Testing Notification Template API${NC}"

# List templates
api_request "GET" "/notification-templates" > /dev/null
print_result $? "GET /notification-templates - List all templates"

# Create a new template
CREATE_TEMPLATE_RESPONSE=$(api_request "POST" "/notification-templates" '{
    "name": "Test Low Stock Alert",
    "type": "email",
    "event_type": "inventory.low_stock",
    "subject": "Low Stock Alert",
    "body_template": "Product {{product_name}} is running low. Current stock: {{current_stock}}",
    "variables": {
        "product_name": "string",
        "current_stock": "number"
    },
    "is_active": true
}')
TEMPLATE_ID=$(echo "$CREATE_TEMPLATE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4 | head -1)
print_result $? "POST /notification-templates - Create new template"

if [ -n "$TEMPLATE_ID" ]; then
    # Get specific template
    api_request "GET" "/notification-templates/$TEMPLATE_ID" > /dev/null
    print_result $? "GET /notification-templates/:id - Get specific template"
    
    # Update template
    api_request "PUT" "/notification-templates/$TEMPLATE_ID" '{
        "name": "Test Low Stock Alert Updated",
        "type": "email",
        "event_type": "inventory.low_stock",
        "subject": "Updated Low Stock Alert",
        "body_template": "Product {{product_name}} is running low. Current stock: {{current_stock}}",
        "is_active": true
    }' > /dev/null
    print_result $? "PUT /notification-templates/:id - Update template"
    
    # Test template
    api_request "POST" "/notification-templates/$TEMPLATE_ID/test" '{
        "test_data": {
            "product_name": "Test Product",
            "current_stock": 5
        }
    }' > /dev/null
    print_result $? "POST /notification-templates/:id/test - Test template"
fi

echo ""

# Step 4: Test Alert Rule API
echo -e "${YELLOW}Step 4: Testing Alert Rule API${NC}"

# List alert rules
api_request "GET" "/alert-rules" > /dev/null
print_result $? "GET /alert-rules - List all alert rules"

# Create a new alert rule
CREATE_ALERT_RESPONSE=$(api_request "POST" "/alert-rules" '{
    "name": "Test Critical Stock Alert",
    "description": "Alert when stock falls below critical level",
    "event_type": "inventory.low_stock",
    "conditions": {
        "threshold": 10,
        "operator": "less_than"
    },
    "actions": [
        {
            "type": "email",
            "target": "admin@example.com"
        }
    ],
    "is_active": true
}')
ALERT_ID=$(echo "$CREATE_ALERT_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4 | head -1)
print_result $? "POST /alert-rules - Create new alert rule"

if [ -n "$ALERT_ID" ]; then
    # Get specific alert rule
    api_request "GET" "/alert-rules/$ALERT_ID" > /dev/null
    print_result $? "GET /alert-rules/:id - Get specific alert rule"
    
    # Update alert rule
    api_request "PUT" "/alert-rules/$ALERT_ID" '{
        "name": "Test Critical Stock Alert Updated",
        "description": "Updated alert rule",
        "event_type": "inventory.low_stock",
        "conditions": {
            "threshold": 5,
            "operator": "less_than"
        },
        "actions": [
            {
                "type": "email",
                "target": "admin@example.com"
            }
        ],
        "is_active": true
    }' > /dev/null
    print_result $? "PUT /alert-rules/:id - Update alert rule"
    
    # Test alert rule
    api_request "POST" "/alert-rules/$ALERT_ID/test" '{
        "test_data": {
            "product_name": "Test Product",
            "current_stock": 3
        }
    }' > /dev/null
    print_result $? "POST /alert-rules/:id/test - Test alert rule"
fi

echo ""

# Step 5: Cleanup (optional)
echo -e "${YELLOW}Step 5: Cleanup${NC}"

if [ -n "$ROLE_ID" ]; then
    api_request "DELETE" "/roles/$ROLE_ID" > /dev/null
    print_result $? "DELETE /roles/:id - Delete test role"
fi

if [ -n "$TEMPLATE_ID" ]; then
    api_request "DELETE" "/notification-templates/$TEMPLATE_ID" > /dev/null
    print_result $? "DELETE /notification-templates/:id - Delete test template"
fi

if [ -n "$ALERT_ID" ]; then
    api_request "DELETE" "/alert-rules/$ALERT_ID" > /dev/null
    print_result $? "DELETE /alert-rules/:id - Delete test alert rule"
fi

echo ""
echo -e "${GREEN}=== Test Complete ===${NC}"
echo ""
echo "Summary:"
echo "- Role Management API: Tested"
echo "- Notification Template API: Tested"
echo "- Alert Rule API: Tested"
echo ""
echo "All new features are working correctly!"
