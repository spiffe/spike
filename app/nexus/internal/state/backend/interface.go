//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package backend defines the storage interface for SPIKE Nexus.
//
// This package provides the Backend interface that all storage implementations
// must satisfy. SPIKE Nexus uses this interface to persist secrets and policies
// with encryption at rest.
//
// Available implementations:
//   - sqlite: Persistent encrypted storage using SQLite (production use)
//   - memory: In-memory storage for development and testing
//   - noop: No-op implementation for embedding in other backends
//   - lite: Encryption-only backend (embeds noop, provides cipher for
//     encryption-as-a-service)
//
// The Backend interface provides:
//   - Secret storage with versioning and soft-delete support
//   - Policy storage for SPIFFE ID and path-based access control
//   - Cipher access for encryption-as-a-service endpoints
//   - Lifecycle management (Initialize/Close)
//
// All implementations must be thread-safe. Secrets and policies are encrypted
// using AES-256-GCM before storage.
package backend

import (
	"context"
	"crypto/cipher"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
)

// DatabaseConfigKey represents a configuration key for database-specific
// backend options. These keys are used in the Config.Options map to provide
// backend-specific configuration values.
type DatabaseConfigKey string

// Database configuration keys for backend initialization.
const (
	// KeyDataDir specifies the directory path where database files are stored.
	KeyDataDir DatabaseConfigKey = "data_dir"

	// KeyDatabaseFile specifies the name of the database file.
	KeyDatabaseFile DatabaseConfigKey = "database_file"

	// KeyJournalMode specifies the SQLite journal mode (e.g., WAL, DELETE).
	KeyJournalMode DatabaseConfigKey = "journal_mode"

	// KeyBusyTimeoutMs specifies the timeout in milliseconds for busy
	// database connections.
	KeyBusyTimeoutMs DatabaseConfigKey = "busy_timeout_ms"

	// KeyMaxOpenConns specifies the maximum number of open database
	// connections.
	KeyMaxOpenConns DatabaseConfigKey = "max_open_conns"

	// KeyMaxIdleConns specifies the maximum number of idle database
	// connections in the pool.
	KeyMaxIdleConns DatabaseConfigKey = "max_idle_conns"

	// KeyConnMaxLifetimeSeconds specifies the maximum lifetime of
	// a connection in seconds.
	KeyConnMaxLifetimeSeconds DatabaseConfigKey = "conn_max_lifetime_seconds"
)

