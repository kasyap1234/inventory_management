#!/bin/bash

# Comprehensive test runner script for the inventory management system
# Run all test suites: unit tests, integration tests, frontend tests, and performance tests

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Log levels
INFO="${BLUE}INFO${NC}"
WARN="${YELLOW}WARN${NC}"
ERROR="${RED}ERROR${NC}"
SUCCESS="${GREEN}SUCCESS${NC}"

echo -e "${INFO} Starting comprehensive test suite..."

# Track test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
START_TIME=$(date +%s)

# Function to run a test suite and track results
run_test_suite() {
    local suite_name="$1"
    local command="$2"
    local description="$3"

    echo -e "\n${BLUE}Running ${suite_name}${NC} - ${description}"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if eval "$command" > /tmp/test_output.log 2>&1; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${SUCCESS} ✓ ${suite_name} passed"
        return 0
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo -e "${ERROR} ✗ ${suite_name} failed"
        echo -e "${ERROR} Output:"
        cat /tmp/test_output.log
        echo ""
        return 1
    fi
}

# Check if required dependencies are installed
check_dependencies() {
    echo -e "${INFO} Checking dependencies..."

    # Check for Go
    if ! command -v go &> /dev/null; then
        echo -e "${ERROR} Go is not installed. Please install Go."
        exit 1
    fi

    # Check for Node.js (for frontend tests)
    if command -v node &> /dev/null; then
        HAS_NODE=true
        echo -e "${SUCCESS} Node.js found - Frontend tests will run"
    else
        HAS_NODE=false
        echo -e "${WARN} Node.js not found - Frontend tests will be skipped"
    fi

    # Check for PostgreSQL (test database)
    if ! command -v psql &> /dev/null && [ -z "$TEST_DATABASE_URL" ]; then
        echo -e "${WARN} PostgreSQL client not found and TEST_DATABASE_URL not set."
        echo -e "${WARN} Some tests may be skipped."
    fi
}

# Run repository layer tests (unit tests for database operations)
run_repository_tests() {
    echo -e "\n${BLUE}=== REPOSITORY LAYER TESTS ===${NC}"

    if ! run_test_suite "Product Repository" \
        "go test ./internal/repositories -v -run TestProductRepository" \
        "Testing product repository CRUD operations"; then
        return 1
    fi

    echo -e "${SUCCESS} All repository layer tests completed"
}

# Run service layer tests (business logic unit tests)
run_service_tests() {
    echo -e "\n${BLUE}=== SERVICE LAYER TESTS ===${NC}"

    if ! run_test_suite "Product Service" \
        "go test ./internal/services -v -run TestProductService" \
        "Testing product service business logic"; then
        return 1
    fi

    echo -e "${SUCCESS} All service layer tests completed"
}

# Run integration tests (end-to-end API tests)
run_integration_tests() {
    echo -e "\n${BLUE}=== INTEGRATION TESTS ===${NC}"

    if ! run_test_suite "Product API Integration" \
        "go test ./tests/integration -v -run TestProductAPI" \
        "Testing product API endpoints end-to-end"; then
        return 1
    fi

    echo -e "${SUCCESS} All integration tests completed"
}

# Run frontend component tests
run_frontend_tests() {
    if [ "$HAS_NODE" = true ]; then
        echo -e "\n${BLUE}=== FRONTEND TESTS ===${NC}"

        # Navigate to frontend directory
        if ! run_test_suite "Frontend Component Tests" \
            "cd frontend && npm test -- --coverage --watchAll=false" \
            "Testing React components with Jest"; then
            return 1
        fi

        echo -e "${SUCCESS} All frontend tests completed"
    else
        echo -e "\n${WARN} Skipping frontend tests - Node.js not available"
    fi
}

# Run performance tests and benchmarks
run_performance_tests() {
    echo -e "\n${BLUE}=== PERFORMANCE TESTS ===${NC}"

    # Run load tests
    if ! run_test_suite "Load Tests" \
        "go test ./tests/performance -v -run TestLoadProductService" \
        "Testing concurrent load handling"; then
        echo -e "${WARN} Load tests failed or were skipped (may be normal if database not configured)"
    fi

    # Run stress tests
    if ! run_test_suite "Stress Tests" \
        "go test ./tests/performance -v -run TestStressProductService" \
        "Testing extreme load conditions"; then
        echo -e "${WARN} Stress tests failed or were skipped (may be normal if database not configured)"
    fi

    # Run benchmarks
    echo -e "\n${BLUE}Running Benchmarks${NC}"
    if go test ./tests/performance -bench=. -benchmem -run=^$ -timeout=30m > /tmp/benchmarks.log 2>&1; then
        echo -e "${SUCCESS} ✓ Benchmarks completed"
        echo -e "${INFO} Benchmark results:"
        cat /tmp/benchmarks.log | grep -E "^Benchmark|PASS"
    else
        echo -e "${WARN} Benchmarks failed or were skipped"
    fi
}

