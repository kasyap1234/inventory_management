# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

Agromart is a multi-tenant inventory management system for AgroTech companies. It consists of:
- **Backend**: Go with Echo framework, PostgreSQL, Redis, MinIO, Razorpay integration
- **Frontend**: Next.js 15 with React 19, TanStack Query, Tailwind CSS, Playwright tests
- **Infrastructure**: Docker Compose (PostgreSQL, Redis, MinIO, Mailpit), Task automation

## Architecture Summary

### Backend Layers (Go)
```
internal/
├── handlers/       # HTTP endpoints - extract context, call services, format responses
├── services/       # Business logic - orchestration, domain rules, validations
├── repositories/   # Data access - pgx SQL queries with tenant_id filtering
├── middleware/     # JWT auth, RBAC (role-based access control), rate limiting
├── models/         # Domain models
├── common/         # Context utils, error helpers, structured logging
├── config/         # Configuration loading
├── caching/        # Redis cache service
├── validation/     # Input validation with go-playground/validator
├── jobs/background # Background job handlers (asynq + Redis)
└── analytics/      # Analytics and reporting
```

### Frontend Structure (Next.js)
```
frontend/
├── app/            # App Router - pages, layouts, auth flows
├── components/     # Reusable UI components with Tailwind + shadcn/ui
├── hooks/          # TanStack Query hooks for data fetching
├── lib/            # Utilities (API client, types, helpers)
├── types/          # TypeScript type definitions
└── tests/          # Playwright E2E + Vitest unit tests
```

## Critical Patterns

### Multi-Tenancy (REQUIRED)
Every data operation must be tenant-scoped. This is non-negotiable.

**Handler**: Extract tenant from JWT-populated context:
```go
tenantID, ok := common.GetTenantIDFromContext(ctx)
if !ok {
    return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
}
```

**Repository**: Always include `tenant_id` in WHERE clauses:
```go
query := `SELECT * FROM products WHERE tenant_id = $1 AND id = $2`
```

### RBAC Permissions
Defined in `config/role_templates.yaml`, supports logical operators:
- `resource.action` - exact permission
- `perm1||perm2` - OR condition
- `perm1&&perm2` - AND condition

Register routes with RBAC in `cmd/main.go`:
```go
protected.GET("/products", handler, rbacMiddleware.RequirePermission("product.read"))
```

### Error Responses
Use helpers from `internal/common`:
```go
common.SendValidationError(c, "email", "Invalid format")
common.SendNotFoundError(c, "Product")
common.SendForbiddenError(c, "Insufficient permissions")
```

### Repository Pattern
- Interface + implementation in same file: `internal/repositories/product_repo.go`
- Constructor returns interface: `func NewProductRepo(db *pgxpool.Pool) ProductRepository`
- Use parameterized queries (`$1, $2`) for SQL injection prevention

### Frontend Data Fetching
- API client in `frontend/lib/api.ts` auto-attaches JWT (HttpOnly cookie) + CSRF token
- TanStack Query hooks in `frontend/hooks/` for caching and data management
- Components use Tailwind CSS with shadcn/ui patterns

## Common Development Commands

### Full Stack
```bash
task setup              # Initial setup: install deps, start DB, run migrations
task run                # Start DB, backend :8080, frontend :3000
task dev                # Alias for task run
task stop               # Stop all services
```

### Database
```bash
task db:start           # Start PostgreSQL, Redis, MinIO, Mailpit
task db:migrate         # Run SQL migrations
task db:psql            # Open PostgreSQL interactive shell
task db:reset           # Destructive: wipe all data and volumes
```

### Backend (Go)
```bash
task backend:run        # Build and run backend
task backend:dev        # Run with hot reload (requires air: go install github.com/air-verse/air@latest)
task backend:build      # Build binary only
task lint:backend       # go vet + staticcheck
task fmt:backend        # gofmt + goimports
task test:backend       # Run all Go tests
task test:backend:short # Quick tests (short flag)
task test:backend:coverage  # Generate coverage report
```

### Frontend (Next.js)
```bash
task frontend:run       # Start dev server :3000
task frontend:build     # Build for production
task frontend:start     # Run production build
task lint:frontend      # ESLint check
task fmt:frontend       # Format with Prettier
task test:frontend      # Run Playwright + Vitest
task test:frontend:e2e  # Playwright E2E only
task test:frontend:unit # Vitest unit tests only
task test:frontend:watch # Vitest watch mode
task test:frontend:ui   # Playwright UI mode (interactive debugging)
```

### Run Single Test
- **Go**: `go test ./internal/services/product_service_test.go -v -run TestProductCreation`
- **Frontend E2E**: `cd frontend && bun run test:e2e -- tests/e2e/auth.spec.ts`
- **Frontend Unit**: `cd frontend && bun run test:unit -- --testNamePattern="Component renders"`

