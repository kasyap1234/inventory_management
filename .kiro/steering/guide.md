---
inclusion: always
---

# Agromart2 Development Guide

## Core Principles

- Do not create markdown documentation files after completing tasks
- Multi-tenant architecture: All operations must be tenant-scoped using `tenant_id`
- Context-based tenant extraction: Use `common.GetTenantIDFromContext(ctx)` in handlers
- UUID validation: Use `common.ValidateUUID()` for all UUID parameters
- Error handling: Return structured errors with proper HTTP status codes

## Backend Patterns

### Handler Layer
- Constructor pattern: `NewXxxHandlers(service, middleware) *XxxHandlers`
- Tenant extraction from context at the start of each handler
- UUID validation using `common.ValidateUUID()` helper
- Structured JSON responses with `message` and data fields
- HTTP status codes: 200 (OK), 201 (Created), 400 (Bad Request), 401 (Unauthorized), 500 (Internal Server Error)

### Service Layer
- Interface-based design for testability
- Constructor pattern: `NewXxxService(repo, ...) XxxService`
- Business logic validation before repository calls
- Cache integration: Check cache first, fallback to DB, then cache the result
- Cache invalidation on updates/deletes
- Context propagation throughout the call chain

### Repository Layer
- Direct database access using pgx/v5
- Tenant isolation in all queries: `WHERE tenant_id = $1`
- Parameterized queries to prevent SQL injection
- Return domain models, not database rows
- Pagination support: `LIMIT` and `OFFSET` parameters

### Database Conventions
- All tables have `tenant_id uuid NOT NULL` for multi-tenancy
- Primary keys are UUIDs
- Timestamps: `created_at` and `updated_at` (auto-managed)
- Foreign keys with proper constraints and indexes
- Use `COALESCE` for nullable fields in aggregations

### Validation
- Input validation in handlers before service calls
- Required field checks with descriptive error messages
- Range validation for numeric fields (prices > 0, quantities >= 0)
- UUID format validation using `common.ValidateUUID()`
- Date format: "2006-01-02" for parsing

### Caching Strategy
- Cache key format: `tenant:{tenantID}:resource:{resourceID}`
- TTL: 15 minutes for frequently accessed data
- Cache-aside pattern: Read from cache, fallback to DB, write to cache
- Invalidate on write operations (create, update, delete)
- Log cache errors but don't fail operations

### Bulk Operations
- Support batch processing with validation modes: "strict", "skip_invalid"
- Transaction modes: "atomic", "best_effort"
- Return detailed results with success/failure counts
- Progress tracking for long-running operations
- Item-level error reporting

## Frontend Patterns

### Component Structure
- Client components: Use `'use client'` directive
- Props interface: Define TypeScript interfaces for all props
- Controlled forms: Use `useState` for form data
- Conditional rendering: Optional callbacks with `?.()` syntax

### Data Fetching
- TanStack Query for server state management
- Query keys: `['resource', id]` or `['resources', filters]`
- Mutations with optimistic updates and cache invalidation
- Error handling with toast notifications

### API Integration
- Centralized API client in `lib/api.ts`
- Axios for HTTP requests with interceptors
- Base URL from environment variables
- JWT token in Authorization header

### Form Handling
- Controlled inputs with `value` and `onChange`
- Form submission with `e.preventDefault()`
- Loading states during mutations
- Success callbacks for navigation/modal closing

### Styling
- Tailwind CSS utility classes
- Consistent spacing: `space-y-4`, `gap-4`
- Grid layouts: `grid grid-cols-2 gap-4`
- Responsive design with breakpoint prefixes

## File Organization

### Backend
- `internal/handlers/`: HTTP request handlers
- `internal/services/`: Business logic layer
- `internal/repositories/`: Data access layer
- `internal/models/`: Domain models and DTOs
- `internal/middleware/`: HTTP middleware (auth, RBAC, logging)
- `internal/common/`: Shared utilities (validation, logging, context)

### Frontend
- `app/`: Next.js App Router pages
- `components/`: Reusable UI components
- `hooks/`: Custom React hooks
- `lib/`: Utility functions and API client
- `types/`: TypeScript type definitions

## Naming Conventions

### Go
- Files: `snake_case.go` (e.g., `product_handlers.go`)
- Types: `PascalCase` (e.g., `ProductService`)
- Interfaces: `PascalCase` ending with interface name (e.g., `ProductRepository`)
- Functions: `PascalCase` for exported, `camelCase` for private
- Variables: `camelCase`

### TypeScript/React
- Files: `PascalCase.tsx` for components, `camelCase.ts` for utilities
- Components: `PascalCase` (e.g., `ProductForm`)
- Hooks: `camelCase` starting with "use" (e.g., `useAuth`)
- Props interfaces: `ComponentNameProps`

## Security

- JWT authentication required for all protected routes
- RBAC middleware for permission checks
- CSRF protection on state-changing operations
- Input sanitization and validation
- SQL injection prevention via parameterized queries
- Tenant isolation enforced at database level

## Testing

- Unit tests: `*_test.go` files alongside source
- Test helpers in `testhelpers/` directory
- Mock repositories for service testing
- Integration tests in `tests/integration/`
- Use testify for assertions
