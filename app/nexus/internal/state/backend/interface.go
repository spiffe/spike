//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"github.com/spiffe/spike/app/nexus/internal/state/store"
)

type DatabaseConfigKey string

const (
	KeyDataDir                DatabaseConfigKey = "data_dir"
	KeyDatabaseFile           DatabaseConfigKey = "database_file"
	KeyJournalMode            DatabaseConfigKey = "journal_mode"
	KeyBusyTimeoutMs          DatabaseConfigKey = "busy_timeout_ms"
	KeyMaxOpenConns           DatabaseConfigKey = "max_open_conns"
	KeyMaxIdleConns           DatabaseConfigKey = "max_idle_conns"
	KeyConnMaxLifetimeSeconds DatabaseConfigKey = "conn_max_lifetime_seconds"
)

// Backend defines the interface for secret storage and management backends
type Backend interface {
	// Initialize initializes the backend
	Initialize(ctx context.Context) error
	// Close closes the backend
	Close(ctx context.Context) error

	// StoreSecret stores a secret at the specified path
	StoreSecret(ctx context.Context, path string, secret store.Secret) error
	// LoadSecret loads a secret from the specified path
	LoadSecret(ctx context.Context, path string) (*store.Secret, error)
	// StoreAdminToken stores an admin token
	StoreAdminToken(ctx context.Context, token string) error
	// LoadAdminToken loads an admin token
	LoadAdminToken(ctx context.Context) (string, error)
}

// Config holds configuration for backend initialization
type Config struct {
	// Common configuration fields
	EncryptionKey string
	Location      string // Could be a file path, S3 bucket, etc.

	// Backend-specific configuration
	Options map[DatabaseConfigKey]any
}

// Factory creates a new backend instance
type Factory func(cfg Config) (Backend, error)

// registry of available backends
var backends = make(map[string]Factory)