### CI Pipeline
```bash
task ci                 # deps:all → lint:all → test:all → build:all
task check              # lint:all + test:all
```

## Key Files and Locations

| Purpose | Location |
|---------|----------|
| Route registration & app initialization | `cmd/main.go` |
| Context utilities & tenant extraction | `internal/common/context_utils.go` |
| RBAC middleware enforcement | `internal/middleware/rbac.go` |
| Role & permission templates | `config/role_templates.yaml` |
| Frontend API client | `frontend/lib/api.ts` |
| Database schema | `migrations/*.sql` (numbered sequence) |
| Environment config | `.env` (copy from `.env.example`) |
| Docker services | `docker-compose.yml` |
| Task definitions | `Taskfile.yml` |

## Adding New Features

### New API Endpoint
1. Add handler method in `internal/handlers/{resource}_handlers.go`
2. Add service method in `internal/services/{resource}_service.go`
3. Add repository method in `internal/repositories/{resource}_repo.go` (with `tenant_id` filter)
4. Register route in `cmd/main.go` with RBAC middleware
5. Add tests in `*_test.go` files

### New Permission
Add to `config/role_templates.yaml` under appropriate roles.

### New Background Job
1. Define job type constant in `internal/jobs/`
2. Create handler in `internal/jobs/background/`
3. Register in `cmd/main.go` via asynq task mux
4. Enqueue from service using asynq client

### New Frontend Page
1. Create page structure in `frontend/app/{feature}/`
2. Add data-fetching hook in `frontend/hooks/`
3. Create UI components in `frontend/components/`
4. Add E2E tests in `frontend/tests/e2e/`

## Database & Migrations

Database is PostgreSQL 15+. Migrations are SQL files in `migrations/` with numbered prefixes (e.g., `001_init_schema.sql`).

Run migrations:
```bash
task db:migrate         # Via script
./run_migrations.sh     # Manual
```

All tables must include `tenant_id` foreign key to `tenants(id)` for multi-tenancy.

## Testing

### Backend Testing
- Use `testhelpers/` for common test utilities
- Mock repositories and external services
- Tests use in-memory PostgreSQL or test containers
- Run specific test: `go test ./path -run TestName -v`

### Frontend Testing
- **E2E**: Playwright in `frontend/tests/e2e/`, `frontend/tests/ui/`, `frontend/tests/integration/`
- **Unit**: Vitest in component directories
- **Browsers**: Chromium, Firefox, WebKit, Mobile Chrome/Safari
- Run UI mode for debugging: `task test:frontend:ui`
- Install browsers: `task test:frontend:install`

## Environment Configuration

Key variables in `.env`:
- `DATABASE_URL` - PostgreSQL connection string (required)
- `JWT_SECRET` - Auth token secret (required in production)
- `CSRF_SECRET` - CSRF protection (required in production)
- `ENFORCE_HTTPS` - Must be "true" in production
- `GO_ENV` - "production" or "development"
- `MINIO_*` - MinIO S3-compatible storage config
- `RAZORPAY_*` - Payment gateway credentials
- `RESEND_API_KEY` - Email service (optional)

MinIO buckets `product-images` and `invoices` are auto-created at startup.

## Special Features

### Image Uploads
Presigned URL flow:
1. `POST /products/:id/images/presign` - get presigned upload URL
2. Upload directly to MinIO
3. `POST /products/:id/images/finalize` - thumbnails generated, metadata persisted

### Payments
- Razorpay integration with webhook support
- One-time orders: `POST /payments/orders` → `POST /payments/verify`
- Subscriptions: `POST /subscriptions` with lifecycle webhooks
- Config endpoint: `GET /payments/config` (public key only)

### Notifications
Multiple channels: Email (Resend/SMTP), SMS (Twilio), Push (WebSocket-based device tokens)

### Analytics
Dashboard metrics: sales, stock, orders. Uses caching layer with Redis for performance.

## Important Codebase Rules

From `.github/copilot-instructions.md`:
1. **Tenant-scoped operations are non-negotiable** - every query must filter by `tenant_id`
2. **Use structured logging** via `internal/common` logger
3. **Validate inputs** with go-playground/validator (backend) and Zod/schema (frontend)
4. **Handle errors gracefully** with proper HTTP status codes and error responses
5. **Test critical paths** - auth, payments, inventory operations require test coverage
6. **Follow Go conventions** - interfaces named with `-er` suffix, constructors return interfaces
7. **Frontend uses HttpOnly cookies** for JWT storage (XSRF-protected with custom header)
