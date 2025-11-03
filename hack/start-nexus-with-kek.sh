#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Start SPIKE Nexus with KEK rotation enabled for testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Configuration
TEST_DB="${TEST_DB:-/tmp/spike-kek-test.db}"
NEXUS_PORT="${NEXUS_PORT:-8553}"

log_info "=========================================="
log_info "Starting SPIKE Nexus with KEK Rotation"
log_info "=========================================="
echo ""

# Check if database needs migration
if [ -f "$TEST_DB" ]; then
    log_warn "Database exists at: $TEST_DB"
    log_warn "Run migration if needed: ./hack/migrate-kek-schema.sh $TEST_DB"
else
    log_info "Database will be created at: $TEST_DB"
fi

echo ""

# Build Nexus if needed
log_step "Building SPIKE Nexus..."
cd "$PROJECT_ROOT"
go build -o spike-nexus ./app/nexus/cmd/main.go

if [ $? -ne 0 ]; then
    log_error "Failed to build SPIKE Nexus"
    exit 1
fi

log_info "Build successful"
echo ""

# Configure environment
log_step "Configuring environment..."
export SPIKE_KEK_ROTATION_ENABLED=true
export SPIKE_KEK_ROTATION_DAYS=90
export SPIKE_KEK_MAX_WRAPS=20000000
export SPIKE_KEK_GRACE_DAYS=180
export SPIKE_KEK_LAZY_REWRAP_ENABLED=true
export SPIKE_KEK_MAX_REWRAP_QPS=100

export SPIKE_BACKEND_STORE=sqlite
export SPIKE_BACKEND_SQLITE_PATH="$TEST_DB"

log_info "KEK Rotation Configuration:"
log_info "  - Enabled: true"
log_info "  - Rotation Days: 90"
log_info "  - Max Wraps: 20000000"
log_info "  - Grace Days: 180"
log_info "  - Lazy Rewrap: true"
log_info "  - Max Rewrap QPS: 100"
echo ""
log_info "Backend Configuration:"
log_info "  - Store: sqlite"
log_info "  - Database: $TEST_DB"
log_info "  - Port: $NEXUS_PORT"
echo ""

# Start Nexus
log_step "Starting SPIKE Nexus..."
log_info "Press Ctrl+C to stop"
echo ""
echo "=========================================="
echo ""

exec ./spike-nexus

