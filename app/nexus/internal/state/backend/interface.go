//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package backend provides the interfaces and configurations necessary for
// implementing a secure and flexible storage backend for managing secrets and
// policies. It includes definitions for interactions like initializing
// backends, storing, retrieving, and deleting secrets and policies, as well as
// abstractions for backend configuration and factory creation.
//
// The backend package is designed to be extensible, allowing implementation of
// various storage mechanisms such as file-based, SQL databases, or cloud-based
// solutions.
package backend

import (
	"context"
	"crypto/cipher"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/kv"
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
	StoreSecret(ctx context.Context, path string, secret kv.Value) error
	// LoadSecret loads a secret from the specified path
	LoadSecret(ctx context.Context, path string) (*kv.Value, error)
	// LoadAllSecrets retrieves all secrets stored in the backend.
	// Returns a map of secret paths to their values or an error.
	LoadAllSecrets(ctx context.Context) (map[string]*kv.Value, error)

	// StorePolicy stores a policy object in the backend storage.
	StorePolicy(ctx context.Context, policy data.Policy) error

	// LoadPolicy retrieves a policy by its ID from the backend storage.
	// It returns the policy object and an error, if any.
	LoadPolicy(ctx context.Context, id string) (*data.Policy, error)

	// LoadAllPolicies retrieves all policies stored in the backend.
	// Returns a map of policy IDs to their policy objects or an error.
	LoadAllPolicies(ctx context.Context) (map[string]*data.Policy, error)

	// DeletePolicy removes a policy object identified by the given ID from
	// storage.
	// - `ctx` is the context for managing cancellations and timeouts.
	// - `id` is the identifier of the policy to delete.
	// Returns an error if the operation fails.
	DeletePolicy(ctx context.Context, id string) error

	// GetCipher retrieves the AEAD cipher used for encryption and decryption.
	GetCipher() cipher.AEAD
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
