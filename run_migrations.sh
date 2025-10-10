#!/bin/bash

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

POSTGRES_CONTAINER="inventory_management-postgres-1"
POSTGRES_USER="testuser"
POSTGRES_DB="testdb"
MIGRATIONS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/migrations"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Running Database Migrations          ${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Ensure the postgres container is running
if ! docker ps --format '{{.Names}}' | grep -qx "$POSTGRES_CONTAINER"; then
    echo -e "${RED}PostgreSQL container ($POSTGRES_CONTAINER) is not running. Please start it first.${NC}"
    exit 1
fi

echo -e "${YELLOW}Waiting for PostgreSQL to accept connections...${NC}"
until docker exec "$POSTGRES_CONTAINER" pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; do
    sleep 1
done
echo -e "${GREEN}✓ PostgreSQL is ready${NC}\n"

# Ensure required extensions exist
echo -e "${YELLOW}Ensuring required extensions are available...${NC}"
docker exec -i "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
echo -e "${GREEN}✓ Extensions verified${NC}\n"

echo -e "${YELLOW}Applying SQL migrations...${NC}\n"

apply_migration() {
    local file_path="$1"
    local file_name
    file_name="$(basename "$file_path")"

    echo -e "${BLUE}→ Applying ${file_name}...${NC}"
    if cat "$file_path" | docker exec -i "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 >/dev/null; then
        echo -e "${GREEN}   ✓ ${file_name} applied${NC}\n"
    else
        echo -e "${RED}   ✗ Failed to apply ${file_name}${NC}"
        exit 1
    fi
}

declare -a ORDERED_MIGRATIONS=(
    "complete_auth_schema.sql"
    "schema.sql"
    "20250831110324_create_business_tables_fixed.sql"
    "20240831_fix_product_permissions.sql"
    "20250831170400_fix_user_tenant_constraints.sql"
    "20250201120000_add_password_hash_to_users.sql"
    "20250831170500_add_invoice_fields_and_sequence.sql"
    "20250831171700_add_category_hierarchy.sql"
    "20250831171701_insert_sample_categories.sql"
    "20250831180000_fix_email_uniqueness_constraint.sql"
    "20250901120000_create_analytics_tables.sql"
    "20240130000001_optimize_search_indexes.sql"
    "20251009120000_add_performance_indexes.sql"
)

declare -a SKIP_MIGRATIONS=(
    "20240130000002_table_partitioning.sql"
    "20250107_performance_indexes.sql"
    "20250831110324_create_business_tables.sql"
)

should_skip() {
    local filename="$1"
    for skip in "${SKIP_MIGRATIONS[@]}"; do
        if [[ "$filename" == "$skip" ]]; then
            return 0
        fi
    done
    return 1
}

for filename in "${ORDERED_MIGRATIONS[@]}"; do
    migration_path="$MIGRATIONS_DIR/$filename"
    if [ -f "$migration_path" ]; then
        apply_migration "$migration_path"
    fi
done

while IFS= read -r -d '' migration_file; do
    basename_file="$(basename "$migration_file")"
    skip=false
    for ordered in "${ORDERED_MIGRATIONS[@]}"; do
        if [[ "$basename_file" == "$ordered" ]]; then
            skip=true
            break
        fi
    done
    if ! $skip && ! should_skip "$basename_file"; then
        apply_migration "$migration_file"
    fi
done < <(find "$MIGRATIONS_DIR" -maxdepth 1 -type f -name '*.sql' -print0 | sort -z)

echo -e "${YELLOW}Verifying database schema...${NC}"
TABLE_COUNT=$(docker exec "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | tr -d '[:space:]')

if [[ "$TABLE_COUNT" =~ ^[0-9]+$ && "$TABLE_COUNT" -gt 0 ]]; then
    echo -e "${GREEN}✓ Database schema verified: ${TABLE_COUNT} tables present${NC}\n"
else
    echo -e "${RED}✗ Failed to verify database schema${NC}"
    exit 1
fi

echo -e "${BLUE}Tables in database:${NC}"
docker exec "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "\dt"

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
