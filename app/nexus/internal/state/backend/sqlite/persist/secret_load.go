//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
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

	"github.com/spiffe/spike-sdk-go/kv"

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
//  4. Unmarshals JSON data into map[string]string format
//  5. Assembles the complete kv.Value structure
func (s *DataStore) loadSecretInternal(
	ctx context.Context, path string,
) (*kv.Value, error) {
	var secret kv.Value

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
		return nil, fmt.Errorf("failed to iterate secret versions: %w", err)
	}

	return &secret, nil
}
