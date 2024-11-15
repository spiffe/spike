//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"github.com/spiffe/spike/app/nexus/internal/config"
	"sync"
	"time"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/memory"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite"
	"github.com/spiffe/spike/app/nexus/internal/state/store"
	"github.com/spiffe/spike/internal/log"
)

var (
	memoryBackend *memory.NoopStore
	sqliteBackend *sqlite.DataStore
	backendMu     sync.RWMutex
)

// AsyncPersistSecret asynchronously stores a secret from the KV store to the
// backend cache. It retrieves the secret from the provided path and attempts to
// cache it in a background goroutine. If the backend is not available or if the
// store operation fails, it will only log a warning since the KV store remains
// the source of truth.
//
// Parameters:
//   - kv: A pointer to the KV store containing the secrets
//   - path: The path where the secret is stored in the KV store
//
// The function does not return any errors since it handles them internally
// through logging. Cache failures are non-fatal as the KV store is considered
// the authoritative data source.
func AsyncPersistSecret(kv *store.KV, path string) {
	be := Backend()

	// Get the full secret for caching
	secret := kv.GetRawSecret(path)
	if secret != nil {
		go func() {
			if be == nil {
				return // No cache available
			}

			ctx, cancel := context.WithTimeout(
				context.Background(),
				env.DatabaseOperationTimeout(),
			)
			defer cancel()

			if err := be.StoreSecret(ctx, path, *secret); err != nil {
				// Log error but continue - memory is source of truth
				log.Log().Warn("asyncPersistSecret",
					"msg", "Failed to cache secret",
					"path", path,
					"err", err.Error(),
				)
			}
		}()
	}
}

// ReadSecret attempts to retrieve a secret from the backend cache at the
// specified path and version. If no specific version is provided (version = 0),
// it uses the secret's current version.
//
// Parameters:
//   - path: The path where the secret is stored in the cache
//   - version: The specific version of the secret to retrieve.
//     If 0, uses the current version
//
// Returns:
//   - A pointer to the store.Secret if found and not deleted, nil otherwise
//
// The function returns nil in several cases:
//   - When the backend is not available
//   - When there's an error loading from the cache
//   - When the requested version doesn't exist
//   - When the secret version has been deleted (DeletedTime is set)
func ReadSecret(path string, version int) *store.Secret {
	be := Backend()
	if be == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cachedSecret, err := be.LoadSecret(ctx, path)
	if err != nil {
		log.Log().Warn("readSecret",
			"msg", "Failed to load secret from cache",
			"path", path,
			"err", err.Error(),
		)
		return nil
	}

	if cachedSecret != nil {
		if version == 0 {
			version = cachedSecret.Metadata.CurrentVersion
		}

		if sv, ok := cachedSecret.Versions[version]; ok && sv.DeletedTime == nil {
			return cachedSecret
		}
	}

	return nil
}

// InitializeSqliteBackend creates and initializes a SQLite backend instance
// using the provided root key for encryption. The backend is configured using
// environment variables for database settings such as directory location,
// connection limits, and journal mode.
//
// Parameters:
//   - rootKey: The encryption key used to secure the SQLite database
//
// Returns:
//   - A backend.Backend interface if successful, nil if initialization fails
//
// The function attempts two main operations:
//  1. Creating the SQLite backend with the provided configuration
//  2. Initializing the backend with a 30-second timeout
//
// If either operation fails, it logs a warning and returns nil. This allows
// the system to continue operating with in-memory state only. Configuration
// options include:
//   - Database directory and filename
//   - Journal mode settings
//   - Connection pool settings (max open, max idle, lifetime)
//   - Busy timeout settings
func InitializeSqliteBackend(rootKey string) backend.Backend {
	opts := map[backend.DatabaseConfigKey]any{}

	opts[backend.KeyDataDir] = env.DatabaseDir()
	opts[backend.KeyDatabaseFile] = "spike.db"
	opts[backend.KeyJournalMode] = env.DatabaseJournalMode()
	opts[backend.KeyBusyTimeoutMs] = env.DatabaseBusyTimeoutMs()
	opts[backend.KeyMaxOpenConns] = env.DatabaseMaxOpenConns()
	opts[backend.KeyMaxIdleConns] = env.DatabaseMaxIdleConns()
	opts[backend.KeyConnMaxLifetimeSeconds] = env.DatabaseConnMaxLifetimeSec()

	// Create SQLite backend configuration
	cfg := backend.Config{
		// Use the root key as the encryption key
		EncryptionKey: rootKey,
		Options:       opts,
	}

	// Initialize SQLite backend
	dbBackend, err := sqlite.New(cfg)
	if err != nil {
		// Log error but don't fail initialization
		// The system can still work with just in-memory state
		log.Log().Warn("initializeSqliteBackend",
			"msg", "Failed to create SQLite backend",
			"err", err.Error(),
		)
		return nil
	}

	ctxC, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := dbBackend.Initialize(ctxC); err != nil {
		log.Log().Warn("initializeSqliteBackend",
			"msg", "Failed to initialize SQLite backend",
			"err", err.Error(),
		)
		return nil
	}

	return dbBackend
}

