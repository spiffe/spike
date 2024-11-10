//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"fmt"
	"github.com/spiffe/spike/app/nexus/internal/state/store"
)

// Backend defines the interface for secret storage and management backends
type Backend interface {
	Initialize(ctx context.Context) error
	Close(ctx context.Context) error

	StoreSecret(ctx context.Context, path string, secret store.Secret) error
	LoadSecret(ctx context.Context, path string) (*store.Secret, error)
	StoreAdminToken(ctx context.Context, token string) error
	LoadAdminToken(ctx context.Context) (string, error)
}

// Operation represents types of operations that can be performed
type Operation string

const (
	OpRead   Operation = "read"
	OpWrite  Operation = "write"
	OpDelete Operation = "delete"
	OpList   Operation = "list"
)

// Config holds configuration for backend initialization
type Config struct {
	// Common configuration fields
	EncryptionKey string
	Location      string // Could be a file path, S3 bucket, etc.

	// Backend-specific configuration
	Options map[string]any
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
