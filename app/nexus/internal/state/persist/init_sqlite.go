//	  \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//	\\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"encoding/hex"

	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite"
	"github.com/spiffe/spike/internal/config"
)

// initializeSqliteBackend creates and initializes an SQLite backend instance
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
func initializeSqliteBackend(rootKey *[32]byte) backend.Backend {
	const fName = "initializeSqliteBackend"
	const dbName = "spike.db"

	opts := map[backend.DatabaseConfigKey]any{}

	opts[backend.KeyDataDir] = config.NexusDataFolder()
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
