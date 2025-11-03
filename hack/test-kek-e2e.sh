#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# End-to-end test for KEK rotation functionality
# Tests actual secret operations with envelope encryption

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

TEST_DB="/tmp/spike-kek-test-$(date +%s).db"
NEXUS_PORT=8553
BASE_URL="http://localhost:${NEXUS_PORT}"
TEST_PASSED=0
TEST_FAILED=0

cleanup() {
    log_info "Cleaning up..."
    if [ -f "$TEST_DB" ]; then
        rm -f "$TEST_DB"
    fi
}

trap cleanup EXIT

test_secret_operations() {
    log_step "Testing secret creation and retrieval"
    
    # Create a test secret
    local secret_path="test/kek/secret1"
    local secret_value="supersecret123"
    
    log_info "Creating secret at path: $secret_path"
    response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "{\"path\":\"$secret_path\",\"values\":{\"password\":\"$secret_value\"}}" \
        "$BASE_URL/v1/secrets")
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
        log_info "Secret created successfully"
        ((TEST_PASSED++))
    else
        log_error "Failed to create secret: HTTP $http_code"
        echo "$body"
        ((TEST_FAILED++))
        return 1
    fi
    
    # Retrieve the secret
    log_info "Retrieving secret from path: $secret_path"
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/secrets/$secret_path")
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        log_info "Secret retrieved successfully"
        if echo "$body" | grep -q "$secret_value"; then
            log_info "Secret value matches"
            ((TEST_PASSED++))
        else
            log_error "Secret value mismatch"
            ((TEST_FAILED++))
            return 1
        fi
    else
        log_error "Failed to retrieve secret: HTTP $http_code"
        echo "$body"
        ((TEST_FAILED++))
        return 1
    fi
}

test_kek_rotation() {
    log_step "Testing KEK rotation"
    
    # Get current KEK
    log_info "Getting current KEK before rotation"
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/kek/current")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        old_kek_id=$(echo "$body" | jq -r '.kek_id' 2>/dev/null || echo "")
        log_info "Current KEK ID: $old_kek_id"
        ((TEST_PASSED++))
    else
        log_error "Failed to get current KEK: HTTP $http_code"
        ((TEST_FAILED++))
        return 1
    fi
    
    # Trigger rotation
    log_info "Triggering KEK rotation"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/v1/kek/rotate")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        new_kek_id=$(echo "$body" | jq -r '.new_kek_id' 2>/dev/null || echo "")
        log_info "Rotation successful. New KEK ID: $new_kek_id"
        
        if [ "$new_kek_id" != "$old_kek_id" ]; then
            log_info "KEK ID changed as expected"
            ((TEST_PASSED++))
        else
            log_warn "KEK ID did not change after rotation"
            ((TEST_FAILED++))
        fi
    else
        log_error "Failed to rotate KEK: HTTP $http_code"
        echo "$body"
        ((TEST_FAILED++))
        return 1
    fi
}

test_kek_list() {
    log_step "Testing KEK listing"
    
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/kek/list")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        kek_count=$(echo "$body" | jq -r '.count' 2>/dev/null || echo "0")
        log_info "Found $kek_count KEKs"
        
        if [ "$kek_count" -gt 0 ]; then
            log_info "KEK list retrieved successfully"
            ((TEST_PASSED++))
        else
            log_warn "No KEKs found in list"
            ((TEST_FAILED++))
        fi
    else
        log_error "Failed to list KEKs: HTTP $http_code"
        echo "$body"
        ((TEST_FAILED++))
        return 1
    fi
}

test_kek_stats() {
    log_step "Testing KEK statistics"
    
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/kek/stats")
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        log_info "KEK stats retrieved successfully"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        ((TEST_PASSED++))
    else
        log_error "Failed to get KEK stats: HTTP $http_code"
        echo "$body"
        ((TEST_FAILED++))
        return 1
    fi
}

test_secret_after_rotation() {
    log_step "Testing secret access after rotation"
    
    local secret_path="test/kek/secret2"
    local secret_value="rotatedsecret456"
    
    log_info "Creating secret after rotation"
    response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "{\"path\":\"$secret_path\",\"values\":{\"token\":\"$secret_value\"}}" \
        "$BASE_URL/v1/secrets")
    
    http_code=$(echo "$response" | tail -n 1)
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
        log_info "Secret created with new KEK"
        ((TEST_PASSED++))
    else
        log_error "Failed to create secret after rotation: HTTP $http_code"
        ((TEST_FAILED++))
        return 1
    fi
    
    # Retrieve old secret (should trigger lazy rewrap)
    log_info "Retrieving old secret (should trigger lazy rewrap)"
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/secrets/test/kek/secret1")
    
    http_code=$(echo "$response" | tail -n 1)
    
    if [ "$http_code" -eq 200 ]; then
        log_info "Old secret still accessible after rotation"
        ((TEST_PASSED++))
    else
        log_error "Failed to access old secret: HTTP $http_code"
        ((TEST_FAILED++))
        return 1
    fi
}

check_nexus_running() {
    log_info "Checking if SPIKE Nexus is running..."
    
    if curl -s -f "$BASE_URL/v1/kek/stats" > /dev/null 2>&1; then
        log_info "SPIKE Nexus is running with KEK support"
        return 0
    else
        log_error "SPIKE Nexus is not running or KEK rotation is not enabled"
        log_info ""
        log_info "To start SPIKE Nexus with KEK rotation:"
        log_info "  export SPIKE_KEK_ROTATION_ENABLED=true"
        log_info "  export SPIKE_BACKEND_STORE=sqlite"
        log_info "  export SPIKE_BACKEND_SQLITE_PATH=$TEST_DB"
        log_info "  ./spike-nexus"
        return 1
    fi
}

main() {
    log_info "=========================================="
    log_info "SPIKE KEK End-to-End Test Suite"
    log_info "=========================================="
    echo ""
    
    if ! check_nexus_running; then
        log_error "Cannot run tests without SPIKE Nexus"
        exit 1
    fi
    
    echo ""
    log_info "Starting test execution..."
    echo ""
    
    # Run tests
    test_kek_stats || true
    echo ""
    
    test_secret_operations || true
    echo ""
    
    test_kek_list || true
    echo ""
    
    test_kek_rotation || true
    echo ""
    
    test_secret_after_rotation || true
    echo ""
    
    # Summary
    log_info "=========================================="
    log_info "Test Results"
    log_info "=========================================="
    log_info "Passed: $TEST_PASSED"
    if [ $TEST_FAILED -eq 0 ]; then
        log_info "Failed: $TEST_FAILED"
    else
        log_error "Failed: $TEST_FAILED"
    fi
    echo ""
    
    if [ $TEST_FAILED -eq 0 ]; then
        log_info "All tests passed!"
        exit 0
    else
        log_error "Some tests failed"
        exit 1
    fi
}

main "$@"

