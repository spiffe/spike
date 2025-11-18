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
//   - *kv.Value: The complete secret with all versions and metadata.
//     Returns nil if the secret does not exist.
//   - error: An error if any database or decryption operation fails.
//     Returns nil error with nil secret for non-existent paths.
//
// Special behavior:
//   - Returns (nil, nil) when the secret doesn't exist (sql.ErrNoRows)
//   - Returns (nil, error) for actual errors (database, decryption,
//     unmarshaling)
//   - Automatically handles deleted versions by setting DeletedTime when present
//
// The function handles the following operations:
//  1. Queries secret metadata from the `secret_metadata` table
//  2. Fetches all versions from the `secrets` table
//  3. Decrypts each version using the DataStore's cipher
//  4. Unmarshals JSON data into `map[string]string` format
//  5. Assembles the complete kv.Value structure
func (s *DataStore) loadSecretInternal(
	ctx context.Context, path string,
) (*kv.Value, error) {
	const fName = "loadSecretInternal"
	if ctx == nil {
		log.FatalLn(fName, "message", sdkErrors.ErrCodeNilContext)
	}

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
			return nil, nil
		}

		failErr := sdkErrors.ErrDataLoadFailed
		return nil, errors.Join(failErr, err)
	}

	// Load versions
	rows, err := s.db.QueryContext(ctx, ddl.QuerySecretVersions, path)
	if err != nil {
		failErr := sdkErrors.ErrQueryFailure
		return nil, errors.Join(failErr, err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			failErr := sdkErrors.ErrFileCloseFailed
			log.Log().Warn(fName, "message", errors.Join(failErr, err).Error())
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
			failErr := sdkErrors.ErrQueryFailure
			return nil, errors.Join(failErr, err)
		}

		decrypted, err := s.decrypt(encrypted, nonce)
		if err != nil {
			failErr := sdkErrors.ErrCryptoDecryptionFailed
			return nil, errors.Join(failErr, err)
		}

		var values map[string]string
		if err := json.Unmarshal(decrypted, &values); err != nil {
			failErr := sdkErrors.ErrUnmarshalFailure
			return nil, errors.Join(failErr, err)
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
		failErr := sdkErrors.ErrQueryFailure
		return nil, errors.Join(failErr, err)
	}

	return &secret, nil
}
