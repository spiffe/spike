//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	data2 "github.com/spiffe/spike-sdk-go/api/entity/data"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/entity/data"
	"github.com/spiffe/spike/pkg/store"
)

// Package provides an encrypted SQLite-based implementation of a data store
// backend. It supports storing and loading encrypted secrets and admin tokens
// with versioning support.

// DataStore implements the backend.Backend interface providing encrypted storage
// capabilities using SQLite as the underlying database. It uses AES-GCM for
// encryption and implements proper locking mechanisms for concurrent access.
type DataStore struct {
	db        *sql.DB      // Database connection handle
	cipher    cipher.AEAD  // Encryption cipher for data protection
	mu        sync.RWMutex // Mutex for thread-safe operations
	closeOnce sync.Once    // Ensures the database is closed only once
	opts      *Options     // Configuration options for the data store
}

// New creates a new DataStore instance with the provided configuration.
// It validates the encryption key and initializes the AES-GCM cipher.
//
// The encryption key must be 16, 24, or 32 bytes in length (for AES-128,
// AES-192, or AES-256 respectively).
//
// Returns an error if:
// - The options are invalid
// - The encryption key is malformed or has an invalid length
// - The cipher initialization fails
func New(cfg backend.Config) (backend.Backend, error) {
	opts, err := parseOptions(cfg.Options)
	if err != nil {
		return nil, fmt.Errorf("invalid sqlite options: %w", err)
	}

	key, err := hex.DecodeString(cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}

	// Validate key length
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf(
			"invalid encryption key length: must be 16, 24, or 32 bytes",
		)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &DataStore{
		cipher: gcm,
		opts:   opts,
	}, nil
}

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

	dbPath := filepath.Join(s.opts.DataDir, s.opts.DatabaseFile)

	// We don't need a username/password for SQLite.
	// Access to SQLite is controlled by regular filesystem permissions.
	db, err := sql.Open(
		"sqlite3",
		fmt.Sprintf("%s?_journal_mode=%s&_busy_timeout=%d",
			dbPath,
			s.opts.JournalMode,
			s.opts.BusyTimeoutMs))
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(s.opts.MaxOpenConns)
	db.SetMaxIdleConns(s.opts.MaxIdleConns)
	db.SetConnMaxLifetime(s.opts.ConnMaxLifetime)

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

