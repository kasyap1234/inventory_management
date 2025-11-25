# Agromart Codebase Instructions

## Architecture Overview

This is a **multi-tenant inventory management system** with a Go backend (Echo framework) and Next.js 15 frontend. All data operations are **tenant-scoped** - every model includes `tenant_id` and all queries must filter by it.

### Backend Structure (`/internal`)
```
handlers/     → HTTP handlers (request parsing, response formatting)
services/     → Business logic (orchestrates repos, implements domain rules)
repositories/ → Data access (direct pgx queries, no ORM)
models/       → Data structures with JSON/DB tags
middleware/   → JWT auth, RBAC, rate limiting, audit logging
common/       → Shared utilities (context, validation, logging)
```

### Key Data Flow
1. Request → JWT middleware extracts `user_id` + `tenant_id` into context
2. Handler extracts IDs via `common.GetTenantIDFromContext(ctx)`, `common.GetUserIDFromContext(ctx)`
3. Service performs business logic with tenant scoping
4. Repository executes tenant-scoped SQL queries

## Critical Patterns

### Multi-Tenancy (REQUIRED for all data operations)
```go
// In handlers - always extract tenant from context
tenantID, ok := common.GetTenantIDFromContext(ctx)
if !ok {
    return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
}

// In repositories - always include tenant_id in queries
query := `SELECT * FROM products WHERE tenant_id = $1 AND id = $2`
```

### RBAC Permissions
Permissions follow `resource.action` pattern (e.g., `product.create`, `inventory.update`).
```go
// Single permission
protected.GET("/products", handler, rbacMiddleware.RequirePermission("product.list"))

// OR logic (either permission works)
rbacMiddleware.RequirePermission("tenant.read||system.admin")

// AND logic (both required)
rbacMiddleware.RequirePermission("inventory.read&&audit.read")
```

Role templates defined in `config/role_templates.yaml`. Wildcard `*` grants all permissions.

### Repository Pattern
- Interfaces live next to implementations (`internal/repositories/*.go`). Each repo exposes an interface (e.g., `UserRepository`) and a private struct that embeds `*pgxpool.Pool` and implements the methods.
- Constructors return the interface type (`func NewUserRepo(db *pgxpool.Pool) UserRepository`) so services only depend on the abstraction.
- Use parameterized SQL with `$1, $2` and always include `tenant_id` in `WHERE` clauses; any missing tenant filters will be immediately flagged by schema constraints.
- Prefer helper methods under `common/transaction_manager.go` or `common/db_optimization.go` when executing multi-statement transactions or materialized-view refreshes.

### Error Handling
Use structured error responses from `internal/common`:
```go
return common.SendValidationError(c, "email", "Invalid email format")
return common.SendNotFoundError(c, "Product")
return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{...})
```

## Developer Workflows

### Local Setup
```bash
docker-compose up -d          # Start Postgres (5440), Redis (6379), MinIO (9000), MailHog (8025)
./run_migrations.sh           # Apply SQL migrations
go run cmd/main.go            # Backend on :8080
cd frontend && bun dev        # Frontend on :3000
```

### Database
- Migrations in `/migrations/*.sql` (numbered sequence)
- Connection: `postgresql://testuser:testpass@localhost:5440/testdb`
- Use pgAdmin at localhost:5050 for debugging

### Testing
```bash
# Backend
go test ./...
go test ./internal/services/... -v

# Frontend (Playwright)
cd frontend
bun run test:e2e              # Full E2E suite
bun run test:auth             # Auth-specific tests
```

## Frontend Conventions

- **Next.js 15** with App Router, Turbopack, React 19
- **TanStack Query** for data fetching (`@tanstack/react-query`)
- API client in `frontend/lib/api.ts` - auto-attaches JWT + CSRF tokens
- Components in `frontend/components/`, hooks in `frontend/hooks/`
- Tailwind CSS with dark mode support (`next-themes`)

## Background Jobs & Scheduling

- Asynq handles the queueing machinery (see `cmd/main.go` for client/server/mux setup and `internal/jobs/` for task handlers). Register new jobs via `mux.HandleFunc(...)` and give them a `jobs.Type*` string to keep naming consistent.
- Repeated work (materialized view refresh, vacuum/analyze, notification retries) are scheduled as goroutines in `cmd/main.go`—keep context timeouts fresh and cancel when exiting to avoid leaks.
- Tally exporters/importers depend on `internal/jobs/tally.go` plus `services` for orchestrating DB snapshots; reuse those services when adding accounting-sync jobs.

