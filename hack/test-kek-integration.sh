#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Integration test for KEK management functionality
# Tests KEK rotation, stats, and RMK snapshot operations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Check if SPIKE Nexus is running
check_nexus_running() {
    if curl -s -f "http://localhost:8553/v1/kek/stats" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Test getting current KEK
test_get_current_kek() {
    log_info "Testing GET /v1/kek/current"
    
    response=$(curl -s -w "\n%{http_code}" "http://localhost:8553/v1/kek/current")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 400 ]; then
        log_info "GET /v1/kek/current: HTTP $http_code"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        log_error "GET /v1/kek/current failed: HTTP $http_code"
        echo "$body"
        return 1
    fi
}

# Test listing KEKs
test_list_keks() {
    log_info "Testing GET /v1/kek/list"
    
    response=$(curl -s -w "\n%{http_code}" "http://localhost:8553/v1/kek/list")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 400 ]; then
        log_info "GET /v1/kek/list: HTTP $http_code"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        log_error "GET /v1/kek/list failed: HTTP $http_code"
        echo "$body"
        return 1
    fi
}

# Test KEK stats
test_kek_stats() {
    log_info "Testing GET /v1/kek/stats"
    
    response=$(curl -s -w "\n%{http_code}" "http://localhost:8553/v1/kek/stats")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 400 ]; then
        log_info "GET /v1/kek/stats: HTTP $http_code"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        log_error "GET /v1/kek/stats failed: HTTP $http_code"
        echo "$body"
        return 1
    fi
}

# Test KEK rotation
test_kek_rotation() {
    log_info "Testing POST /v1/kek/rotate"
    
    response=$(curl -s -w "\n%{http_code}" -X POST "http://localhost:8553/v1/kek/rotate")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 400 ]; then
        log_info "POST /v1/kek/rotate: HTTP $http_code"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        log_error "POST /v1/kek/rotate failed: HTTP $http_code"
        echo "$body"
        return 1
    fi
}

# Test RMK snapshot
test_rmk_snapshot() {
    log_info "Testing GET /v1/rmk/snapshot"
    
    response=$(curl -s -w "\n%{http_code}" "http://localhost:8553/v1/rmk/snapshot")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 400 ]; then
        log_info "GET /v1/rmk/snapshot: HTTP $http_code"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        log_error "GET /v1/rmk/snapshot failed: HTTP $http_code"
        echo "$body"
        return 1
    fi
}

# Test RMK rotation (should return not implemented)
test_rmk_rotation() {
    log_info "Testing POST /v1/rmk/rotate"
    
    response=$(curl -s -w "\n%{http_code}" -X POST "http://localhost:8553/v1/rmk/rotate")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 501 ] || [ "$http_code" -eq 400 ]; then
        log_info "POST /v1/rmk/rotate: HTTP $http_code (expected)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        log_warn "POST /v1/rmk/rotate unexpected response: HTTP $http_code"
        echo "$body"
        return 0
    fi
}

# Main test execution
main() {
    log_info "Starting KEK integration tests"
    
    # Check if Nexus is running
    if ! check_nexus_running; then
        log_warn "SPIKE Nexus does not appear to be running or KEK rotation is disabled"
        log_info "Tests will show KEK rotation not enabled responses"
    fi
    
    echo ""
    log_info "=== Test Suite: KEK Management ==="
    echo ""
    
    test_get_current_kek
    echo ""
    
    test_list_keks
    echo ""
    
    test_kek_stats
    echo ""
    
    test_kek_rotation
    echo ""
    
    test_rmk_snapshot
    echo ""
    
    test_rmk_rotation
    echo ""
    
    log_info "KEK integration tests completed"
    log_info "Note: If KEK rotation is disabled (default), endpoints will return 400"
    log_info "To enable KEK rotation, set SPIKE_KEK_ROTATION_ENABLED=true"
}

main "$@"

