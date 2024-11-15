//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"fmt"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite"

	"github.com/spiffe/spike/app/nexus/internal/state/store"
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
	Options map[sqlite.DatabaseConfigKey]any
}

// Factory creates a new backend instance
type Factory func(cfg Config) (Backend, error)

// registry of available backends
var backends = make(map[string]Factory)

// Register adds a new backend to the registry
func Register(name string, factory Factory) {
	backends[name] = factory
}

// New creates a new backend of the specified type
func New(backendType string, cfg Config) (Backend, error) {
	factory, exists := backends[backendType]
	if !exists {
		return nil, fmt.Errorf("unknown backend type: %s", backendType)
	}
	return factory(cfg)
}
