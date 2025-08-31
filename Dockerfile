# Multi-stage Dockerfile for Agromart2 SaaS Platform

# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata postgresql-client redis curl

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-s -w -X main.version=${BUILD_VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o agromart2 ./cmd/main.go

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies for health checks
RUN apk --no-cache add ca-certificates tzdata curl postgresql-client redis

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/agromart2 .

# Copy migrations, docs, and config files
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs

# Create directory for temporary files
RUN mkdir -p /tmp && chown -R appuser:appuser /tmp

# Change ownership and set USER
RUN chown -R appuser:appuser /app
USER appuser

# Environment variables
ENV PORT=8080
ENV DATABASE_URL=""
ENV REDIS_URL=""
ENV JWT_SECRET=""
ENV SUPABASE_URL=""
ENV SUPABASE_ANON_KEY=""

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:${PORT}/health || exit 1

# Expose port
EXPOSE ${PORT}

# Run the binary
CMD ["./agromart2"]