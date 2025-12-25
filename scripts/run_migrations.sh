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
MAX_RETRIES=30
RETRY_COUNT=0
until docker exec "$POSTGRES_CONTAINER" pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1 && \
      docker exec "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT 1" >/dev/null 2>&1; do
    RETRY_COUNT=$((RETRY_COUNT+1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        echo -e "${RED}PostgreSQL failed to become ready after $MAX_RETRIES attempts${NC}"
        exit 1
    fi
    echo -e "${YELLOW}Waiting for PostgreSQL... ($RETRY_COUNT/$MAX_RETRIES)${NC}"
    sleep 2
done
echo -e "${GREEN}✓ PostgreSQL is ready${NC}\n"

# Ensure required extensions exist
echo -e "${YELLOW}Ensuring required extensions are available...${NC}"
docker exec -i "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
echo -e "${GREEN}✓ Extensions verified${NC}\n"

# Create schema_migrations table to track applied migrations
echo -e "${YELLOW}Setting up migration tracking...${NC}"
docker exec -i "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 <<EOF
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT NOW()
);
EOF
echo -e "${GREEN}✓ Migration tracking ready${NC}\n"

# Function to check if migration has been applied
is_migration_applied() {
    local version="$1"
    local result
    result=$(docker exec "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version = '$version';" | tr -d '[:space:]')
    [[ "$result" == "1" ]]
}

# Function to mark migration as applied
mark_migration_applied() {
    local version="$1"
    docker exec "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "INSERT INTO schema_migrations (version) VALUES ('$version') ON CONFLICT DO NOTHING;" >/dev/null
}

echo -e "${YELLOW}Applying SQL migrations...${NC}\n"

apply_migration() {
    local file_path="$1"
    local file_name
    file_name="$(basename "$file_path")"
    local version="${file_name%.sql}"

    # Check if already applied
    if is_migration_applied "$version"; then
        echo -e "${BLUE}→ Skipping ${file_name} (already applied)${NC}"
        return 0
    fi

    echo -e "${BLUE}→ Applying ${file_name}...${NC}"
    if cat "$file_path" | docker exec -i "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 >/dev/null; then
        mark_migration_applied "$version"
        echo -e "${GREEN}   ✓ ${file_name} applied${NC}\n"
    else
        echo -e "${RED}   ✗ Failed to apply ${file_name}${NC}"
        exit 1
    fi
}

declare -a ORDERED_MIGRATIONS=(
    "001_initial_schema.sql"
    "002_create_auth_system.sql"
    "003_create_business_tables.sql"
    "004_create_order_items_table.sql"
    "005_fix_product_permissions.sql"
    "006_fix_user_tenant_constraints.sql"
    "007_add_password_hash_to_users.sql"
    "008_fix_email_uniqueness_constraint.sql"
    "009_add_invoice_fields_and_sequence.sql"
    "010_add_category_hierarchy.sql"
    "011_insert_sample_categories.sql"
    "012_create_analytics_tables.sql"
    "013_create_notification_system.sql"
    "014_enhance_audit_logs_table.sql"
    "015_create_product_images_table.sql"
    "016_add_missing_rbac_permissions.sql"
    "017_create_order_status_history.sql"
    "018_add_status_columns.sql"
    "019_add_tenant_contact_info.sql"
    "020_add_user_roles_columns.sql"
    "021_optimize_search_indexes.sql"
    "022_add_core_performance_indexes.sql"
    "023_add_comprehensive_indexes.sql"
)

for filename in "${ORDERED_MIGRATIONS[@]}"; do
    migration_path="$MIGRATIONS_DIR/$filename"
    if [ -f "$migration_path" ]; then
        apply_migration "$migration_path"
    fi
done

# Apply any remaining migrations not in the ordered list (sorted by filename)
while IFS= read -r -d '' migration_file; do
    basename_file="$(basename "$migration_file")"
    found=false
    for ordered in "${ORDERED_MIGRATIONS[@]}"; do
        if [[ "$basename_file" == "$ordered" ]]; then
            found=true
            break
        fi
    done
    if ! $found; then
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