// ReadAdminToken attempts to retrieve the admin token from the backend cache.
// It uses a 5-second timeout for the operation.
//
// Returns:
//   - The cached admin token as a string if successful
//   - An empty string if:
//   - The backend is not available
//   - The load operation fails
//   - No token is found in the cache
//
// Errors during the load operation are logged but not returned, as the system
// considers the in-memory state to be the source of truth.
func ReadAdminToken() string {
	be := Backend()
	if be == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cachedToken, err := be.LoadAdminToken(ctx)
	if err != nil {
		// Log error but continue - memory is source of truth
		log.Log().Warn("readAdminToken",
			"msg", "Failed to load admin token from cache",
			"err", err.Error(),
		)
		return ""
	}

	return cachedToken
}

// AsyncPersistAdminToken asynchronously stores the provided admin token in the
// backend cache. The operation is performed in a background goroutine with a
// 5-second timeout.
//
// Parameters:
//   - token: The admin token to be cached
//
// The function returns immediately while the storage operation continues
// asynchronously. Any errors during the storage operation are logged but not
// returned, as the in-memory token remains the source of truth. If the backend
// is not available, the goroutine terminates without attempting storage.
func AsyncPersistAdminToken(token string) {
	go func() {
		be := Backend()
		if be == nil {
			return // No cache available
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			config.SpikeNexusAdminTokenPersistTimeoutSecs*time.Second,
		)
		defer cancel()

		if err := be.StoreAdminToken(ctx, token); err != nil {
			// Log error but continue - memory is source of truth
			log.Log().Warn("asyncPersistAdminToken",
				"msg", "Failed to cache admin token",
				"err", err.Error(),
			)
		}
	}()
}

// InitializeBackend creates and returns a backend storage implementation based
// on the configured store type in the environment. The function is thread-safe
// through a mutex lock.
//
// Parameters:
//   - rootKey: The encryption key used for backend initialization (used by
//     SQLite backend)
//
// Returns:
//   - A backend.Backend interface implementation:
//   - memory.NoopStore for 'memory' store type or unknown types
//   - SQLite backend for 'sqlite' store type
//
// The actual backend type is determined by env.BackendStoreType():
//   - env.Memory: Returns a no-op memory store
//   - env.Sqlite: Initializes and returns a SQLite backend
//   - default: Falls back to a no-op memory store
//
// The function is safe for concurrent access as it uses a mutex to protect the
// initialization process.
func InitializeBackend(rootKey string) backend.Backend {
	backendMu.Lock()
	defer backendMu.Unlock()

	storeType := env.BackendStoreType()

	switch storeType {
	case env.Memory:
		return &memory.NoopStore{}
	case env.Sqlite:
		return InitializeSqliteBackend(rootKey)
	default:
		return &memory.NoopStore{}
	}
}

// Backend returns the currently initialized backend storage instance. The
// function is thread-safe through a read mutex lock.
//
// Returns:
//   - A backend.Backend interface pointing to the current backend instance:
//   - memoryBackend for 'memory' store type or unknown types
//   - sqliteBackend for 'sqlite' store type
//
// The return value is determined by env.BackendStoreType():
//   - env.Memory: Returns the memory backend instance
//   - env.Sqlite: Returns the SQLite backend instance
//   - default: Falls back to the memory backend instance
//
// The function is safe for concurrent access as it uses a read mutex to protect
// access to the backend instances. Unlike InitializeBackend, this function
// returns existing instances rather than creating new ones.
func Backend() backend.Backend {
	backendMu.RLock()
	defer backendMu.RUnlock()

	storeType := env.BackendStoreType()
	switch storeType {
	case env.Memory:
		return memoryBackend
	case env.Sqlite:
		return sqliteBackend
	default:
		return memoryBackend
	}
}