# Generate coverage reports
generate_coverage() {
    echo -e "\n${BLUE}=== COVERAGE REPORTS ===${NC}"

    # Backend coverage
    echo -e "${INFO} Generating backend coverage report..."
    if go test ./internal/... ./tests/... -coverprofile=/tmp/golang_coverage.out -covermode=count > /dev/null 2>&1; then
        go tool cover -html=/tmp/golang_coverage.out -o coverage.html
        echo -e "${SUCCESS} Backend coverage report generated: coverage.html"
        go tool cover -func=/tmp/golang_coverage.out | tail -1
    else
        echo -e "${WARN} Backend coverage report generation failed"
    fi

    # Frontend coverage (if available)
    if [ "$HAS_NODE" = true ] && [ -d "frontend" ]; then
        echo -e "${INFO} Frontend coverage should be available in frontend/coverage/"
    fi
}

# Run linter checks
run_linting() {
    echo -e "\n${BLUE}=== LINTING ===${NC}"

    if command -v golangci-lint &> /dev/null; then
        if run_test_suite "Go Linting" \
            "golangci-lint run --timeout=5m" \
            "Running comprehensive Go linting checks"; then
            echo -e "${SUCCESS} Go code passed linting"
        fi
    else
        echo -e "${WARN} golangci-lint not installed, skipping Go linting"
    fi

    if [ "$HAS_NODE" = true ] && command -v eslint &> /dev/null; then
        if cd frontend && eslint . --quiet > /dev/null 2>&1; then
            echo -e "${SUCCESS} Frontend code passed linting"
        else
            echo -e "${WARN} Frontend linting found issues or eslint not configured"
        fi
    fi
}

# Main execution function
main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   INVENTORY MANAGEMENT - TEST SUITE   ${NC}"
    echo -e "${BLUE}========================================${NC}"

    check_dependencies

    # Run test suites
    run_repository_tests
    run_service_tests
    run_integration_tests
    run_frontend_tests
    run_performance_tests

    # Optional checks
    run_linting
    generate_coverage

    # Calculate total execution time
    END_TIME=$(date +%s)
    EXECUTION_TIME=$((END_TIME - START_TIME))

    # Print summary
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}             TEST SUMMARY              ${NC}"
    echo -e "${BLUE}========================================${NC}"

    echo -e "${INFO} Total execution time: ${EXECUTION_TIME} seconds"
    echo -e "${INFO} Test suites run: ${TOTAL_TESTS}"
    echo -e "${SUCCESS} Passed: ${PASSED_TESTS}"
    if [ $FAILED_TESTS -gt 0 ]; then
        echo -e "${ERROR} Failed: ${FAILED_TESTS}"
        echo -e "${WARN} Some tests failed - check output above for details"
        exit 1
    else
        echo -e "${SUCCESS} All tests passed! 🎉"
    fi

    # Performance test information
    if [ -f "/tmp/benchmarks.log" ]; then
        echo -e "\n${INFO} Top benchmark results:"
        cat /tmp/benchmarks.log | grep "^Benchmark" | head -5
    fi

    echo -e "${BLUE}========================================${NC}"
}

# Allow running specific test suites
case "$1" in
    "repo"|"repository")
        check_dependencies
        run_repository_tests
        ;;
    "service")
        check_dependencies
        run_service_tests
        ;;
    "integration"|"int")
        check_dependencies
        run_integration_tests
        ;;
    "frontend"|"fe")
        check_dependencies
        [ "$HAS_NODE" = true ] && run_frontend_tests
        ;;
    "performance"|"perf")
        check_dependencies
        run_performance_tests
        ;;
    "lint")
        run_linting
        ;;
    "coverage")
        generate_coverage
        ;;
    *)
        main
        ;;
esac