## Notifications & Alerting Patterns

- Notification delivery is mediated through `services.NotificationDeliveryService`, which composes template, webhook, alert rule, and Firebase hooks (`services/notification_delivery_service.go`). Use it when orchestrating email/SMS/push flows.
- Templates live in `repositories/NewNotificationTemplateRepo` and `handlers.NewNotificationTemplateHandlers`; alert rules and webhook subscriptions use `handlers/alert_rule_handlers.go` and `handlers/webhook_subscription_handlers.go`.
- Device tokens and deliveries are recorded in `notification_delivery_repo.go` and `device_token_repo.go`; search the `notifications` feature flag logic in `cmd/main.go` to see how Firebase/FCM toggles are applied.
- When retrying failed deliveries, `deliverySvc.RetryFailedDeliveries` (called from the goroutine in `cmd/main.go`) expects tenant-scoped results and enforces exponential backoff/jitter—adhere to this pattern when introducing new delivery guarantees.

## Subscription & Billing Hooks

- Razorpay keys and webhook secrets are pulled from env vars in `cmd/main.go` before constructing `services.NewRazorpayService`. Hooked subscriptions (create/update events) funnel through `services.NewSubscriptionService` and `handlers.NewSubscriptionHandlers`.
- Subscription enforcement runs via `services.SubscriptionMiddlewareService`; register it on handlers to gate feature flags (`services.NewWarehouseService` uses it as an example).
- Razorpay webhook handling and validation happen through `handlers.WebhookHandlers`; tag new webhook endpoints with the same RBAC/CSRF considerations as the existing `/v1/webhooks` route.

## Configuration & Secrets

- Copy `.env.example` to `.env` before running locally; keys like `DATABASE_URL`, `JWT_SECRET`, `MINIO_*`, `REDIS_*`, `RAZORPAY_*` are required for most features.
- `config/tally.toml` drives Asynq queue settings + Tally import/export behavior; change queue priorities or concurrency there rather than in code.
- Production deployments expect `ENFORCE_HTTPS=true`, `CSRF_SECRET`, and secure JWT/minio/redis credentials; missing secrets trigger fatal logs in `cmd/main.go`.

## Caching, Rate Limiting & Sessions

- `caching.NewRedisCacheService` (internal/caching/cache_service.go) is the centralized Redis client for product/inventory/category caches, RBAC token caching, session storage, and rate-limit counters.
- RBAC caching uses keys like `agromart:rbac:permission:<tenant>:<user>:<perm>` to avoid repeated SQL; `cacheSvc` is wired into `services.NewRBACServiceWithCache`.
- Rate limiting middleware (`internal/middleware/rate_limit.go`) calls `cacheService.IsRateLimited`/`IncrementRateLimit` to honor the per-resource counters defined in `cmd/main.go` (10 req/s, burst 20).
- Sessions and CSRF tokens are stored via the same cache service which makes `common.NewStructuredLogger()` log cache connection attempts during startup if Redis is unreachable.

## Middleware & Security Stack

- JWT validation happens in `internal/middleware/jwt.go`; every protected route is wrapped by `middleware.JWTMiddleware(userRepo, jwtSecret)` before RBAC checks. The middleware also populates the request context with user/tenant IDs for downstream services.
- RBAC middleware (`internal/middleware/rbac.go`) parses permission strings with `||`/`&&` so handlers can compose complex rules without duplicating logic; cached checks use `services.NewRBACServiceWithCache` and Redis keys like `agromart:rbac:permission:<tenant>:<user>:<perm>`.
- Rate limiting is enforced via `internal/middleware/rate_limit.go` (10 req/s, burst 20) plus the `caching.CacheService` to throttle by client IP; use `RateLimiter.Middleware()` early in the pipeline to protect heavy routes.
- Performance middleware (`internal/middleware/performance.go`) bundles gzip, request timeout (30s), and body limit (10M) so you rarely need to add them manually.
- CORS is configured for local/staging domains in `cmd/main.go`, along with `SecurityHeaders`, `RequestID`, and CSRF protection (`internal/middleware/csrf.go`) that skips `/v1/security/csrf` and `/v1/webhooks/razorpay`.
- Health/metrics/docs routes (`/health*`, `/metrics`, `/docs/*`) bypass auth for monitoring; their handlers live under `internal/handlers/health_checks.go`, `handlers/metrics_handler.go`, and `handlers` directory.