// Backend defines the interface for secret and policy storage backends.
// Implementations must provide thread-safe operations for storing, retrieving,
// and managing secrets and policies, along with lifecycle management methods
// for initialization and cleanup.
type Backend interface {
	// Initialize prepares the backend for use by setting up the necessary
	// resources such as database connections, file system directories, or
	// network connections. It must be called before any other operations.
	//
	// Parameters:
	//   - ctx: Context for managing request lifetime and cancellation.
	//
	// Returns:
	//   - *sdkErrors.SDKError: An error if initialization fails. Returns nil
	//     on success.
	Initialize(ctx context.Context) *sdkErrors.SDKError

	// Close releases all resources held by the backend, including database
	// connections, file handles, or network connections. It should be called
	// when the backend is no longer needed.
	//
	// Parameters:
	//   - ctx: Context for managing request lifetime and cancellation.
	//
	// Returns:
	//   - *sdkErrors.SDKError: An error if cleanup fails. Returns nil on
	//     success.
	Close(ctx context.Context) *sdkErrors.SDKError

	// StoreSecret persists a secret at the specified path. The secret is
	// encrypted before storage. If a secret already exists at the path, it
	// is overwritten.
	//
	// Parameters:
	//   - ctx: Context for managing request lifetime and cancellation.
	//   - path: The namespace path where the secret will be stored.
	//   - secret: The secret value to store.
	//
	// Returns:
	//   - *sdkErrors.SDKError: An error if the operation fails. Returns nil
	//     on success.
	StoreSecret(
		ctx context.Context, path string, secret kv.Value,
	) *sdkErrors.SDKError

	// LoadSecret retrieves a secret from the specified path. The secret is
	// automatically decrypted before being returned.
	//
	// The returned secret may have Metadata.CurrentVersion == 0, which
	// indicates that all versions have been deleted (no valid version
	// exists).
	//
	// Parameters:
	//   - ctx: Context for managing request lifetime and cancellation.
	//   - path: The namespace path of the secret to retrieve.
	//
	// Returns:
	//   - *kv.Value: The decrypted secret value, or nil if not found.
	//   - *sdkErrors.SDKError: An error if the operation fails. Returns nil
	//     on success.
	LoadSecret(ctx context.Context, path string) (*kv.Value, *sdkErrors.SDKError)

	// LoadAllSecrets retrieves all secrets stored in the backend. Secrets
	// are automatically decrypted before being returned.
	//
	// Returned secrets may have Metadata.CurrentVersion == 0, which indicates
	// that all versions have been deleted (no valid version exists).
	//
	// Parameters:
	//   - ctx: Context for managing request lifetime and cancellation.
	//
	// Returns:
	//   - map[string]*kv.Value: A map of secret paths to their decrypted
	//     values.
	//   - *sdkErrors.SDKError: An error if the operation fails. Returns nil
	//     on success.
	LoadAllSecrets(
		ctx context.Context,
	) (map[string]*kv.Value, *sdkErrors.SDKError)

	// StorePolicy persists a policy object in the backend storage. If a
	// policy with the same ID already exists, it is overwritten.
	//
	// Parameters:
	//   - ctx: Context for managing request lifetime and cancellation.
	//   - policy: The policy object to store.
	//
	// Returns:
	//   - *sdkErrors.SDKError: An error if the operation fails. Returns nil
	//     on success.
	StorePolicy(ctx context.Context, policy data.Policy) *sdkErrors.SDKError

	// LoadPolicy retrieves a policy by its ID from the backend storage.
	//
	// Parameters:
	//   - ctx: Context for managing request lifetime and cancellation.
	//   - id: The unique identifier of the policy to retrieve.
	//
	// Returns:
	//   - *data.Policy: The policy object, or nil if not found.
	//   - *sdkErrors.SDKError: An error if the operation fails. Returns nil
	//     on success.
	LoadPolicy(
		ctx context.Context, id string,
	) (*data.Policy, *sdkErrors.SDKError)

	// LoadAllPolicies retrieves all policies stored in the backend.
	//
	// Parameters:
	//   - ctx: Context for managing request lifetime and cancellation.
	//
	// Returns:
	//   - map[string]*data.Policy: A map of policy IDs to their policy
	//     objects.
	//   - *sdkErrors.SDKError: An error if the operation fails. Returns nil
	//     on success.
	LoadAllPolicies(
		ctx context.Context,
	) (map[string]*data.Policy, *sdkErrors.SDKError)

	// DeletePolicy removes a policy object identified by the given ID from
	// the backend storage.
	//
	// Parameters:
	//   - ctx: Context for managing request lifetime and cancellation.
	//   - id: The unique identifier of the policy to delete.
	//
	// Returns:
	//   - *sdkErrors.SDKError: An error if the operation fails. Returns nil
	//     on success.
	DeletePolicy(ctx context.Context, id string) *sdkErrors.SDKError

	// GetCipher retrieves the AEAD cipher used for encryption and decryption.
	//
	// INTENDED USAGE:
	//   - Cipher API routes (v1/cipher/encrypt, v1/cipher/decrypt) for
	//     encryption-as-a-service functionality
	//   - Backend-internal encryption operations (via private
	//     encrypt/decrypt methods)
	//
	// IMPORTANT: Other API routes (secrets, policies) should NOT use this method
	// directly. They should rely on the backend's own StoreSecret/LoadSecret
	// methods which handle encryption internally, because Backend implementations
	// may return nil if cipher access is not appropriate for their specific use
	// case.
	GetCipher() cipher.AEAD
}

// Config holds configuration parameters for backend initialization. It
// provides both common settings applicable to all backend types and
// backend-specific options.
type Config struct {
	// EncryptionKey is the key used for encrypting and decrypting secrets.
	// Must be 32 bytes (256 bits) for AES-256 encryption.
	EncryptionKey string

	// Location specifies the storage location for the backend. The
	// interpretation depends on the backend type (e.g., file path for
	// SQLite, S3 bucket for cloud storage).
	Location string

	// Options contains backend-specific configuration parameters. The keys
	// and values depend on the backend implementation. For database
	// backends, use DatabaseConfigKey constants as keys.
	Options map[DatabaseConfigKey]any
}

// Factory is a function type that creates and returns a new backend instance
// configured with the provided settings.
//
// Parameters:
//   - cfg: Configuration parameters for the backend.
//
// Returns:
//   - Backend: The initialized backend instance.
//   - *sdkErrors.SDKError: An error if backend creation fails. Returns nil
//     on success.
type Factory func(cfg Config) (Backend, *sdkErrors.SDKError)
