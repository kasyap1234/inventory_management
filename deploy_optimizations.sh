#!/bin/bash

# ============================================
# Deployment Script for Performance Optimizations
# ============================================

set -e  # Exit on error

echo "🚀 Starting Performance Optimization Deployment..."
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if PostgreSQL is running
check_postgres() {
    echo "Checking PostgreSQL connection..."
    if psql -U postgres -d inventory_db -c "SELECT 1" > /dev/null 2>&1; then
        print_success "PostgreSQL is running"
        return 0
    else
        print_error "PostgreSQL is not accessible"
        return 1
    fi
}

# Apply database optimizations
apply_database_optimizations() {
    echo ""
    echo "📊 Applying database optimizations..."
    
    if [ -f "DATABASE_OPTIMIZATION.sql" ]; then
        psql -U postgres -d inventory_db -f DATABASE_OPTIMIZATION.sql
        print_success "Database optimizations applied"
    else
        print_warning "DATABASE_OPTIMIZATION.sql not found"
    fi
}

# Set up materialized view refresh
setup_mv_refresh() {
    echo ""
    echo "⏰ Setting up materialized view refresh..."
    
    # Check if pg_cron is available
    if psql -U postgres -d inventory_db -c "CREATE EXTENSION IF NOT EXISTS pg_cron" > /dev/null 2>&1; then
        psql -U postgres -d inventory_db -c "
            SELECT cron.schedule(
                'refresh-analytics',
                '*/5 * * * *',
                'SELECT refresh_analytics_views()'
            );
        "
        print_success "Materialized view refresh scheduled (every 5 minutes)"
    else
        print_warning "pg_cron not available. Set up manual cron job:"
        echo "  */5 * * * * psql -U postgres -d inventory_db -c \"SELECT refresh_analytics_views()\""
    fi
}

# Build frontend
build_frontend() {
    echo ""
    echo "🎨 Building optimized frontend..."
    
    cd frontend
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        echo "Installing dependencies..."
        npm install
    fi
    
    # Build production bundle
    npm run build
    
    if [ $? -eq 0 ]; then
        print_success "Frontend built successfully"
    else
        print_error "Frontend build failed"
        exit 1
    fi
    
    cd ..
}

# Build backend
build_backend() {
    echo ""
    echo "⚙️  Building optimized backend..."
    
    # Build with optimizations
    go build -ldflags="-s -w" -o agromart cmd/main.go
    
    if [ $? -eq 0 ]; then
        print_success "Backend built successfully"
    else
        print_error "Backend build failed"
        exit 1
    fi
}

# Run database analysis
analyze_database() {
    echo ""
    echo "📈 Analyzing database performance..."
    
    psql -U postgres -d inventory_db << EOF
-- Analyze all tables
ANALYZE products;
ANALYZE inventory;
ANALYZE orders;
ANALYZE invoices;

-- Check cache hit ratio
SELECT 
    'Cache Hit Ratio' as metric,
    ROUND(sum(heap_blks_hit)::numeric / NULLIF(sum(heap_blks_hit) + sum(heap_blks_read), 0) * 100, 2) || '%' as value
FROM pg_statio_user_tables;

-- Check index usage
SELECT 
    'Indexes Created' as metric,
    COUNT(*)::text as value
FROM pg_indexes 
WHERE schemaname = 'public';

-- Check table sizes
SELECT 
    'Total Database Size' as metric,
    pg_size_pretty(pg_database_size(current_database())) as value;
EOF
    
    print_success "Database analysis complete"
}

# Verify optimizations
verify_optimizations() {
    echo ""
    echo "🔍 Verifying optimizations..."
    
    # Check if indexes exist
    INDEX_COUNT=$(psql -U postgres -d inventory_db -t -c "SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public'")
    
    if [ "$INDEX_COUNT" -gt 20 ]; then
        print_success "Database indexes verified ($INDEX_COUNT indexes)"
    else
        print_warning "Expected more indexes (found $INDEX_COUNT)"
    fi
    
    # Check if materialized views exist
    MV_COUNT=$(psql -U postgres -d inventory_db -t -c "SELECT COUNT(*) FROM pg_matviews")
    
    if [ "$MV_COUNT" -gt 0 ]; then
        print_success "Materialized views verified ($MV_COUNT views)"
    else
        print_warning "No materialized views found"
    fi
    
    # Check if frontend build exists
    if [ -d "frontend/.next" ]; then
        print_success "Frontend build verified"
    else
        print_warning "Frontend build not found"
    fi
    
    # Check if backend binary exists
    if [ -f "agromart" ]; then
        print_success "Backend binary verified"
    else
        print_warning "Backend binary not found"
    fi
}

# Main deployment flow
main() {
    echo "============================================"
    echo "  Performance Optimization Deployment"
    echo "============================================"
    echo ""
    
    # Check prerequisites
    if ! check_postgres; then
        print_error "Please start PostgreSQL and try again"
        exit 1
    fi
    
    # Ask for confirmation
    read -p "This will apply database optimizations and rebuild the application. Continue? (y/n) " -n 1 -r
    echo ""
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Deployment cancelled"
        exit 0
    fi
    
    # Apply optimizations
    apply_database_optimizations
    setup_mv_refresh
    analyze_database
    
    # Build applications
    build_frontend
    build_backend
    
    # Verify everything
    verify_optimizations
    
    echo ""
    echo "============================================"
    print_success "Deployment Complete!"
    echo "============================================"
    echo ""
    echo "Next steps:"
    echo "1. Review the performance metrics above"
    echo "2. Start the backend: ./agromart"
    echo "3. Start the frontend: cd frontend && npm start"
    echo "4. Monitor performance with: psql -U postgres -d inventory_db -f check_performance.sql"
    echo ""
    echo "📚 Documentation:"
    echo "  - PERFORMANCE_OPTIMIZATION_COMPLETE.md"
    echo "  - OPTIMIZATION_SUMMARY.md"
    echo "  - IMPLEMENTATION_GUIDE.md"
    echo ""
}

# Run main function
main
