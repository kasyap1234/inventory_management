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

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Stopping Agromart Application Stack  ${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Stop backend
if [ -f "backend.pid" ]; then
    BACKEND_PID=$(cat backend.pid)
    echo -e "${YELLOW}Stopping backend (PID: $BACKEND_PID)...${NC}"
    if kill -0 $BACKEND_PID 2>/dev/null; then
        kill $BACKEND_PID
        echo -e "${GREEN}✓ Backend stopped${NC}"
    else
        echo -e "${YELLOW}Backend process not running${NC}"
    fi
    rm backend.pid
else
    echo -e "${YELLOW}No backend PID file found${NC}"
fi

# Stop frontend
if [ -f "frontend.pid" ]; then
    FRONTEND_PID=$(cat frontend.pid)
    echo -e "${YELLOW}Stopping frontend (PID: $FRONTEND_PID)...${NC}"
    if kill -0 $FRONTEND_PID 2>/dev/null; then
        kill $FRONTEND_PID
        echo -e "${GREEN}✓ Frontend stopped${NC}"
    else
        echo -e "${YELLOW}Frontend process not running${NC}"
    fi
    rm frontend.pid
else
    echo -e "${YELLOW}No frontend PID file found${NC}"
fi

# Stop Docker services
echo -e "${YELLOW}Stopping Docker services...${NC}"
if docker compose version >/dev/null 2>&1; then
    docker compose down
else
    docker-compose down
fi

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Docker services stopped${NC}"
else
    echo -e "${RED}Failed to stop Docker services${NC}"
fi

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  All Services Stopped Successfully!   ${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${YELLOW}Note: To preserve data, volumes are not removed.${NC}"
echo -e "${YELLOW}To remove all data, run:${NC} ${RED}docker compose down -v${NC}\n"
