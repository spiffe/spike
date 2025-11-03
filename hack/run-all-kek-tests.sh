#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Comprehensive test runner for KEK functionality
# Runs unit tests, integration tests, and optionally e2e tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

FAILED=0

run_unit_tests() {
    log_step "Running KEK unit tests..."
    echo ""
    
    cd "$PROJECT_ROOT"
    
    if go test -v ./app/nexus/internal/state/kek/... -cover; then
        log_info "Unit tests passed"
        return 0
    else
        log_error "Unit tests failed"
        FAILED=1
        return 1
    fi
}

run_build_test() {
    log_step "Testing build of all KEK packages..."
    echo ""
    
    cd "$PROJECT_ROOT"
    
    if go build ./app/nexus/internal/state/kek/... && \
       go build ./app/nexus/internal/route/kek/... && \
       go build ./internal/config/...; then
        log_info "All packages build successfully"
        return 0
    else
        log_error "Build failed"
        FAILED=1
        return 1
    fi
}

run_integration_tests() {
    log_step "Running integration tests..."
    echo ""
    
    log_warn "Integration tests require SPIKE Nexus to be running"
    log_info "Checking if Nexus is available..."
    
    if curl -s -f "http://localhost:8553/v1/kek/stats" > /dev/null 2>&1; then
        log_info "Nexus is running, executing integration tests"
        "$SCRIPT_DIR/test-kek-integration.sh"
        return $?
    else
        log_warn "Nexus is not running, skipping integration tests"
        log_info "To run integration tests:"
        log_info "  Terminal 1: ./hack/start-nexus-with-kek.sh"
        log_info "  Terminal 2: ./hack/run-all-kek-tests.sh"
        return 0
    fi
}

run_e2e_tests() {
    log_step "Running end-to-end tests..."
    echo ""
    
    if curl -s -f "http://localhost:8553/v1/kek/stats" > /dev/null 2>&1; then
        log_info "Nexus is running, executing e2e tests"
        "$SCRIPT_DIR/test-kek-e2e.sh"
        return $?
    else
        log_warn "Nexus is not running, skipping e2e tests"
        return 0
    fi
}

main() {
    log_info "=========================================="
    log_info "KEK Comprehensive Test Suite"
    log_info "=========================================="
    echo ""
    
    # Run unit tests
    run_unit_tests
    echo ""
    
    # Test builds
    run_build_test
    echo ""
    
    # Run integration tests if available
    run_integration_tests || true
    echo ""
    
    # Run e2e tests if available
    run_e2e_tests || true
    echo ""
    
    # Summary
    log_info "=========================================="
    log_info "Test Suite Complete"
    log_info "=========================================="
    
    if [ $FAILED -eq 0 ]; then
        log_info "All tests passed!"
        exit 0
    else
        log_error "Some tests failed"
        exit 1
    fi
}

main "$@"

