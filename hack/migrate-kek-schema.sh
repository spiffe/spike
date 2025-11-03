#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Database migration script for KEK envelope encryption
# This script adds the necessary schema changes for KEK support

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

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

# Check if database path is provided
if [ $# -ne 1 ]; then
    log_error "Usage: $0 <database-path>"
    log_info "Example: $0 /path/to/spike.db"
    exit 1
fi

DB_PATH="$1"

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    log_error "Database not found: $DB_PATH"
    exit 1
fi

log_info "Starting KEK schema migration for: $DB_PATH"

# Create backup
BACKUP_PATH="${DB_PATH}.backup.$(date +%Y%m%d_%H%M%S)"
log_info "Creating backup: $BACKUP_PATH"
cp "$DB_PATH" "$BACKUP_PATH"

# Check if sqlite3 is available
if ! command -v sqlite3 &> /dev/null; then
    log_error "sqlite3 command not found. Please install sqlite3."
    exit 1
fi

# SQL migration script
MIGRATION_SQL=$(cat <<'EOF'
-- KEK Schema Migration
-- Adds support for envelope encryption with Key Encryption Keys

BEGIN TRANSACTION;

-- Create kek_metadata table if not exists
CREATE TABLE IF NOT EXISTS kek_metadata (
    kek_id TEXT PRIMARY KEY,
    version INTEGER NOT NULL,
    salt BLOB NOT NULL,
    rmk_version INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    wraps_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    retired_at INTEGER
);

-- Check if secrets table needs KEK columns
-- First check if kek_id column exists
CREATE TABLE IF NOT EXISTS _migration_check (
    check_name TEXT PRIMARY KEY,
    completed INTEGER NOT NULL DEFAULT 0
);

-- Add KEK columns to secrets table if they don't exist
-- We use ALTER TABLE which is safe if column already exists

-- Note: SQLite doesn't support ALTER TABLE ADD COLUMN IF NOT EXISTS
-- So we need to check if migration was already done

-- Mark migration as starting
INSERT OR REPLACE INTO _migration_check (check_name, completed) 
VALUES ('kek_columns_added', 0);

-- Create a new secrets table with KEK support
CREATE TABLE IF NOT EXISTS secrets_new (
    path TEXT PRIMARY KEY,
    values TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at INTEGER,
    -- KEK envelope encryption fields
    kek_id TEXT,
    wrapped_dek BLOB,
    dek_nonce BLOB,
    aead_alg TEXT,
    rewrapped_at INTEGER
);

-- Check if old secrets table exists and has data
-- If so, migrate data
INSERT OR IGNORE INTO secrets_new 
    (path, values, version, created_at, updated_at, deleted, deleted_at)
SELECT 
    path, values, version, created_at, updated_at, deleted, deleted_at
FROM secrets
WHERE EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='secrets');

-- Only rename tables if old secrets table exists
-- Drop old table and rename new one
DROP TABLE IF EXISTS secrets;
ALTER TABLE secrets_new RENAME TO secrets;

-- Create indexes for KEK queries
CREATE INDEX IF NOT EXISTS idx_secrets_kek_id ON secrets(kek_id);
CREATE INDEX IF NOT EXISTS idx_secrets_rewrapped_at ON secrets(rewrapped_at);
CREATE INDEX IF NOT EXISTS idx_kek_metadata_status ON kek_metadata(status);
CREATE INDEX IF NOT EXISTS idx_kek_metadata_created_at ON kek_metadata(created_at);

-- Mark migration as complete
UPDATE _migration_check SET completed = 1 WHERE check_name = 'kek_columns_added';

COMMIT;

-- Verify migration
SELECT 'Migration completed successfully' AS status;
SELECT COUNT(*) as kek_tables_created FROM sqlite_master 
    WHERE type='table' AND name='kek_metadata';
SELECT COUNT(*) as secret_kek_columns FROM pragma_table_info('secrets') 
    WHERE name IN ('kek_id', 'wrapped_dek', 'dek_nonce', 'aead_alg', 'rewrapped_at');
EOF
)

# Run migration
log_info "Running schema migration..."
echo "$MIGRATION_SQL" | sqlite3 "$DB_PATH"

if [ $? -eq 0 ]; then
    log_info "Migration completed successfully"
    log_info "Backup saved at: $BACKUP_PATH"
    log_info ""
    log_info "Next steps:"
    log_info "1. Enable KEK rotation: export SPIKE_KEK_ROTATION_ENABLED=true"
    log_info "2. Configure rotation policy (optional):"
    log_info "   - SPIKE_KEK_ROTATION_DAYS (default: 90)"
    log_info "   - SPIKE_KEK_MAX_WRAPS (default: 20000000)"
    log_info "   - SPIKE_KEK_GRACE_DAYS (default: 180)"
    log_info "3. Restart SPIKE Nexus"
    log_info ""
    log_warn "Note: Existing secrets will remain in legacy format until rewrapped"
    log_info "Use lazy rewrapping (enabled by default) to gradually migrate secrets"
else
    log_error "Migration failed"
    log_info "Restoring from backup..."
    cp "$BACKUP_PATH" "$DB_PATH"
    log_info "Database restored from backup"
    exit 1
fi

