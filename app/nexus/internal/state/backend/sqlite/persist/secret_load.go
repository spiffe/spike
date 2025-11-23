//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

// loadSecretInternal retrieves a secret and all its versions from the database
// for the specified path. It performs the actual database operations including
// loading metadata, fetching all versions, and decrypting the secret data.
//
// The function first queries for secret metadata (current version, timestamps),
// then retrieves all versions of the secret, decrypts each version, and
// reconstructs the complete secret structure.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - path: The secret path to load
//
// Returns:
//   - *kv.Value: The complete secret with all versions and metadata
//   - *sdkErrors.SDKError: An error if the secret is not found or any database
//     or decryption operation fails. Returns nil on success.
//
// Possible errors:
//   - ErrEntityNotFound: If the secret does not exist at the specified path
//   - ErrEntityLoadFailed: If loading secret metadata fails
//   - ErrEntityQueryFailed: If querying versions fails or rows.Scan fails
//   - ErrCryptoDecryptionFailed: If decrypting a version fails
//   - ErrDataUnmarshalFailure: If unmarshaling JSON data fails
//
// Special behavior:
//   - Automatically handles deleted versions by setting DeletedTime when
//     present
//
// The function handles the following operations:
//  1. Queries secret metadata from the secret_metadata table
//  2. Fetches all versions from the secrets table
//  3. Decrypts each version using the DataStore's cipher
//  4. Unmarshals JSON data into map[string]string format
//  5. Assembles the complete kv.Value structure
func (s *DataStore) loadSecretInternal(
	ctx context.Context, path string,
) (*kv.Value, *sdkErrors.SDKError) {
	const fName = "loadSecretInternal"

	validateContext(ctx, fName) // TODO: this can be an SDK utility function and can be used across the codebase instead of just this package.

	var secret kv.Value

	// Load metadata
	err := s.db.QueryRowContext(ctx, ddl.QuerySecretMetadata, path).Scan(
		&secret.Metadata.CurrentVersion,
		&secret.Metadata.OldestVersion,
		&secret.Metadata.CreatedTime,
		&secret.Metadata.UpdatedTime,
		&secret.Metadata.MaxVersions)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sdkErrors.ErrEntityNotFound
		}

		return nil, sdkErrors.ErrEntityLoadFailed
	}

	// Load versions
	rows, err := s.db.QueryContext(ctx, ddl.QuerySecretVersions, path)
	if err != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			failErr := sdkErrors.ErrFSFileCloseFailed.Wrap(err)
			log.WarnErr(fName, *failErr)
		}
	}(rows)

	secret.Versions = make(map[int]kv.Version)
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
			return nil, sdkErrors.ErrEntityQueryFailed.Wrap(err)
		}

		decrypted, err := s.decrypt(encrypted, nonce)
		if err != nil {
			return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
		}

		var values map[string]string
		if err := json.Unmarshal(decrypted, &values); err != nil {
			return nil, sdkErrors.ErrDataUnmarshalFailure.Wrap(err)
		}

		sv := kv.Version{
			Data:        values,
			CreatedTime: createdTime,
		}
		if deletedTime.Valid {
			sv.DeletedTime = &deletedTime.Time
		}

		secret.Versions[version] = sv
	}

	if err := rows.Err(); err != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(err)
	}

	// Integrity check: If CurrentVersion is non-zero, it must exist in
	// the Versions map. CurrentVersion==0 indicates a "shell secret"
	// where all versions are deleted, which is valid.
	if secret.Metadata.CurrentVersion != 0 {
		if _, exists := secret.Versions[secret.Metadata.CurrentVersion]; !exists {
			return nil, sdkErrors.ErrEntityQueryFailed.Wrap( // TODO: have a special dataintegritycheck error on the SDK for this case.
				errors.New(
					"data integrity violation: current version not found",
				),
			)
		}
	}

	return &secret, nil
}
