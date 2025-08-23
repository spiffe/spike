//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"encoding/hex"

	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/lite"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/memory"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite"
	"github.com/spiffe/spike/internal/config"
)

// InitializeSqliteBackend creates and initializes an SQLite backend instance
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
// the system to continue operating with an in-memory state only. Configuration
// options include:
//   - Database directory and filename
//   - Journal mode settings
//   - Connection pool settings (max open, max idle, lifetime)
//   - Busy timeout settings
func InitializeSqliteBackend(rootKey *[32]byte) backend.Backend {
	const fName = "initializeSqliteBackend"
	const dbName = "spike.db"

	opts := map[backend.DatabaseConfigKey]any{}

	opts[backend.KeyDataDir] = config.SpikeNexusDataFolder()
	opts[backend.KeyDatabaseFile] = dbName
	opts[backend.KeyJournalMode] = env.DatabaseJournalMode()
	opts[backend.KeyBusyTimeoutMs] = env.DatabaseBusyTimeoutMs()
	opts[backend.KeyMaxOpenConns] = env.DatabaseMaxOpenConns()
	opts[backend.KeyMaxIdleConns] = env.DatabaseMaxIdleConns()
	opts[backend.KeyConnMaxLifetimeSeconds] = env.DatabaseConnMaxLifetimeSec()

	// Create SQLite backend configuration
	cfg := backend.Config{
		// Use a copy of the root key as the encryption key.
		// The caller will securely zero out the original root key.
		EncryptionKey: hex.EncodeToString(rootKey[:]),
		Options:       opts,
	}

	// Initialize SQLite backend
	dbBackend, err := sqlite.New(cfg)
	if err != nil {
		// Log error but don't fail initialization
		// The system can still work with just in-memory state
		log.Log().Warn(fName,
			"message", "Failed to create SQLite backend",
			"err", err.Error(),
		)
		return nil
	}

	ctxC, cancel := context.WithTimeout(
		context.Background(), env.DatabaseInitializationTimeout(),
	)
	defer cancel()

	if err := dbBackend.Initialize(ctxC); err != nil {
		log.Log().Warn(fName,
			"message", "Failed to initialize SQLite backend",
			"err", err.Error(),
		)
		return nil
	}

	return dbBackend
}

const shardSize = 32

// InitializeLiteBackend creates and initializes a Lite backend instance
// using the provided root key for encryption. The Lite backend is a
// lightweight alternative to SQLite for persistent storage. The Lite mode
// does not use any backing store and relies on persisting encrypted data
// on object storage (like S3, or Minio).
//
// Parameters:
//   - rootKey: A 32-byte encryption key used to secure the Lite database.
//     The backend will use this key directly for encryption operations.
//
// Returns:
//   - A backend.Backend interface implementation if successful
//   - nil if initialization fails
//
// Error Handling:
// If the backend creation fails, the function logs a warning and returns nil
// rather than propagating the error. This allows the system to gracefully
// degrade to using only in-memory state without blocking startup.
//
// Example:
//
//	var rootKey [32]byte
//	// ... populate rootKey with secure random data ...
//	backend := InitializeLiteBackend(&rootKey)
//	if backend == nil {
//	    // Handle fallback to in-memory only operation
//	}
//
// Note: Unlike the SQLite backend, the Lite backend does not require a
// separate Initialize() call or timeout configuration.
func InitializeLiteBackend(rootKey *[shardSize]byte) backend.Backend {
	const fName = "initializeLiteBackend"
	dbBackend, err := lite.New(rootKey)
	if err != nil {
		// Log error but don't fail initialization
		// The system can still work with just in-memory state
		log.Log().Warn(fName,
			"message", "Failed to create SQLite backend",
			"err", err.Error(),
		)
		return nil
	}
	return dbBackend
}

var be backend.Backend

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
//
// Note: This function modifies the package-level be variable. Subsequent calls
// will reinitialize the backend, potentially losing any existing state.
func InitializeBackend(rootKey *[shardSize]byte) {
	const fName = "initializeBackend"

	log.Log().Info(fName,
		"message", "Initializing backend", "storeType", env.BackendStoreType())

	backendMu.Lock()
	defer backendMu.Unlock()

	storeType := env.BackendStoreType()

	switch storeType {
	case env.Lite:
		be = InitializeLiteBackend(rootKey)
	case env.Memory:
		be = &memory.InMemoryStore{}
	case env.Sqlite:
		be = InitializeSqliteBackend(rootKey)
	default:
		be = &memory.InMemoryStore{}
	}

	log.Log().Info(fName, "message", "Backend initialized", "storeType", storeType)
}