// StoreSecret stores a secret at the specified path with its metadata and versions.
// It performs the following operations atomically within a transaction:
// - Updates the secret metadata (current version, creation time, update time)
// - Stores all secret versions with their respective data encrypted using AES-GCM
//
// The secret data is JSON-encoded before encryption.
//
// Returns an error if:
// - The transaction fails to begin or commit
// - Data marshaling fails
// - Encryption fails
// - Database operations fail
//
// This method is thread-safe.
func (s *DataStore) StoreSecret(
	ctx context.Context, path string, secret store.Secret,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false

	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				fmt.Printf("failed to rollback transaction: %v\n", err)
			}
		}
	}(tx)

	// Update metadata
	_, err = tx.ExecContext(ctx, queryUpdateSecretMetadata,
		path, secret.Metadata.CurrentVersion,
		secret.Metadata.CreatedTime, secret.Metadata.UpdatedTime)
	if err != nil {
		return fmt.Errorf("failed to store secret metadata: %w", err)
	}

	// Update versions
	for version, sv := range secret.Versions {
		data, err := json.Marshal(sv.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal secret values: %w", err)
		}

		encrypted, nonce, err := s.encrypt(data)
		if err != nil {
			return fmt.Errorf("failed to encrypt secret data: %w", err)
		}

		_, err = tx.ExecContext(ctx, queryUpsertSecret,
			path, version, nonce, encrypted, sv.CreatedTime, sv.DeletedTime)
		if err != nil {
			return fmt.Errorf("failed to store secret version: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true

	return nil
}

// LoadSecret retrieves a secret and all its versions from the specified path.
// It performs the following operations:
// - Loads the secret metadata
// - Retrieves all secret versions
// - Decrypts and unmarshals the version data
//
// Returns:
// - (nil, nil) if the secret doesn't exist
// - (nil, error) if any operation fails
// - (*store.Secret, nil) with the decrypted secret and all its versions on success
//
// This method is thread-safe.
func (s *DataStore) LoadSecret(
	ctx context.Context, path string,
) (*store.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var secret store.Secret

	// Load metadata
	err := s.db.QueryRowContext(ctx, querySecretMetadata, path).Scan(
		&secret.Metadata.CurrentVersion,
		&secret.Metadata.CreatedTime,
		&secret.Metadata.UpdatedTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load secret metadata: %w", err)
	}

	// Load versions
	rows, err := s.db.QueryContext(ctx, querySecretVersions, path)
	if err != nil {
		return nil, fmt.Errorf("failed to query secret versions: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	secret.Versions = make(map[int]store.Version)
	for rows.Next() {
		var (
			version     int
			nonce       []byte
			encrypted   []byte
			createdTime time.Time
			deletedTime sql.NullTime
		)

		if err := rows.Scan(
			&version, &nonce,
			&encrypted, &createdTime, &deletedTime,
		); err != nil {
			return nil, fmt.Errorf("failed to scan secret version: %w", err)
		}

		decrypted, err := s.decrypt(encrypted, nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secret version: %w", err)
		}

		var values map[string]string
		if err := json.Unmarshal(decrypted, &values); err != nil {
			return nil, fmt.Errorf("failed to unmarshal secret values: %w", err)
		}

		sv := store.Version{
			Data:        values,
			CreatedTime: createdTime,
		}
		if deletedTime.Valid {
			sv.DeletedTime = &deletedTime.Time
		}

		secret.Versions[version] = sv
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate secret versions: %w", err)
	}

	return &secret, nil
}

// StoreAdminToken encrypts and stores an admin token in the database.
// The token is encrypted using AES-GCM before storage.
//
// Returns an error if:
// - Encryption fails
// - Database operation fails
//
// This method is thread-safe.
func (s *DataStore) StoreAdminToken(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	encrypted, nonce, err := s.encrypt([]byte(token))
	if err != nil {
		return fmt.Errorf("failed to encrypt admin token: %w", err)
	}

	_, err = s.db.ExecContext(ctx, queryInsertAdminToken, nonce, encrypted)
	if err != nil {
		return fmt.Errorf("failed to store admin token: %w", err)
	}

	return nil
}

// LoadAdminSigningToken retrieves and decrypts the stored admin token.
//
// Returns:
// - ("", nil) if no token exists
// - ("", error) if loading or decryption fails
// - (token, nil) with the decrypted token on success
//
// This method is thread-safe.
func (s *DataStore) LoadAdminSigningToken(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nonce, encrypted []byte
	err := s.db.QueryRowContext(
		ctx, querySelectAdminSigningToken,
	).Scan(&nonce, &encrypted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to load admin token: %w", err)
	}

	decrypted, err := s.decrypt(encrypted, nonce)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt admin token: %w", err)
	}

	return string(decrypted), nil
}

// LoadAdminRecoveryMetadata retrieves the admin recovery metadata from the
// database. It returns an empty RecoveryMetadata struct if no record exists, or
// an error if the query fails.
//
// The method is thread-safe and uses a read lock when accessing the database.
// It queries the admin_recovery_metadata table for the single record with id=1,
// containing the token hash, encrypted root key, and salt used for admin
// recovery.
//
// Returns:
//   - RecoveryMetadata: The retrieved credentials containing recovery token
//     hash, encrypted root key, and salt
//   - error: nil on success, sql.ErrNoRows if no record exists, or wrapped
//     error on query failure
func (s *DataStore) LoadAdminRecoveryMetadata(ctx context.Context) (data.RecoveryMetadata, error) {
	// TODO: data namespace should come from SDK.

	s.mu.RLock()
	defer s.mu.RUnlock()

	const querySelectAdminCredentials = `
	SELECT token_hash, encrypted_root_key, salt 
	FROM admin_recovery_metadata
	WHERE id = 1`

	var creds data.RecoveryMetadata
	err := s.db.QueryRowContext(
		ctx, querySelectAdminCredentials,
	).Scan(&creds.RecoveryTokenHash, &creds.EncryptedRootKey, &creds.Salt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return data.RecoveryMetadata{}, nil
		}
		return data.RecoveryMetadata{}, fmt.Errorf("failed to load admin recovery metadata: %w", err)
	}

	return creds, nil
}

// StoreAdminRecoveryMetadata saves or updates the admin recovery metadata in
// the database. The operation is performed atomically within a serializable
// transaction.
//
// The method is thread-safe and uses a write lock when accessing the database.
// It uses REPLACE INTO to ensure only one record exists in the
// admin_recovery_metadata table with id=1. The record includes token hash,
// encrypted root key, salt, and creation timestamp.
//
// Parameters:
//   - ctx: Context for the database operation
//   - credentials: RecoveryMetadata containing the token hash, encrypted root
//     key, and salt to be stored
//
// Returns:
//   - error: nil on success, or wrapped error describing the failure
//     (transaction, query, or commit errors)
func (s *DataStore) StoreAdminRecoveryMetadata(
	ctx context.Context, credentials data.RecoveryMetadata,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx,
		&sql.TxOptions{Isolation: sql.LevelSerializable},
	)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false

	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				fmt.Printf("failed to rollback transaction: %v\n", err)
			}
		}
	}(tx)

	// Since there's only one admin recovery metadata record, we can use REPLACE INTO
	// or you could use the queryUpsertAdminCredentials constant
	_, err = tx.ExecContext(ctx, `
		REPLACE INTO admin_recovery_metadata (id, token_hash, encrypted_root_key, salt, created_at)
		VALUES (1, ?, ?, ?, CURRENT_TIMESTAMP)`,
		credentials.RecoveryTokenHash, credentials.EncryptedRootKey, credentials.Salt)
	if err != nil {
		return fmt.Errorf("failed to store admin recovery metadata: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true

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

func (s *DataStore) StorePolicy(ctx context.Context, policy data2.Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				fmt.Printf("failed to rollback transaction: %v\n", err)
			}
		}
	}(tx)

	_, err = tx.ExecContext(ctx, queryUpsertPolicy,
		policy.Id,
		policy.Name,
		policy.SpiffeIdPattern,
		policy.PathPattern,
		policy.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store policy: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true
	return nil
}

func (s *DataStore) DeletePolicy(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				fmt.Printf("failed to rollback transaction: %v\n", err)
			}
		}
	}(tx)

	_, err = tx.ExecContext(ctx, queryDeletePolicy, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true
	return nil
}

func (s *DataStore) LoadPolicy(ctx context.Context, id string) (*data2.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var policy data2.Policy
	policy.Id = id // Set the ID since we queried with it

	err := s.db.QueryRowContext(ctx, queryLoadPolicy, id).Scan(
		&policy.Name,
		&policy.SpiffeIdPattern,
		&policy.PathPattern,
		&policy.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	// Compile the patterns just like in CreatePolicy
	if policy.SpiffeIdPattern != "*" {
		idRegex, err := regexp.Compile(policy.SpiffeIdPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid spiffeid pattern: %w", err)
		}
		policy.IdRegex = idRegex
	}

	if policy.PathPattern != "*" {
		pathRegex, err := regexp.Compile(policy.PathPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid path pattern: %w", err)
		}
		policy.PathRegex = pathRegex
	}

	return &policy, nil
}
