//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"encoding/json"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike-sdk-go/validation"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

// StoreSecret stores a secret at the specified path with its metadata and
// versions. It performs the following operations atomically within a
// transaction:
//   - Updates the secret metadata (current version, creation time, update time)
//   - Stores all secret versions with their respective data encrypted using
//     AES-GCM
//
// The secret data is JSON-encoded before encryption. This method is
// thread-safe.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - path: The secret path where the secret will be stored
//   - secret: The secret value containing metadata and versions to store
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or one of the following errors:
//   - ErrTransactionBeginFailed: If the transaction fails to begin
//   - ErrEntityQueryFailed: If database operations fail
//   - ErrDataMarshalFailure: If data marshaling fails
//   - ErrCryptoEncryptionFailed: If encryption fails
//   - ErrTransactionCommitFailed: If the transaction fails to commit
func (s *DataStore) StoreSecret(
	ctx context.Context, path string, secret kv.Value,
) *sdkErrors.SDKError {
	// Pre-encrypt all versions before starting the transaction
	type encryptedVersion struct {
		version     int
		nonce       []byte
		encrypted   []byte
		createdTime any
		deletedTime any
	}

	encryptedVersions := make([]encryptedVersion, 0, len(secret.Versions))
	for version, sv := range secret.Versions {
		md, marshalErr := json.Marshal(sv.Data)
		if marshalErr != nil {
			return sdkErrors.ErrDataMarshalFailure.Wrap(marshalErr)
		}

		encrypted, nonce, encryptErr := s.encrypt(md)
		if encryptErr != nil {
			return sdkErrors.ErrCryptoEncryptionFailed.Wrap(encryptErr)
		}

		encryptedVersions = append(encryptedVersions, encryptedVersion{
			version:     version,
			nonce:       nonce,
			encrypted:   encrypted,
			createdTime: sv.CreatedTime,
			deletedTime: sv.DeletedTime,
		})
	}

	return s.withSerializableTx(ctx, "StoreSecret",
		func(tx *sql.Tx) *sdkErrors.SDKError {
			// Update metadata
			_, metaErr := tx.ExecContext(ctx, ddl.QueryUpdateSecretMetadata,
				path, secret.Metadata.CurrentVersion, secret.Metadata.OldestVersion,
				secret.Metadata.CreatedTime,
				secret.Metadata.UpdatedTime, secret.Metadata.MaxVersions,
			)
			if metaErr != nil {
				return sdkErrors.ErrEntityQueryFailed.Wrap(metaErr)
			}

			// Update versions
			for _, ev := range encryptedVersions {
				_, execErr := tx.ExecContext(ctx, ddl.QueryUpsertSecret,
					path, ev.version, ev.nonce, ev.encrypted,
					ev.createdTime, ev.deletedTime)
				if execErr != nil {
					return sdkErrors.ErrEntityQueryFailed.Wrap(execErr)
				}
			}

			return nil
		})
}

// LoadSecret retrieves a secret and all its versions from the specified path.
// It performs the following operations:
// - Loads the secret metadata
// - Retrieves all secret versions
// - Decrypts and unmarshals the version data
//
// This method is thread-safe.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - path: The secret path to load
//
// Returns:
//   - *kv.Value: The decrypted secret with all its versions, or nil if the
//     secret doesn't exist
//   - *sdkErrors.SDKError: nil on success, or one of the following errors:
//   - ErrEntityLoadFailed: If loading secret metadata fails
//   - ErrEntityQueryFailed: If querying versions fails
//   - ErrCryptoDecryptionFailed: If decrypting a version fails
//   - ErrDataUnmarshalFailure: If unmarshaling JSON data fails
func (s *DataStore) LoadSecret(
	ctx context.Context, path string,
) (*kv.Value, *sdkErrors.SDKError) {
	const fName = "LoadSecret"

	validation.NonNilContextOrDie(ctx, fName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.loadSecretInternal(ctx, path)
}

// LoadAllSecrets retrieves all secrets from the database. It returns a map
// where the keys are secret paths and the values are the corresponding
// secrets. Each secret includes its metadata and all versions with decrypted
// data. This method is thread-safe.
//
// If any individual secret fails to load or decrypt (due to corruption or
// invalid data), the error is logged as a warning and that secret is skipped.
// This allows the system to continue operating with valid secrets even when
// some secrets are corrupted.
//
// Contexts that are canceled or reach their deadline will result in the
// operation being interrupted early and returning an error.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - map[string]*kv.Value: A map of secret paths to their corresponding
//     secret values. May be incomplete if some secrets failed to load (check
//     logs for warnings).
//   - *sdkErrors.SDKError: nil on success, or an error if the database query
//     itself fails or if iterating over rows fails. Individual secret load
//     failures do not cause the function to return an error.
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
) (map[string]*kv.Value, *sdkErrors.SDKError) {
	fName := "LoadAllSecrets"

	validation.NonNilContextOrDie(ctx, fName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all secret paths
	rows, err := s.db.QueryContext(ctx, ddl.QueryPathsFromMetadata)
	if err != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(err)
	}
	defer func(rows *sql.Rows) {
		closeErr := rows.Close()
		if closeErr != nil {
			failErr := *sdkErrors.ErrFSFileCloseFailed.Wrap(closeErr)
			log.WarnErr(fName, failErr)
		}
	}(rows)

	// Map to store all secrets
	secrets := make(map[string]*kv.Value)

	// Iterate over paths
	for rows.Next() {
		var path string
		if scanErr := rows.Scan(&path); scanErr != nil {
			failErr := sdkErrors.ErrEntityQueryFailed.Wrap(scanErr)
			failErr.Msg = "failed to scan secret path row, skipping"
			log.WarnErr(fName, *failErr)
			continue
		}

		// Load the full secret for this path
		secret, loadErr := s.loadSecretInternal(ctx, path)
		if loadErr != nil {
			failErr := sdkErrors.ErrEntityLoadFailed.Wrap(loadErr)
			failErr.Msg = "failed to load secret at path " + path + ", skipping"
			log.WarnErr(fName, *failErr)
			continue
		}

		if secret != nil {
			secrets[path] = secret
		}
	}

	if rowErr := rows.Err(); rowErr != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(rowErr)
	}

	return secrets, nil
}
