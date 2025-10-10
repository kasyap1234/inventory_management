#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

if [ -f ".env" ]; then
    set -a
    source ./.env
    set +a
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Starting Agromart Application Stack  ${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if a port is in use
port_in_use() {
    lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1
}

# Function to wait for a service to be ready
wait_for_service() {
    local host=$1
    local port=$2
    local service=$3
    local max_attempts=30
    local attempt=1

    echo -e "${YELLOW}Waiting for $service to be ready...${NC}"
    while ! nc -z "$host" "$port" >/dev/null 2>&1; do
        if [ $attempt -eq $max_attempts ]; then
            echo -e "${RED}$service failed to start after $max_attempts attempts${NC}"
            return 1
        fi
        echo -e "${YELLOW}Attempt $attempt/$max_attempts: $service not ready yet...${NC}"
        sleep 2
        attempt=$((attempt + 1))
    done
    echo -e "${GREEN}✓ $service is ready!${NC}\n"
    return 0
}

# Check prerequisites
echo -e "${BLUE}Checking prerequisites...${NC}"
MISSING_DEPS=0

if ! command_exists docker; then
    echo -e "${RED}✗ Docker is not installed${NC}"
    MISSING_DEPS=1
else
    echo -e "${GREEN}✓ Docker is installed${NC}"
fi

if ! command_exists docker-compose && ! docker compose version >/dev/null 2>&1; then
    echo -e "${RED}✗ Docker Compose is not installed${NC}"
    MISSING_DEPS=1
else
    echo -e "${GREEN}✓ Docker Compose is installed${NC}"
fi

if ! command_exists bun; then
    echo -e "${RED}✗ Bun is not installed${NC}"
    MISSING_DEPS=1
else
    echo -e "${GREEN}✓ Bun is installed ($(bun --version))${NC}"
fi

if ! command_exists go; then
    echo -e "${RED}✗ Go is not installed${NC}"
    MISSING_DEPS=1
else
    echo -e "${GREEN}✓ Go is installed ($(go version))${NC}"
fi

if [ $MISSING_DEPS -eq 1 ]; then
    echo -e "\n${RED}Please install missing dependencies and try again${NC}"
    exit 1
fi

echo ""

# Step 1: Start Docker services (PostgreSQL, Redis, MinIO)
echo -e "${BLUE}Step 1: Starting infrastructure services (PostgreSQL, Redis, MinIO, MailHog)...${NC}"
if docker compose version >/dev/null 2>&1; then
    docker compose up -d postgres redis minio mailhog
else
    docker-compose up -d postgres redis minio mailhog
fi

if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to start infrastructure services${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Infrastructure services started${NC}\n"

# Wait for services to be ready
wait_for_service localhost 5440 "PostgreSQL"
wait_for_service localhost 6379 "Redis"
wait_for_service localhost 9003 "MinIO"
wait_for_service localhost 1025 "MailHog SMTP"

# Run database migrations
echo -e "${BLUE}Running database migrations...${NC}"
if [ -f "./run_migrations.sh" ]; then
    ./run_migrations.sh
    if [ $? -ne 0 ]; then
        echo -e "${YELLOW}Warning: Database migrations encountered errors${NC}"
        echo -e "${YELLOW}This may be normal if tables already exist${NC}\n"
    fi
else
    echo -e "${YELLOW}Skipping migrations (run_migrations.sh not found)${NC}\n"
fi

# Step 2: Build and start the backend
echo -e "${BLUE}Step 2: Building and starting the backend...${NC}"

# Set environment variables for the backend
export PORT="${PORT:-8081}"
export DATABASE_URL="${DATABASE_URL:-postgresql://testuser:testpass@localhost:5440/testdb}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379}"
export JWT_SECRET="${JWT_SECRET:-development_secret_key_not_for_production}"
# Supabase credentials - load from .env if available
export SUPABASE_URL="${SUPABASE_URL:-}"
export SUPABASE_ANON_KEY="${SUPABASE_ANON_KEY:-}"

# Always (re)build the backend to pick up the latest code changes
echo -e "${YELLOW}Building backend...${NC}"
if ! go build -o main cmd/main.go; then
    echo -e "${RED}Failed to build backend${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Backend built successfully${NC}"

# Start the backend in the background
echo -e "${YELLOW}Starting backend server on port 8081...${NC}"
nohup ./main > backend.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > backend.pid
echo -e "${GREEN}✓ Backend started (PID: $BACKEND_PID)${NC}\n"

# Wait for backend to be ready
sleep 3
wait_for_service localhost 8081 "Backend API"

# Step 3: Start the frontend
echo -e "${BLUE}Step 3: Starting the frontend...${NC}"
cd "$PROJECT_ROOT/frontend"

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    bun install
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to install frontend dependencies${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Frontend dependencies installed${NC}"
fi

# Start the frontend in development mode
echo -e "${YELLOW}Starting frontend development server...${NC}"
nohup bun run dev > ../frontend.log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > ../frontend.pid
echo -e "${GREEN}✓ Frontend started (PID: $FRONTEND_PID)${NC}\n"

cd "$PROJECT_ROOT"

# Wait for frontend to be ready
sleep 5
wait_for_service localhost 3000 "Frontend"

# Summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  All Services Started Successfully!   ${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${BLUE}Service URLs:${NC}"
echo -e "  Frontend:          ${GREEN}http://localhost:3000${NC}"
echo -e "  Backend API:       ${GREEN}http://localhost:8081${NC}"
echo -e "  PostgreSQL:        ${GREEN}localhost:5440${NC}"
echo -e "  Redis:             ${GREEN}localhost:6379${NC}"
echo -e "  MinIO Console:     ${GREEN}http://localhost:9004${NC}"
echo -e "  MinIO API:         ${GREEN}http://localhost:9003${NC}"
echo -e "  MailHog UI:        ${GREEN}http://localhost:8025${NC}"
echo -e "  pgAdmin (optional):${GREEN}http://localhost:5050${NC}"
echo -e "  Redis Commander:   ${GREEN}http://localhost:8082${NC}\n"

echo -e "${BLUE}Credentials:${NC}"
echo -e "  PostgreSQL:        ${YELLOW}testuser / testpass${NC}"
echo -e "  MinIO:             ${YELLOW}minioadmin / minioadmin${NC}"
echo -e "  pgAdmin:           ${YELLOW}admin@agromart.com / admin${NC}\n"

echo -e "${BLUE}Logs:${NC}"
echo -e "  Backend:           ${YELLOW}tail -f backend.log${NC}"
echo -e "  Frontend:          ${YELLOW}tail -f frontend.log${NC}"
echo -e "  Docker services:   ${YELLOW}docker compose logs -f${NC}\n"

echo -e "${BLUE}Process IDs:${NC}"
echo -e "  Backend PID:       ${YELLOW}$BACKEND_PID${NC}"
echo -e "  Frontend PID:      ${YELLOW}$FRONTEND_PID${NC}\n"

echo -e "${YELLOW}To stop all services, run:${NC} ${GREEN}./run_stop.sh${NC}\n"
