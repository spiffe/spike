//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"crypto/cipher"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spiffe/spike-sdk-go/config/env"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// Initialize prepares the DataStore for use by:
// - Creating the necessary data directory
// - Opening the SQLite database connection
// - Configuring connection pool settings
// - Creating required database tables
//
// It returns an error if:
// - The backend is already initialized
// - The data directory creation fails
// - The database connection fails
// - Table creation fails
//
// This method is thread-safe.
func (s *DataStore) Initialize(ctx context.Context) *sdkErrors.SDKError {
	const fName = "Initialize"
	if ctx == nil {
		log.FatalLn(fName, "message", sdkErrors.ErrCodeNilContext)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return sdkErrors.ErrAlreadyInitialized
	}

	if err := s.createDataDir(); err != nil {
		failErr := sdkErrors.ErrDirectoryCreationFailed
		return errors.Join(failErr, err)
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
		failErr := sdkErrors.ErrFileOpenFailed
		return errors.Join(failErr, err)
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
	if err := s.createTables(ctx, db); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			return closeErr
		}
		failErr := sdkErrors.ErrCreationFailed
		return errors.Join(failErr, err)
	}

	s.db = db
	return nil
}

// Close safely closes the database connection.
// It ensures the database is closed only once even if called multiple times.
//
// Returns any error encountered while closing the database connection.
func (s *DataStore) Close(_ context.Context) error {
	var err error
	s.closeOnce.Do(func() {
		err = s.db.Close()
	})
	return err
}

// GetCipher retrieves the AEAD cipher instance from the DataStore.
func (s *DataStore) GetCipher() cipher.AEAD {
	return s.Cipher
}
