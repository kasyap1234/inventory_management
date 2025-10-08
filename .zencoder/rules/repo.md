---
description: Repository Information Overview
alwaysApply: true
---

# Agromart Inventory Management System Information

## Summary
A comprehensive inventory management system built with Go backend and Next.js frontend. The application provides features for managing products, inventory, orders, invoices, warehouses, suppliers, and distributors with role-based access control.

## Structure
- **cmd/**: Contains the main application entry point
- **internal/**: Core application code (models, repositories, services, handlers)
- **frontend/**: Next.js web application with TypeScript and Tailwind CSS
- **migrations/**: SQL database migration scripts
- **config/**: Configuration files
- **docs/**: API documentation
- **k8s/**: Kubernetes deployment configurations
- **tests/**: Integration and performance tests

## Language & Runtime
**Backend Language**: Go
**Version**: 1.25.0
**Build System**: Go modules
**Package Manager**: Go modules

**Frontend Language**: TypeScript
**Framework**: Next.js 15.5.4
**Package Manager**: npm

## Dependencies

### Backend Dependencies
**Main Dependencies**:
- github.com/labstack/echo/v4: Web framework
- github.com/jackc/pgx/v5: PostgreSQL driver
- github.com/hibiken/asynq: Distributed task queue
- github.com/redis/go-redis/v9: Redis client
- github.com/minio/minio-go/v7: MinIO S3 client
- github.com/golang-jwt/jwt/v5: JWT authentication
- github.com/BurntSushi/toml: TOML configuration

**Development Dependencies**:
- github.com/stretchr/testify: Testing framework
- github.com/DATA-DOG/go-sqlmock: SQL mocking

### Frontend Dependencies
**Main Dependencies**:
- react: 19.1.0
- next: 15.5.4
- @tanstack/react-query: API data fetching
- axios: HTTP client
- recharts: Data visualization
- tailwindcss: CSS framework

## Build & Installation

### Backend
```bash
# Build the backend
go build -o main cmd/main.go

# Run the backend
./main
```

### Frontend
```bash
# Install dependencies
cd frontend
npm install

# Development mode
npm run dev

# Production build
npm run build
npm start
```

### Full Stack Deployment
```bash
# Start all services (PostgreSQL, Redis, MinIO, Backend, Frontend)
./run_start.sh

# Stop all services
./run_stop.sh
```

## Docker
**Dockerfile**: Multi-stage build for optimized production deployment
**Image**: Alpine-based with minimal dependencies
**Configuration**: Docker Compose setup with PostgreSQL, Redis, MinIO, pgAdmin, and Redis Commander

```bash
# Build and run with Docker Compose
docker compose up -d
```

## Testing
**Framework**: github.com/stretchr/testify
**Test Location**: 
- Unit tests: Located alongside implementation files (*_test.go)
- Integration tests: tests/integration/
- Performance tests: tests/performance/

**Run Command**:
```bash
# Run all tests
go test ./...

# Run specific tests
go test ./internal/services/...
```

## Database
**Type**: PostgreSQL 17
**Migrations**: SQL migration files in migrations/ directory
**Schema**: Role-based access control with multi-tenant support

## Services
**Authentication**: JWT-based with refresh tokens
**Storage**: MinIO for file storage (product images)
**Caching**: Redis for performance optimization
**Background Jobs**: Asynq for task processing
**Analytics**: Built-in data analysis for inventory and sales