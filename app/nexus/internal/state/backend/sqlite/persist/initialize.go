//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike-sdk-go/validation"
)

// Initialize prepares the DataStore for use by creating the data directory,
// opening the SQLite database connection, configuring connection pool
// settings, and creating required database tables.
//
// The initialization process follows these steps:
//   - Validates that the backend is not already initialized
//   - Creates the data directory if it does not exist
//   - Opens a SQLite database connection with the configured journal mode
//     and busy timeout
//   - Configures connection pool settings (max open/idle connections and
//     connection lifetime)
//   - Creates database tables unless SPIKE_DATABASE_SKIP_SCHEMA_CREATION
//     is set
//
// Parameters:
//   - ctx: Context for managing request lifetime and cancellation.
//
// Returns:
//   - *sdkErrors.SDKError: An error if the backend is already initialized,
//     the data directory creation fails, the database connection fails, or
//     table creation fails. Returns nil on success.
//
// This method is thread-safe and uses a mutex to prevent concurrent
// initialization attempts.
func (s *DataStore) Initialize(ctx context.Context) *sdkErrors.SDKError {
	const fName = "Initialize"

	validation.NonNilContextOrDie(ctx, fName)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return sdkErrors.ErrStateAlreadyInitialized
	}

	if err := s.createDataDir(); err != nil {
		failErr := sdkErrors.ErrFSDirectoryCreationFailed.Wrap(err)
		return failErr
	}

	dbPath := filepath.Join(s.Opts.DataDir, s.Opts.DatabaseFile)

	// We don't need a username/password for SQLite.
	// Access to SQLite is controlled by regular filesystem permissions.
	db, err := sql.Open(
		"sqlite3",
		fmt.Sprintf("%s?_journal_mode=%s&_busy_timeout=%d",
			dbPath, s.Opts.JournalMode, s.Opts.BusyTimeoutMs),
	)
	if err != nil {
		failErr := sdkErrors.ErrFSFileOpenFailed.Wrap(err)
		return failErr
	}

	// Set connection pool settings
	db.SetMaxOpenConns(s.Opts.MaxOpenConns)
	db.SetMaxIdleConns(s.Opts.MaxIdleConns)
	db.SetConnMaxLifetime(s.Opts.ConnMaxLifetime)

	// Use the existing database if the schema is not to be created.
	if env.DatabaseSkipSchemaCreationVal() {
		s.db = db
		return nil
	}

	// Create tables
	if createErr := s.createTables(ctx, db); createErr != nil {
		closeErr := db.Close()
		if closeErr != nil {
			return createErr.Wrap(closeErr)
		}
		return createErr
	}

	s.db = db
	return nil
}

// Close safely closes the database connection. It ensures the database is
// closed only once, even if called multiple times, by using sync.Once.
//
// Parameters:
//   - ctx: Context parameter (currently unused but maintained for interface
//     compatibility).
//
// Returns:
//   - *sdkErrors.SDKError: An error if closing the database connection
//     fails, wrapped in ErrFSFileCloseFailed. Returns nil on success.
//     Later calls always return nil since the close operation only
//     executes once.
//
// This method is thread-safe.
func (s *DataStore) Close(_ context.Context) *sdkErrors.SDKError {
	var err error
	s.closeOnce.Do(func() {
		err = s.db.Close()
	})
	if err != nil {
		return sdkErrors.ErrStoreCloseFailed.Wrap(err)
	}
	return nil
}
