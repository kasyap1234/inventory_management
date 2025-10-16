# Technology Stack

## Backend
- **Language**: Go 1.25
- **Framework**: Echo v4 (HTTP router/middleware)
- **Database**: PostgreSQL 17.5 with pgx/v5 driver
- **Caching**: Redis 8
- **Authentication**: JWT with golang-jwt/jwt/v5
- **Background Jobs**: Asynq (Redis-based job queue)
- **File Storage**: MinIO (S3-compatible object storage)
- **Validation**: go-playground/validator/v10
- **Testing**: testify/v1.11.1

## Frontend
- **Framework**: Next.js 15.5.4 with React 19.1.0
- **Language**: TypeScript 5
- **Package Manager**: Bun 1.2.20
- **Styling**: Tailwind CSS v4
- **State Management**: TanStack React Query v5.90.2
- **HTTP Client**: Axios 1.12.2
- **UI Components**: Lucide React icons, custom components
- **Charts**: Recharts 3.2.1
- **Notifications**: React Hot Toast

## Infrastructure
- **Containerization**: Docker with multi-stage builds
- **Orchestration**: Docker Compose, Kubernetes ready
- **Database Migrations**: Custom SQL migration system
- **Environment**: .env configuration

## Common Commands

### Backend Development
```bash
# Run the application
go run cmd/main.go

# Build binary
go build -o agromart2 cmd/main.go

# Run tests
go test ./...

# Run specific test packages
go test ./internal/handlers/...
go test ./internal/services/...

# Database migrations
./run_migrations.sh
```

### Frontend Development
```bash
# Install dependencies
bun install

# Development server
bun run dev

# Build for production
bun run build

# Start production server
bun run start

# Lint code
bun run lint
```

### Docker Operations
```bash
# Build and run with Docker Compose
docker-compose up --build

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Testing & Debugging
```bash
# API endpoint testing
./test_api_endpoints.sh

# End-to-end workflow testing
./test_end_to_end_workflow_complete.sh

# MinIO functionality testing
./test_minio_complete.sh
```