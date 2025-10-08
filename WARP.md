# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

Agromart2 is a multi-tenant SaaS inventory management platform designed for agricultural businesses. It provides comprehensive inventory, order, invoice, and product management capabilities with role-based access control (RBAC). The platform integrates with Tally ERP for accounting and supports background job processing for asynchronous operations.

**Key Features:**
- Multi-tenant architecture with tenant isolation
- Role-based access control (RBAC) for fine-grained permissions
- Product catalog with image storage (MinIO/S3)
- Inventory tracking across multiple warehouses
- Order and invoice management with PDF generation
- Tally ERP integration for data import/export
- Real-time analytics and caching with Redis
- Background job processing with Asynq
- Subscription management with Razorpay integration
- Notification system (Email/SMS via SendGrid/Twilio)

## Technology Stack

### Backend
- **Language:** Go 1.25
- **Framework:** Echo v4 (HTTP router)
- **Database:** PostgreSQL 17 with pgx/v5 driver
- **Cache/Queue:** Redis 8
- **Object Storage:** MinIO (S3-compatible)
- **Background Jobs:** Asynq (Redis-backed task queue)
- **Authentication:** JWT with golang-jwt/jwt/v5
- **PDF Generation:** gofpdf

### Frontend
- **Framework:** Next.js 15.5.4 with Turbopack
- **React:** 19.1.0
- **State Management:** TanStack Query (React Query)
- **Styling:** Tailwind CSS 4
- **UI Components:** Custom components with Lucide icons
- **Charts:** Recharts

### Infrastructure
- **Containerization:** Docker & Docker Compose
- **Services:** PostgreSQL (port 5440), Redis (port 6379), MinIO (ports 9003/9004)
- **Admin Tools:** pgAdmin (port 5050), Redis Commander (port 8082)

## Development Commands

### Starting the Full Stack

```bash
# Start all services (backend, frontend, database, redis, minio)
./run_start.sh

# Stop all services
./run_stop.sh
```

### Backend Development

```bash
# Build the backend
go build -o main cmd/main.go

# Run the backend (requires infrastructure services running)
export DATABASE_URL="postgresql://testuser:testpass@localhost:5440/testdb"
export JWT_SECRET="development_secret_key_not_for_production"
export PORT=8081
./main

# Run backend tests
go test ./...

# Run tests for a specific package
go test ./internal/services -v

# Run a single test
go test ./internal/handlers -run TestProductHandlers

# Test with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Format code
go fmt ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

### Frontend Development

```bash
cd frontend

# Install dependencies
bun install

# Run development server on port 3000
bun run dev

# Build for production
bun run build

# Start production server
bun run start

# Run linter
bun run lint
```

### Database Management

```bash
# Run migrations
./run_migrations.sh

# Connect to PostgreSQL directly
docker exec -it inventory_management-postgres-1 psql -U testuser -d testdb

# View database tables
docker exec inventory_management-postgres-1 psql -U testuser -d testdb -c "\dt"

# Export database schema
docker exec inventory_management-postgres-1 pg_dump -U testuser -d testdb --schema-only > schema_backup.sql
```

### Docker Operations

```bash
# Start only infrastructure services
docker compose up -d postgres redis minio

# View logs
docker compose logs -f

# View logs for specific service
docker compose logs -f postgres

# Stop and remove volumes (WARNING: deletes all data)
docker compose down -v

# Rebuild containers
docker compose build --no-cache
```

## Architecture Overview

### Layer Structure

The backend follows a clean architecture pattern with clear separation of concerns:

```
cmd/
  main.go                    # Application entry point, service initialization

internal/
  handlers/                  # HTTP handlers (presentation layer)
  services/                  # Business logic layer
  repositories/              # Data access layer
  models/                    # Domain models
  middleware/                # HTTP middleware (JWT, RBAC, performance)
  jobs/                      # Background job handlers
  analytics/                 # Analytics aggregation service
  caching/                   # Redis caching abstraction
  common/                    # Shared utilities
  config/                    # Configuration loading

