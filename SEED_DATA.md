# Seed Data & Default Credentials

This document contains the default credentials and seed data created during database migration.

## Super Admin Account

| Field    | Value                        |
|----------|------------------------------|
| Email    | superadmin@agromart.com      |
| Password | SuperSecretPass123!          |
| Role     | Platform Admin (Super Admin) |

> **Note:** The super admin password is configured via the `SUPER_ADMIN_PASSWORD` environment variable in `.env`.

---

## Default Tenant

| Field     | Value                                  |
|-----------|----------------------------------------|
| ID        | aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa   |
| Name      | Agromart Development                   |
| Subdomain | agromart-dev                           |

---

## Default Roles

| Role  | ID                                     |
|-------|----------------------------------------|
| Admin | bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb   |
| User  | cccccccc-cccc-cccc-cccc-cccccccccccc   |

---

## Acme Sample Tenant (Optional Demo Data)

The migration `041_seed_acme_sample_data.sql` seeds comprehensive demo data if the Acme tenant exists.

### Acme Tenant Credentials

| Field     | Value                                  |
|-----------|----------------------------------------|
| Email     | acme@gmail.com                         |
| Password  | AcmeDemo123!                           |
| Tenant ID | ff5c2ea7-4d2d-4928-a34f-faac6be42c71   |
| User ID   | 812ae9d8-6a2d-4382-a0ca-bacf044613d6   |

> **Note:** To use the Acme sample data, you must first create a tenant with the above ID by signing up with the email `acme@gmail.com`. After signup, run the migrations again to populate the demo data.

### Acme Sample Data Includes

- **Warehouses:** Main Warehouse (Mumbai), East Warehouse (Pune)
- **Categories:** Fertilizers, Irrigation, Seeds
- **Products:** NPK Fertilizer, Drip Irrigation Kit, Hybrid Basmati Rice Seeds
- **Suppliers:** Agri Supplies Ltd, Green Harvest Co
- **Distributors:** Farm Distribution Co, Coastal Agro Partners
- **Orders:** 2 Purchase Orders, 3 Sales Orders
- **Invoices:** Sample paid and overdue invoices
- **Inventory:** Stock levels with reservations
- **Notifications:** Sample low stock and shipment alerts

---

## Environment Configuration

Default credentials are configured in `.env` (copy from `.env.example`):

```bash
# Database
DATABASE_URL=postgresql://testuser:testpass@localhost:5440/testdb?sslmode=disable
POSTGRES_PASSWORD=testpass

# Super Admin
SUPER_ADMIN_EMAIL=superadmin@agromart.com
SUPER_ADMIN_PASSWORD=SuperSecretPass123!

# JWT
JWT_SECRET=your-256-bit-secret-key-change-in-production
```

---

## Quick Start

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Start the development server:
   ```bash
   task dev
   ```

3. Login with super admin credentials:
   - **Email:** superadmin@agromart.com
   - **Password:** SuperSecretPass123!
