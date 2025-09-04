//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

// StoreSecret stores a secret at the specified path with its metadata and
// versions. It performs the following operations atomically within a
// transaction:
//   - Updates the secret metadata (current version, creation time, update time)
//   - Stores all secret versions with their respective data encrypted using
//     AES-GCM
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
	ctx context.Context, path string, secret kv.Value,
) error {
	const fName = "StoreSecret"
	if ctx == nil {
		log.FatalLn(fName, "message", "nil context")
	}

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
		path, secret.Metadata.CurrentVersion, secret.Metadata.OldestVersion,
		secret.Metadata.CreatedTime, secret.Metadata.UpdatedTime, secret.Metadata.MaxVersions)
	if err != nil {
		return fmt.Errorf("failed to kv secret metadata: %w", err)
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
			return fmt.Errorf("failed to kv secret version: %w", err)
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
//   - (nil, nil) if the secret doesn't exist
//   - (nil, error) if any operation fails
//   - (*kv.Secret, nil) with the decrypted secret and all its versions on
//     success
//
// This method is thread-safe.
func (s *DataStore) LoadSecret(
	ctx context.Context, path string,
) (*kv.Value, error) {
	const fName = "LoadSecret"
	if ctx == nil {
		log.FatalLn(fName, "message", "nil context")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.loadSecretInternal(ctx, path)
}

// LoadAllSecrets retrieves all secrets from the database.
// It returns a map where the keys are secret paths and the values are the
// corresponding secrets.
// Each secret includes its metadata and all versions with decrypted data.
// If an error occurs during the retrieval process, it returns nil and the
// error. This method acquires a read lock to ensure consistent access to the
// database.
//
// Contexts that are canceled or reach their deadline will result in the
// operation being interrupted early and returning an error.
//
// Example usage:
//
//	secrets, err := dataStore.LoadAllSecrets(context.Background())
//	if err != nil {
//	    log.Fatalf("Failed to load secrets: %v", err)
//	}
//	for path, secret := range secrets {
//	    fmt.Printf("Secret at path %s has %d versions\n", path,
//	      len(secret.Versions))
//	}
func (s *DataStore) LoadAllSecrets(
	ctx context.Context,
) (map[string]*kv.Value, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all secret paths
	rows, err := s.db.QueryContext(ctx, ddl.QueryPathsFromMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to query secret paths: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	// Map to store all secrets
	secrets := make(map[string]*kv.Value)

	// Iterate over paths
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("failed to scan path: %w", err)
		}

		// Load the full secret for this path
		secret, err := s.loadSecretInternal(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("failed to load secret at path %s: %w", path, err)
		}

		if secret != nil {
			secrets[path] = secret
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate secret paths: %w", err)
	}

	return secrets, nil
}