migrations/                  # SQL migration scripts
frontend/                    # Next.js frontend application
```

### Multi-Tenant Architecture

The system uses a **shared database, shared schema** approach with tenant isolation:

- Each entity (users, products, orders) has a `tenant_id` column
- All queries are scoped by tenant_id (extracted from JWT token)
- Middleware ensures tenant context is always present
- Foreign key constraints enforce referential integrity within tenants

### Authentication Flow

1. **Signup:** `POST /v1/auth/signup` - Creates user, assigns default role, issues JWT
2. **Login:** `POST /v1/auth/login` - Validates credentials, issues access + refresh tokens
3. **Token Refresh:** `POST /v1/auth/refresh` - Issues new access token using refresh token
4. **Logout:** `POST /v1/auth/logout` - Invalidates tokens (blacklisting in Redis)

All protected routes require JWT in `Authorization: Bearer <token>` header.

### RBAC (Role-Based Access Control)

- **Roles:** Admin, User (custom roles can be created)
- **Permissions:** Fine-grained permissions (e.g., `products:read`, `orders:create`)
- **Assignment:** Users → Roles → Permissions (many-to-many relationships)
- **Enforcement:** `RBACMiddleware.RequirePermission()` checks user permissions
- **Context:** User ID and Tenant ID are extracted from JWT and stored in request context

### Background Jobs (Asynq)

Background jobs run asynchronously using Redis as the queue backend:

- **Tally Export:** Export invoices/orders to Tally ERP
- **Tally Import:** Import data from Tally to database
- **Analytics Refresh:** Periodically refresh materialized analytics views
- **Inventory Alerts:** Send low-stock notifications

Jobs are enqueued via Asynq client and processed by Asynq server workers running in the main process.

### Caching Strategy

Redis is used for:
- **Authentication:** Token blacklisting, refresh token storage
- **RBAC:** User permissions caching (TTL-based)
- **Product Data:** Frequently accessed products
- **Analytics:** Aggregated metrics

Cache invalidation happens on data mutations (create, update, delete).

## Testing Strategy

### Unit Tests

Test files are located alongside implementation files with `_test.go` suffix.

**Mocking:**
- Database operations use `pgxmock` or `sqlmock`
- External services (MinIO, Redis, Razorpay) are mocked with interfaces

**Example test locations:**
- `internal/services/*_test.go` - Service layer tests
- `internal/handlers/*_test.go` - Handler tests
- `internal/jobs/*_test.go` - Background job tests

### Integration Tests

Located in `tests/integration/`:
- `permissions_rbac_test.go` - RBAC integration tests
- End-to-end workflow tests

### Performance Tests

Located in `tests/performance/`:
- `load_test.go` - Load testing scenarios

### Test Helpers

`testhelpers/` contains shared test utilities:
- `testing.go` - Common test fixtures and helpers
- `product_test.go` - Product-specific test helpers

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Database
DATABASE_URL=postgresql://testuser:testpass@localhost:5440/testdb

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-256-bit-secret-key-change-in-production

# Server
PORT=8081

# MinIO
MINIO_ENDPOINT=localhost:9003
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false

# Optional: External services
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-supabase-anon-key
RAZORPAY_KEY_ID=your-razorpay-key
RAZORPAY_KEY_SECRET=your-razorpay-secret
SENDGRID_API_KEY=your-sendgrid-key
TWILIO_ACCOUNT_SID=your-twilio-sid
TWILIO_AUTH_TOKEN=your-twilio-token
TWILIO_PHONE_NUMBER=your-twilio-phone
```

### Tally Configuration

Edit `config/tally.toml` for Tally ERP integration:

```toml
[tally]
api_endpoint = "https://api.tallysolutions.com"
api_key = "your_api_key"
api_secret = "your_api_secret"

[queuing]
redis_addr = "localhost:6379"
concurrency = 10
queue_priorities = { critical = 6, default = 3, low = 1 }

[export_import]
mode = "csv"  # or "rest"
timeout_seconds = 300
max_retry_attempts = 3
```

## Common Patterns

### Adding a New Entity

1. **Create Model:** `internal/models/entity.go`
   - Define struct with json/db tags
   - Include `tenant_id` for multi-tenant isolation

2. **Create Repository:** `internal/repositories/entity_repo.go`
   - Implement CRUD operations
   - Always filter by `tenant_id` in queries
   - Use pgx connection pool

3. **Create Service:** `internal/services/entity_service.go`
   - Implement business logic
   - Handle caching if needed
   - Return domain errors, not database errors

4. **Create Handlers:** `internal/handlers/entity_handlers.go`
   - Parse request body/params
   - Extract tenant/user context from request
   - Call service methods
   - Return proper HTTP status codes

5. **Register Routes:** `cmd/main.go`
   - Add routes to protected group
   - Apply RBAC middleware with appropriate permissions

6. **Add Permissions:** Create migration to add new permissions to `permissions` table

### Working with Context

Always extract tenant and user IDs from context:

```go
import "agromart2/internal/common"

func (h *Handler) SomeHandler(c echo.Context) error {
    ctx := c.Request().Context()
    
    userID, ok := common.GetUserIDFromContext(ctx)
    if !ok {
        return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
    }
    
    tenantID, ok := common.GetTenantIDFromContext(ctx)
    if !ok {
        return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
    }
    
    // Use userID and tenantID in your logic
}
```

### Implementing RBAC

Protect routes with permission checks:

```go
protected.GET("/products", productHandlers.ListProducts, 
    rbacMiddleware.RequirePermission("products:read"))
protected.POST("/products", productHandlers.CreateProduct,
    rbacMiddleware.RequirePermission("products:create"))
```

### Caching Patterns

```go
// Check cache first
cached, err := cacheSvc.Get(ctx, cacheKey)
if err == nil && cached != nil {
    return cached, nil
}

// Fetch from database
data, err := repo.GetData(ctx, id)
if err != nil {
    return nil, err
}

// Store in cache with TTL
_ = cacheSvc.Set(ctx, cacheKey, data, 5*time.Minute)

return data, nil
```

## Performance Optimizations

### Database

- **Connection Pooling:** Configured with 50 max connections, 10 min connections
- **Prepared Statements:** pgx automatically prepares frequently used statements
- **Indexes:** Check `migrations/*.sql` for indexed columns (especially `tenant_id`, foreign keys)
- **Partitioning:** See `migrations/20240130000002_table_partitioning.sql` for large table partitioning

### Middleware

- **Gzip Compression:** Enabled for all responses
- **Rate Limiting:** 100 requests/minute per IP
- **Request Timeout:** 30 seconds
- **Body Limit:** 10MB

### Caching

- Redis used for frequently accessed data
- Permissions cached per user with TTL
- Analytics results cached to reduce database load

### Frontend

- **Turbopack:** Enabled for faster builds and hot reload
- **React Query:** Automatic caching, background refetching
- **Code Splitting:** Next.js automatic code splitting

## Debugging

### Backend Logs

```bash
# Real-time backend logs
tail -f backend.log

# Search logs for errors
grep -i error backend.log

# View Asynq job processing logs
grep "Asynq" backend.log
```

### Frontend Logs

```bash
# Real-time frontend logs
tail -f frontend.log

# Next.js build logs
cat frontend.log | grep "Compil"
```

### Database Debugging

```bash
# Check running queries
docker exec inventory_management-postgres-1 psql -U testuser -d testdb -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' ORDER BY duration DESC;"

# Check table sizes
docker exec inventory_management-postgres-1 psql -U testuser -d testdb -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"

# Check index usage
docker exec inventory_management-postgres-1 psql -U testuser -d testdb -c "SELECT schemaname, tablename, indexname, idx_scan FROM pg_stat_user_indexes ORDER BY idx_scan ASC;"
```

### Redis Debugging

```bash
# Connect to Redis CLI
docker exec -it inventory_management-redis-1 redis-cli

# Inside Redis CLI:
# - View all keys: KEYS *
# - Get key value: GET key_name
# - Check key TTL: TTL key_name
# - Monitor commands: MONITOR
```

### MinIO Debugging

Access MinIO Console at `http://localhost:9004` (credentials: minioadmin/minioadmin)

```bash
# List buckets via CLI
docker exec inventory_management-minio-1 mc ls local/

# Check MinIO server logs
docker logs inventory_management-minio-1
```

## Security Considerations

### Authentication
- JWT tokens expire after 1 hour (access) / 24 hours (refresh)
- Passwords hashed with bcrypt
- Refresh tokens stored in Redis with expiry
- Logout invalidates tokens via Redis blacklist

### Authorization
- All routes (except health, login, signup) require authentication
- RBAC enforced at handler level
- Tenant isolation enforced at query level

### Data Protection
- Database connections use SSL in production
- Sensitive fields (passwords, API keys) never logged
- Environment variables for secrets, never hardcoded

### Input Validation
- Echo's request binding validates struct tags
- SQL injection prevented by using parameterized queries (pgx)
- File uploads validated for type and size

## Production Deployment

### Docker Build

```bash
# Build production image
docker build -t agromart2:latest .

# Run production container
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@host:5432/db" \
  -e REDIS_URL="redis://host:6379" \
  -e JWT_SECRET="production-secret-256-bits" \
  -e MINIO_ENDPOINT="minio.example.com" \
  -e MINIO_ACCESS_KEY="access_key" \
  -e MINIO_SECRET_KEY="secret_key" \
  agromart2:latest
```

### Health Checks

- **Basic:** `GET /health` - Returns 200 if server is up
- **Readiness:** `GET /health/ready` - Returns 200 if all dependencies are ready
- **Detailed:** `GET /health/detailed` - Returns JSON with database connection status

### Monitoring

```bash
# Prometheus metrics endpoint
curl http://localhost:8081/metrics
```

Key metrics exposed:
- HTTP request duration
- Active connections
- Database pool stats
- Redis operations
- Background job queue depth
