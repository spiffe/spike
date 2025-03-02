//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"time"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/memory"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

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
	const fName = "initializeSqliteBackend"

	opts := map[backend.DatabaseConfigKey]any{}

	opts[backend.KeyDataDir] = config.SpikeNexusDataFolder()
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
		log.Log().Warn(fName,
			"msg", "Failed to create SQLite backend",
			"err", err.Error(),
		)
		return nil
	}

	ctxC, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := dbBackend.Initialize(ctxC); err != nil {
		log.Log().Warn(fName,
			"msg", "Failed to initialize SQLite backend",
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
func InitializeBackend(rootKey string) {
	const fName = "initializeBackend"

	log.Log().Info(fName,
		"msg", "Initializing backend", "storeType", env.BackendStoreType())

	backendMu.Lock()
	defer backendMu.Unlock()

	storeType := env.BackendStoreType()

	switch storeType {
	case env.Memory:
		be = &memory.NoopStore{}
	case env.Sqlite:
		be = InitializeSqliteBackend(rootKey)
	default:
		be = &memory.NoopStore{}
	}

	log.Log().Info(fName, "msg", "Backend initialized", "storeType", storeType)
}
