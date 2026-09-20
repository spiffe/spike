//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike-sdk-go/validation"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

// loadSecretInternal retrieves a secret and all its versions from the database
// for the specified path. It performs the actual database operations including
// loading and decrypting metadata, fetching all versions, and decrypting the
// secret data.
//
// The function first queries and decrypts secret metadata (current version,
// timestamps),
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
//   - ErrStateIntegrityCheck: If a decrypted metadata field is not a valid
//     integer, or if the current version is missing from the versions map
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
//  2. Decrypts secret metadata
//  2. Fetches all versions from the secrets table
//  3. Decrypts each version using the DataStore's cipher
//  4. Unmarshals JSON data into a map[string]string format
//  5. Assembles the complete kv.Value structure
func (s *DataStore) loadSecretInternal(
	ctx context.Context, path string,
) (*kv.Value, *sdkErrors.SDKError) {
	const fName = "loadSecretInternal"

	validation.NonNilContextOrDie(ctx, fName)

	ctx, cancel := operationContext(ctx)
	defer cancel()

	var secret kv.Value
	var (
		nonce                   []byte
		encryptedCurrentVersion []byte
		encryptedOldestVersion  []byte
		encryptedCreatedTime    []byte
		encryptedUpdatedTime    []byte
		encryptedMaxVersions    []byte
	)

	// Load metadata
	metaErr := s.db.QueryRowContext(ctx, ddl.QueryLoadMetadata, path).Scan(
		&nonce,
		&encryptedCurrentVersion,
		&encryptedOldestVersion,
		&encryptedCreatedTime,
		&encryptedUpdatedTime,
		&encryptedMaxVersions,
	)
	if metaErr != nil {
		if errors.Is(metaErr, sql.ErrNoRows) {
			return nil, sdkErrors.ErrEntityNotFound
		}

		return nil, sdkErrors.ErrEntityLoadFailed
	}

	// Decrypt metadata using per-field derived nonces.
	currentVersionBytes, decryptErr := decryptWithDerivedNonce(
		s, nonce, nonceFieldSecretMetadataCurrentVersion, encryptedCurrentVersion,
	)
	if decryptErr != nil {
		return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(decryptErr)
	}
	oldestVersionBytes, decryptErr := decryptWithDerivedNonce(
		s, nonce, nonceFieldSecretMetadataOldestVersion, encryptedOldestVersion,
	)
	if decryptErr != nil {
		return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(decryptErr)
	}
	createdBytes, decryptErr := decryptWithDerivedNonce(
		s, nonce, nonceFieldSecretMetadataCreatedTime, encryptedCreatedTime,
	)
	if decryptErr != nil {
		return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(decryptErr)
	}
	updatedBytes, decryptErr := decryptWithDerivedNonce(
		s, nonce, nonceFieldSecretMetadataUpdatedTime, encryptedUpdatedTime,
	)
	if decryptErr != nil {
		return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(decryptErr)
	}
	maxVersionsBytes, decryptErr := decryptWithDerivedNonce(
		s, nonce, nonceFieldSecretMetadataMaxVersions, encryptedMaxVersions,
	)
	if decryptErr != nil {
		return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(decryptErr)
	}

	// Decode into the struct. SPIKE wrote this plaintext itself, so a value
	// that does not parse means the row is corrupt. A secret whose metadata
	// cannot be trusted is not loaded.
	currentVersion, parseErr := parseMetadataInt(
		"current_version", currentVersionBytes)
	if parseErr != nil {
		return nil, parseErr
	}
	oldestVersion, parseErr := parseMetadataInt(
		"oldest_version", oldestVersionBytes)
	if parseErr != nil {
		return nil, parseErr
	}
	createdSec, parseErr := parseMetadataInt64("created_time", createdBytes)
	if parseErr != nil {
		return nil, parseErr
	}
	updatedSec, parseErr := parseMetadataInt64("updated_time", updatedBytes)
	if parseErr != nil {
		return nil, parseErr
	}
	maxVersions, parseErr := parseMetadataInt(
		"max_versions", maxVersionsBytes)
	if parseErr != nil {
		return nil, parseErr
	}

	secret.Metadata.CurrentVersion = currentVersion
	secret.Metadata.OldestVersion = oldestVersion
	secret.Metadata.CreatedTime = time.Unix(createdSec, 0)
	secret.Metadata.UpdatedTime = time.Unix(updatedSec, 0)
	secret.Metadata.MaxVersions = maxVersions

	// Load versions
	rows, queryErr := s.db.QueryContext(ctx, ddl.QuerySecretVersions, path)
	if queryErr != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(queryErr)
	}
	defer func(rows *sql.Rows) {
		closeErr := rows.Close()
		if closeErr != nil {
			failErr := sdkErrors.ErrFSFileCloseFailed.Wrap(closeErr)
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

		if scanErr := rows.Scan(
			&version, &nonce,
			&encrypted, &createdTime, &deletedTime,
		); scanErr != nil {
			return nil, sdkErrors.ErrEntityQueryFailed.Wrap(scanErr)
		}

		decrypted, decryptErr := s.decrypt(encrypted, nonce)
		if decryptErr != nil {
			return nil, sdkErrors.ErrCryptoDecryptionFailed.Wrap(decryptErr)
		}

		var values map[string]string
		if unmarshalErr := json.Unmarshal(decrypted, &values); unmarshalErr != nil {
			return nil, sdkErrors.ErrDataUnmarshalFailure.Wrap(unmarshalErr)
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

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(rowsErr)
	}

	// Integrity check: If CurrentVersion is non-zero, it must exist in
	// the Versions map. CurrentVersion==0 indicates a "shell secret"
	// where all versions are deleted, which is valid.
	if secret.Metadata.CurrentVersion != 0 {
		if _, exists := secret.Versions[secret.Metadata.CurrentVersion]; !exists {
			integrityErr := sdkErrors.ErrStateIntegrityCheck.Clone()
			integrityErr.Msg = "data integrity violation: current version not found"
			return nil, integrityErr
		}
	}

	return &secret, nil
}

// parseMetadataInt decodes a decrypted secret-metadata field into an int.
//
// A value that does not parse indicates a corrupt row and is reported as an
// integrity violation that names the offending field.
//
// Parameters:
//   - field: The metadata field name, used in the error message.
//   - raw: The decrypted plaintext of the field.
//
// Returns:
//   - int: The parsed value.
//   - *sdkErrors.SDKError: ErrStateIntegrityCheck if the value is not a valid
//     integer, nil on success.
func parseMetadataInt(
	field string, raw []byte,
) (int, *sdkErrors.SDKError) {
	value, parseErr := strconv.Atoi(string(raw))
	if parseErr != nil {
		return 0, metadataIntegrityErr(field, parseErr)
	}
	return value, nil
}

// parseMetadataInt64 decodes a decrypted secret-metadata timestamp field
// into an int64 of Unix seconds.
//
// A value that does not parse indicates a corrupt row and is reported as an
// integrity violation that names the offending field.
//
// Parameters:
//   - field: The metadata field name, used in the error message.
//   - raw: The decrypted plaintext of the field.
//
// Returns:
//   - int64: The parsed value.
//   - *sdkErrors.SDKError: ErrStateIntegrityCheck if the value is not a valid
//     integer, nil on success.
func parseMetadataInt64(
	field string, raw []byte,
) (int64, *sdkErrors.SDKError) {
	value, parseErr := strconv.ParseInt(string(raw), 10, 64)
	if parseErr != nil {
		return 0, metadataIntegrityErr(field, parseErr)
	}
	return value, nil
}

// metadataIntegrityErr builds the integrity error returned when a
// secret-metadata field cannot be decoded.
//
// Wrap returns a fresh copy of the sentinel, so setting the message does not
// touch the shared ErrStateIntegrityCheck value.
//
// Parameters:
//   - field: The metadata field name to include in the message.
//   - cause: The underlying parse error.
//
// Returns:
//   - *sdkErrors.SDKError: ErrStateIntegrityCheck wrapping cause, with a
//     message that names the field.
func metadataIntegrityErr(field string, cause error) *sdkErrors.SDKError {
	integrityErr := sdkErrors.ErrStateIntegrityCheck.Wrap(cause)
	integrityErr.Msg = "data integrity violation: secret metadata field " +
		field + " is not a valid integer"
	return integrityErr
}
