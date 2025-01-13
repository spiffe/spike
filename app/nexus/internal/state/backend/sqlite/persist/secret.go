//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
	"github.com/spiffe/spike/pkg/store"
)

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
	_, err = tx.ExecContext(ctx, ddl.QueryUpdateSecretMetadata,
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

		_, err = tx.ExecContext(ctx, ddl.QueryUpsertSecret,
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
	err := s.db.QueryRowContext(ctx, ddl.QuerySecretMetadata, path).Scan(
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
	rows, err := s.db.QueryContext(ctx, ddl.QuerySecretVersions, path)
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
