//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
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
		failErr := sdkErrors.ErrNilContext
		log.FatalErr(fName, *failErr)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		failErr := sdkErrors.ErrTransactionBeginFailed
		return errors.Join(failErr, err)
	}

	committed := false

	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				failErr := sdkErrors.ErrTransactionRollbackFailed
				log.Log().Warn(fName, "message", failErr.Error())
			}
		}
	}(tx)

	// Update metadata
	_, err = tx.ExecContext(ctx, ddl.QueryUpdateSecretMetadata,
		path, secret.Metadata.CurrentVersion, secret.Metadata.OldestVersion,
		secret.Metadata.CreatedTime,
		secret.Metadata.UpdatedTime, secret.Metadata.MaxVersions,
	)
	if err != nil {
		return sdkErrors.ErrStoreQueryFailure.Wrap(err)
	}

	// Update versions
	for version, sv := range secret.Versions {
		md, err := json.Marshal(sv.Data)
		if err != nil {
			return sdkErrors.ErrMarshalFailure.Wrap(err)
		}

		// TODO: check all errors.Join()'s and replace with Wraps.

		encrypted, nonce, err := s.encrypt(md)
		if err != nil {
			return sdkErrors.ErrCryptoEncryptionFailed.Wrap(err)
		}

		_, err = tx.ExecContext(ctx, ddl.QueryUpsertSecret,
			path, version, nonce, encrypted, sv.CreatedTime, sv.DeletedTime)
		if err != nil {
			return sdkErrors.ErrStoreQueryFailure.Wrap(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return sdkErrors.ErrTransactionCommitFailed.Wrap(err)
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
		failErr := sdkErrors.ErrNilContext
		log.FatalErr(fName, *failErr)
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
	fName := "LoadAllSecrets"

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all secret paths
	rows, err := s.db.QueryContext(ctx, ddl.QueryPathsFromMetadata)
	if err != nil {
		return nil, sdkErrors.ErrStoreQueryFailed.Wrap(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			failErr := sdkErrors.ErrFileCloseFailed
			log.FatalErr(fName, *failErr)
		}
	}(rows)

	// Map to store all secrets
	secrets := make(map[string]*kv.Value)

	// Iterate over paths
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, sdkErrors.ErrStoreQueryFailed.Wrap(err)
		}

		// Load the full secret for this path
		secret, err := s.loadSecretInternal(ctx, path)
		if err != nil {
			return nil, sdkErrors.ErrStoreQueryFailure.Wrap(err)
		}

		if secret != nil {
			secrets[path] = secret
		}
	}

	if err := rows.Err(); err != nil {
		return nil, sdkErrors.ErrStoreQueryFailure.Wrap(err)
	}

	return secrets, nil
}
