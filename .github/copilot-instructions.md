# Agromart Codebase Instructions

## Architecture (Multi-Tenant Inventory System)

**Backend**: Go + Echo framework (`/internal`) | **Frontend**: Next.js 15 + TanStack Query (`/frontend`)

**Critical**: Every data operation MUST be tenant-scoped. All models have `tenant_id` and all queries must filter by it.

```
internal/
├── handlers/      # HTTP layer - extract context, call services, format responses
├── services/      # Business logic - orchestrate repos, enforce domain rules
├── repositories/  # Data access - raw pgx SQL queries, always include tenant_id
├── middleware/    # JWT auth, RBAC, rate limiting
└── common/        # Context utilities, error helpers, logging
```

## Essential Patterns

### Multi-Tenancy (Required for ALL data operations)
```go
// Handler: extract tenant from JWT-populated context
tenantID, ok := common.GetTenantIDFromContext(ctx)
if !ok {
    return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
}

// Repository: ALWAYS include tenant_id in WHERE clause
query := `SELECT * FROM products WHERE tenant_id = $1 AND id = $2`
```

### RBAC Permissions (`config/role_templates.yaml`)
```go
// Format: resource.action - supports || (OR) and && (AND)
rbacMiddleware.RequirePermission("product.create")
rbacMiddleware.RequirePermission("tenant.read||system.admin")  // Either works
rbacMiddleware.RequirePermission("inventory.read&&audit.read") // Both required
```

### Error Responses (use `internal/common` helpers)
```go
return common.SendValidationError(c, "email", "Invalid format")
return common.SendNotFoundError(c, "Product")
```

### Repository Pattern
- Interface + implementation in same file: `internal/repositories/product_repo.go`
- Constructor returns interface: `func NewProductRepo(db *pgxpool.Pool) ProductRepository`
- Use `$1, $2` parameterized SQL; always filter by `tenant_id`

## Developer Commands

```bash
# Full stack (recommended)
task run                    # Starts DB, backend :8080, frontend :3000

# Individual services
task db:start && task db:migrate   # Infrastructure + migrations
task backend:run                   # Go backend only
task frontend:run                  # Next.js frontend only

# Database
task db:psql                # Open PostgreSQL shell
task db:reset               # DESTRUCTIVE: Wipe all data

# Testing
go test ./internal/services/... -v  # Backend tests
cd frontend && bun run test:e2e     # Playwright E2E
```

## Adding New Features

**New API endpoint**: 
1. Add handler method in `internal/handlers/{resource}_handlers.go`
2. Add service method in `internal/services/{resource}_service.go`
3. Add repo method in `internal/repositories/{resource}_repo.go` (with tenant_id filter)
4. Register route in `cmd/main.go` with RBAC: `protected.GET("/path", handler, rbacMiddleware.RequirePermission("resource.action"))`

**New permission**: Add to `config/role_templates.yaml` under appropriate roles

**New background job**: Register in `cmd/main.go` via `mux.HandleFunc(jobs.TypeNewJob, handler)` and create handler in `internal/jobs/`

## Frontend Conventions

- API client: `frontend/lib/api.ts` - auto-attaches JWT (HttpOnly cookie) + CSRF token
- Data fetching: TanStack Query hooks in `frontend/hooks/`
- Components: `frontend/components/` with Tailwind + shadcn/ui patterns

## Key Files

| Purpose | Location |
|---------|----------|
| Route registration | `cmd/main.go` (lines 600-800) |
| Context utilities | `internal/common/context_utils.go` |
| RBAC middleware | `internal/middleware/rbac.go` |
| Role definitions | `config/role_templates.yaml` |
| API client | `frontend/lib/api.ts` |
| DB migrations | `migrations/*.sql` (numbered sequence) |
