# Database Migrations

This directory contains all SQL migrations for the Agromart Inventory Management system.

## Quick Start

### Run Migrations
```bash
cd /path/to/inventory_management
./run_migrations.sh
```

The script will:
1. Verify PostgreSQL container is running
2. Create required extensions (pgcrypto)
3. Apply all 21 migrations in the correct order
4. Display final status and default credentials

### Verify Migrations
```bash
docker exec inventory_management-postgres-1 psql -U testuser -d testdb -c "\dt"
```

## Migration Order

Migrations are applied in a specific order to handle dependencies:

1. **Base Schemas** (authentication & core schema)
   - `complete_auth_schema.sql` - Users, roles, permissions
   - `schema.sql` - Tenants and products

2. **Business Tables** (core entities)
   - `20250831110324_create_business_tables_fixed.sql` - All business entities

3. **Fixes & Constraints** (data integrity)
   - `20240831_fix_product_permissions.sql`
   - `20250831170400_fix_user_tenant_constraints.sql`
   - `20250201120000_add_password_hash_to_users.sql`
   - `20250831180000_fix_email_uniqueness_constraint.sql`

4. **Business Logic** (features)
   - `20250831170500_add_invoice_fields_and_sequence.sql`
   - `20250831171700_add_category_hierarchy.sql`
   - `20250831171701_insert_sample_categories.sql`

5. **Analytics & Advanced Features**
   - `20250901120000_create_analytics_tables.sql`
   - `20250114115959_create_order_items_table.sql`
   - `20250114120000_create_enhanced_notification_system.sql`
   - `20250114120002_enhance_audit_logs_table.sql`
   - `20250114120003_create_enhanced_product_images.sql`
   - `20250114120004_add_missing_rbac_permissions.sql`
   - `20250114120005_create_order_status_history.sql`
   - `20250113235959_add_status_to_suppliers_distributors_warehouses.sql`
   - `20250115120000_add_tenant_contact_info.sql`

6. **Performance Optimization** (indexes - final)
   - `20240130000001_optimize_search_indexes.sql` - Search-specific indexes
   - `20251009120000_add_performance_indexes.sql` - Core performance indexes

## Key Tables

- **tenants** - Multi-tenancy support
- **users, roles, permissions** - Authentication & authorization
- **products, categories** - Product catalog
- **warehouses, inventory** - Stock management
- **suppliers, distributors** - Business partners
- **orders, order_items** - Order management
- **invoices** - Billing & GST tracking
- **audit_logs** - Compliance & debugging
- **notifications** - User notifications
- **subscriptions** - Billing management

## Idempotency

All migrations use `CREATE TABLE IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS` statements, making them safe to re-run. Duplicate index creation attempts are harmless and will be skipped.

## Troubleshooting

### PostgreSQL Connection Issues
```bash
# Check if container is running
docker ps | grep postgres

# Start container if needed
docker-compose up -d postgres
```

### Migration Failures
Check the error message and ensure:
1. PostgreSQL is running and healthy
2. Database exists: `testdb`
3. User exists: `testuser`
4. All dependencies (referenced tables) exist before the referencing migration

### Verify Default Data
```sql
-- Check default tenant
SELECT * FROM tenants WHERE subdomain = 'agromart-dev';

-- Check default roles
SELECT * FROM roles WHERE tenant_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
```

## Adding New Migrations

1. Create a new file with naming convention: `YYYYMMDDHHMMSS_description.sql`
2. Use only `CREATE TABLE IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS`
3. Include comments explaining the purpose
4. Add the filename to the `ORDERED_MIGRATIONS` array in `run_migrations.sh`
5. Test by running `./run_migrations.sh`

## Recent Cleanup (2025-01-16)

- ✅ Removed 3 problematic migration files (duplicates, bad timestamps, mixed frameworks)
- ✅ Consolidated 50+ duplicate index creation statements
- ✅ Cleaned migration runner script (no more skip lists)
- ✅ Standardized all migrations to pure SQL format
- ✅ Current status: **CLEAN & PRODUCTION READY**

See `MIGRATION_CLEANUP_REPORT.md` for detailed cleanup history.
