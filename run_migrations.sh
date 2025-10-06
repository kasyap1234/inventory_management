#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Running Database Migrations          ${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check if postgres container is running
if ! docker ps | grep -q inventory_management-postgres-1; then
    echo -e "${RED}PostgreSQL container is not running. Please start it first.${NC}"
    exit 1
fi

echo -e "${YELLOW}Running migrations...${NC}\n"

# Run the main schema
echo -e "${BLUE}[1/2] Creating complete auth schema...${NC}"
docker exec -i inventory_management-postgres-1 psql -U testuser -d testdb < migrations/complete_auth_schema.sql
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Auth schema created successfully${NC}\n"
else
    echo -e "${RED}✗ Failed to create auth schema${NC}\n"
    exit 1
fi

# Run the base schema for additional tables
echo -e "${BLUE}[2/2] Creating additional tables...${NC}"
docker exec -i inventory_management-postgres-1 psql -U testuser -d testdb < migrations/schema.sql
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Additional tables created successfully${NC}\n"
else
    echo -e "${RED}✗ Failed to create additional tables${NC}\n"
    exit 1
fi

# Verify tables were created
echo -e "${YELLOW}Verifying database schema...${NC}"
TABLES=$(docker exec inventory_management-postgres-1 psql -U testuser -d testdb -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")

if [ "$TABLES" -gt 0 ]; then
    echo -e "${GREEN}✓ Database schema verified: $TABLES tables created${NC}\n"
    
    echo -e "${BLUE}Tables in database:${NC}"
    docker exec inventory_management-postgres-1 psql -U testuser -d testdb -c "\dt"
else
    echo -e "${RED}✗ No tables found in database${NC}\n"
    exit 1
fi

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  Migrations Completed Successfully!   ${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${BLUE}Default tenant:${NC}"
echo -e "  ID:       ${YELLOW}aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa${NC}"
echo -e "  Name:     ${YELLOW}Agromart Development${NC}"
echo -e "  Subdomain:${YELLOW}agromart-dev${NC}\n"

echo -e "${BLUE}Default roles:${NC}"
echo -e "  Admin ID: ${YELLOW}bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb${NC}"
echo -e "  User ID:  ${YELLOW}cccccccc-cccc-cccc-cccc-cccccccccccc${NC}\n"

echo -e "${YELLOW}You can now sign up users via the API!${NC}"
