#!/bin/bash

# Setup script for Inventory Management System
# This script installs all dependencies and sets up the development environment

set -e

echo "🚀 Setting up Inventory Management System..."
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go 1.25 or higher.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Go is installed${NC}"

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js is not installed. Please install Node.js 20 or higher.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Node.js is installed${NC}"

# Check if Bun is installed
if ! command -v bun &> /dev/null; then
    echo -e "${BLUE}ℹ Bun is not installed. Installing Bun...${NC}"
    curl -fsSL https://bun.sh/install | bash
fi

echo -e "${GREEN}✓ Bun is installed${NC}"
echo ""

# Backend setup
echo -e "${BLUE}📦 Installing Go dependencies...${NC}"
go mod download
go mod tidy
echo -e "${GREEN}✓ Go dependencies installed${NC}"
echo ""

# Frontend setup
echo -e "${BLUE}📦 Installing frontend dependencies...${NC}"
cd frontend
bun install
cd ..
echo -e "${GREEN}✓ Frontend dependencies installed${NC}"
echo ""

# Check for .env file
if [ ! -f ".env" ]; then
    echo -e "${BLUE}📝 Creating .env file from example...${NC}"
    if [ -f ".env.example" ]; then
        cp .env.example .env
        echo -e "${GREEN}✓ .env file created. Please update it with your configuration.${NC}"
    else
        echo -e "${RED}❌ .env.example not found${NC}"
    fi
else
    echo -e "${GREEN}✓ .env file exists${NC}"
fi
echo ""

# Check for database
echo -e "${BLUE}🗄️  Checking database connection...${NC}"
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}⚠️  DATABASE_URL not set in environment${NC}"
    echo -e "${BLUE}Please set DATABASE_URL in your .env file${NC}"
else
    echo -e "${GREEN}✓ DATABASE_URL is set${NC}"
fi
echo ""

# Build backend
echo -e "${BLUE}🔨 Building backend...${NC}"
go build -o agromart cmd/main.go
echo -e "${GREEN}✓ Backend built successfully${NC}"
echo ""

# Summary
echo -e "${GREEN}✅ Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Update .env with your configuration"
echo "2. Run database migrations: ./run_migrations.sh"
echo "3. Start backend: ./agromart"
echo "4. Start frontend: cd frontend && bun run dev"
echo ""
echo "For more information, see:"
echo "- README.md"
echo "- UI_IMPROVEMENTS.md"
echo "- INSTALLATION_GUIDE.md"
