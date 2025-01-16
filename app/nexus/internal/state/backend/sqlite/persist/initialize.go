//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
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
func (s *DataStore) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return errors.New("backend already initialized")
	}

	if err := s.createDataDir(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(s.Opts.DataDir, s.Opts.DatabaseFile)

	// We don't need a username/password for SQLite.
	// Access to SQLite is controlled by regular filesystem permissions.
	db, err := sql.Open(
		"sqlite3",
		fmt.Sprintf("%s?_journal_mode=%s&_busy_timeout=%d",
			dbPath,
			s.Opts.JournalMode,
			s.Opts.BusyTimeoutMs))
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(s.Opts.MaxOpenConns)
	db.SetMaxIdleConns(s.Opts.MaxIdleConns)
	db.SetConnMaxLifetime(s.Opts.ConnMaxLifetime)

	// Create tables
	if err := s.createTables(ctx, db); err != nil {
		err := db.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to create tables: %w", err)
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