## Observability & Events

- `common.NewStructuredLogger()` yields a zerolog-based logger across services; inspect `internal/common/structured_logger.go` for helper methods like `ErrorWithContext`.
- Critical errors/events register via `common.RegisterEventHandler` in `cmd/main.go`; handlers like `deliverySvc.ProcessEvent` and `analyticsSvc.RecordSearchUsage` expect tenant context to be present.
- Background jobs use Asynq's inspector and client (redis-backed) and start the server in a goroutine while ensuring graceful shutdown in `cmd/main.go`.

## Service & Domain Overview

- **Auth & User Management**: `services/auth_service.go` issues JWTs/refresh tokens, enforces tenant scoping (`repositories/user_repo.go` + `user_role_repo.go`), and drives `handlers/auth_handlers.go`. Email verification, 2FA, and Google auth all funnel through this layer; handlers rely on `common.SendValidationError` for structured mistakes and `common.NewStructuredLogger()` for audit-friendly logs.
- **Inventory & Product**: `services/product_service.go` and `services/inventory_service.go` orchestrate product snapshots, stock adjustments, and batch creation via `repositories/product_repo.go`, `repositories/inventory_repo.go`, and `repositories/batch_repo.go`. Product images upload through `services/minio_service.go` + `repositories/product_image_repo.go` and are surfaced via `handlers/product_handlers.go` (bulk create/price update paths also call `ProductService`).
- **Order, Invoice & Analytics**: `services/order_service.go` handles stock reservation (calls `inventoryService.AdjustStock`). `services/invoice_service.go` ties invoicing to analytics (`internal/services/analytics_service.go`) and also coordinates with `services/supplier_service.go`/`distributor_service.go` for linked entities. Business rules (stock checks, pricing) live in services—handlers should only orchestrate payloads. Tally export/import jobs reuse `internal/jobs/tally.go` and share the same services.
- **Notifications & Alerts**: `services.NotificationDeliveryService` composes template, webhook, alert rule, and Firebase hooks. Template metadata lives in `repositories/notification_template_repo.go`, while alert rules and webhook subscriptions live under `handlers/alert_rule_handlers.go` + `handlers/webhook_subscription_handlers.go`. Notification handlers route to `services.notification_service.go` (email/SMS) and `services.notification_delivery_service.go` (queued deliveries).
- **Security & Audit**: `middleware/rbac.go`, `services/rbac_service.go`, and `handlers/audit_logs_handlers.go` enforce permissions and audit writes (see `internal/handlers/audit_logs_handlers.go`). `handlers/security_handlers.go` exposes CSRF tokens; `internal/middleware/websocket` (if used) should reuse the same RBAC guard. Structured logging in `internal/common/structured_logger.go` surfaces context (tenant, user, request ID).
- **Billing & Subscription**: Razorpay integration lives in `services/razorpay_service.go`; `services/subscription_service.go` and `handlers/subscription_handlers.go` expose payment flows. Premium gating uses `services.subscription_middleware_service.go` (e.g., `services.NewWarehouseService` passes it to handler constructors). Webhooks hit `handlers/webhook_handlers.go` and validate via the shared secret.
- **Caching & Rate Limiting**: `internal/caching/cache_service.go` provides product/inventory/category caching, RBAC token caching, session storage, and rate-limit counters. Invalidations happen after writes in services (`product_service.InvalidateCacheAfterUpdate`). Rate limiting middleware (`internal/middleware/rate_limit.go`) calls `cacheService.IsRateLimited`/`IncrementRateLimit` with keys like `agromart:limit:<tenant>:<path>`.

## Subsystem Deep Dive

- **Tenant Bootstrapping**: `repositories/tenant_repo.go` + `services/tenant_service.go` manage subdomain creation, status transitions, and `tenant_settings`. Handlers (`handlers/tenant_handlers.go`) wrap these with RBAC and emit audit logs.
- **Role Management**: `services/role_management_service.go` + `repositories/role_repo.go` orchestrate template-based roles from `config/role_templates.yaml`. Permission lookups use `repositories/permission_repo.go`/`role_permission_repo.go`, and handlers under `role_handlers.go` enforce `tenant_id` scope.
- **Audit & Activity**: `internal/logging/audit_logger.go` is wired into `services/audit_logs_service.go`; repository writes go through `internal/repositories/audit_logs_repo.go`. Handlers append extra metadata (user, tenant, resource) and reference `common.CreateErrorResponse` for uniform errors.
- **Metrics & Health**: `handlers/health_checks.go` + `handlers/metrics_handler.go` expose `/health` and `/metrics`. Health checks skip auth and can be used locally (`curl http://localhost:8080/health`).

## Key SQL Files

- `migrations/001_initial_schema.sql` sets up tenants, categories, products, and foundational constraints.
- `migrations/009_add_invoice_fields_and_sequence.sql` introduces invoice sequencing needed by `services/invoice_service.go`.
- `migrations/012_create_analytics_tables.sql` and `013_create_notification_system.sql` add tables referenced by `services/analytics_service.go` and notification repos respectively.
- `migrations/015_create_product_images_table.sql` backs the MinIO-backed image flows in `services/minio_service.go`.

## Frontend Component Structure

- The frontend App Router lives under `frontend/app/`; `page.tsx` is the landing surface, while directories like `auth/`, `dashboard/`, and `inventory/` map to routes.
- Shared state & data fetching live in `frontend/hooks/`, e.g., `useInventory.ts` wraps TanStack Query queries defined in `frontend/lib/api.ts`.
- Reusable UI primitives live in `frontend/components/` (cards, forms, tables); look for Tailwind-based styling and `class-variance-authority` for variants.
- Testing harnesses Playwright suites in `frontend/tests/` (smoke, e2e, ui) with environment-aware fixtures; configure `NEXT_PUBLIC_API_URL` and `BASE_URL` via `.env.test` if needed.
- API layer (`frontend/lib/api.ts`) auto-attaches JWT + CSRF tokens, falls back to `/api/v1`, and treats 4xx responses as errors so handlers can inspect `response.data` when catching.

## Testing & QA Hooks

- Backend tests live alongside services/handlers (search `*_test.go` under `internal/`). Most services already have tests (e.g., `inventory_service_test.go`). Run `go test ./...` before pushing and target specific packages when iterating.
- Playwright powers frontend testing (`frontend/tests/*`). `playwright.config.ts` defines multiple projects (chromium/firefox/webkit + mobile viewports), retries on CI, and shared settings such as trace-on-retry, screenshots/videos on failures, and a built-in `webServer` callback that runs `bun run dev`. Always run `bun run test:install` after altering Playwright dependencies.
- Keep the `frontend/tests/ui/` suite for visual regressions; smoke/integration/e2e suites live under `frontend/tests/` with descriptive names (e.g., `tests/e2e/auth.spec.ts`). Each test imports helpers from `frontend/tests/helpers` to stub API responses.
- Integration helpers under `testhelpers/` and shell scripts like `test_end_to_end_workflow_complete.sh` automate common verification flows—update them when database schema or flows change.

## Tooling & Scripts

- Infrastructure helpers (`run_migrations.sh`, `run_start.sh`, `run_stop.sh`, `deploy_optimizations.sh`) wrap docker-compose + Go commands—update them only when the runtime contract changes.
- Envelope QA is supported by scripts like `test_minio_functionality.sh` and `test_output.txt`: keep them synced with schema/database changes.
- Frontend uses Bun; install dependencies with `bun install` (mirrors `package.json`). Use `bun dev` for local runs, `bun run lint` for formatting, and `bun run test:install` if Playwright deps change.

## Key Files Reference

| Purpose | Location |
|---------|----------|
| Route registration | `cmd/main.go` lines 600-800 |
| Auth flow | `internal/handlers/auth_handlers.go` |
| RBAC middleware | `internal/middleware/rbac.go` |
| Context utilities | `internal/common/context_utils.go` |
| Notification delivery | `internal/services/notification_delivery_service.go` |
| Tally job handlers | `internal/jobs/tally.go` |
| Subscription + Razorpay | `internal/services/subscription_service.go` + `handlers/webhook_handlers.go` |
| Role definitions | `config/role_templates.yaml` |
| API client | `frontend/lib/api.ts` |
| DB schema | `migrations/001_initial_schema.sql` |
