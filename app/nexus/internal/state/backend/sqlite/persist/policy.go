//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

// DeletePolicy removes a policy from the database by its ID.
//
// Uses serializable transaction isolation to ensure consistency.
// Automatically rolls back on error.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: Unique identifier of the policy to delete
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or an error if transaction
//     operations fail or policy deletion fails
func (s *DataStore) DeletePolicy(
	ctx context.Context, id string,
) *sdkErrors.SDKError {
	const fName = "DeletePolicy"
	validateContext(ctx, fName)

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		failErr := sdkErrors.ErrTransactionBeginFailed.Wrap(err)
		return failErr
	}

	committed := false
	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				failErr := sdkErrors.ErrTransactionRollbackFailed.Wrap(err)
				log.WarnErr(fName, *failErr)
			}
		}
	}(tx)

	_, err = tx.ExecContext(ctx, ddl.QueryDeletePolicy, id)
	if err != nil {
		failErr := sdkErrors.ErrEntityQueryFailed.Wrap(err)
		return failErr
	}

	if err := tx.Commit(); err != nil {
		failErr := sdkErrors.ErrTransactionCommitFailed.Wrap(err)
		return failErr
	}

	committed = true
	return nil
}

// StorePolicy saves or updates a policy in the database.
//
// Uses serializable transaction isolation to ensure consistency.
// Automatically rolls back on error.
//
// Parameters:
//   - ctx: Context for the database operation
//   - policy: Policy data to store, containing ID, name, patterns, and creation
//     time
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or an error if transaction
//     operations fail, encryption fails, or policy storage fails
func (s *DataStore) StorePolicy(
	ctx context.Context, policy data.Policy,
) *sdkErrors.SDKError {
	const fName = "StorePolicy"
	validateContext(ctx, fName)

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		failErr := sdkErrors.ErrTransactionBeginFailed.Wrap(err)
		return failErr
	}

	committed := false
	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				failErr := sdkErrors.ErrTransactionRollbackFailed.Wrap(err)
				log.WarnErr(fName, *failErr)
			}
		}
	}(tx)

	// Serialize permissions to comma-separated string
	permissionsStr := ""
	if len(policy.Permissions) > 0 {
		permissions := make([]string, len(policy.Permissions))
		for i, perm := range policy.Permissions {
			permissions[i] = string(perm)
		}
		permissionsStr = strings.Join(permissions, ",")
	}

	// Encryption
	nonce, err := generateNonce(s)
	if err != nil {
		return sdkErrors.ErrCryptoNonceGenerationFailed.Wrap(err)
	}
	encryptedSpiffeID, err := encryptWithNonce(
		s, nonce, []byte(policy.SPIFFEIDPattern),
	)
	if err != nil {
		failErr := sdkErrors.ErrCryptoEncryptionFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf(
			"failed to encrypt SPIFFE ID pattern for policy %s", policy.ID,
		)
		return failErr
	}

	encryptedPathPattern, err := encryptWithNonce(
		s, nonce, []byte(policy.PathPattern),
	)

	// TODO: some of these are integrity check errors and shall be logged as such
	// with a dedicated sentinel error kind.

	if err != nil {
		failErr := sdkErrors.ErrCryptoEncryptionFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf(
			"failed to encrypt path pattern for policy %s", policy.ID,
		)
		return failErr
	}
	encryptedPermissions, err := encryptWithNonce(
		s, nonce, []byte(permissionsStr),
	)
	if err != nil {
		failErr := sdkErrors.ErrCryptoEncryptionFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf(
			"failed to encrypt permissions for policy %s", policy.ID,
		)
		return failErr
	}

	_, err = tx.ExecContext(ctx, ddl.QueryUpsertPolicy,
		policy.ID,
		policy.Name,
		nonce,
		encryptedSpiffeID,
		encryptedPathPattern,
		encryptedPermissions,
		policy.CreatedAt.Unix(),
		policy.UpdatedAt.Unix(),
	)

	if err != nil {
		failErr := sdkErrors.ErrEntityQueryFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf("failed to upsert policy %s", policy.ID)
		return failErr
	}

	if err := tx.Commit(); err != nil {
		return sdkErrors.ErrTransactionCommitFailed.Wrap(err)
	}

	committed = true
	return nil
}

// LoadPolicy retrieves a policy from the database and compiles its patterns.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: Unique identifier of the policy to load
//
// Returns:
//   - *data.Policy: Loaded policy with compiled patterns, nil if not found or
//     if an error occurs
//   - *sdkErrors.SDKError: nil on success, sdkErrors.ErrEntityNotFound if the
//     policy does not exist, or an error if database operations fail,
//     decryption fails, or pattern compilation fails
func (s *DataStore) LoadPolicy(
	ctx context.Context, id string,
) (*data.Policy, *sdkErrors.SDKError) {
	const fName = "LoadPolicy"
	validateContext(ctx, fName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var policy data.Policy
	var encryptedSPIFFEIDPattern []byte
	var encryptedPathPattern []byte
	var encryptedPermissions []byte
	var nonce []byte
	var createdTime int64
	var updatedTime int64

	err := s.db.QueryRowContext(ctx, ddl.QueryLoadPolicy, id).Scan(
		&policy.ID,
		&policy.Name,
		&encryptedSPIFFEIDPattern,
		&encryptedPathPattern,
		&encryptedPermissions,
		&nonce,
		&createdTime,
		&updatedTime,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sdkErrors.ErrEntityNotFound
		}
		failErr := sdkErrors.ErrEntityLoadFailed.Wrap(err)
		return nil, failErr
	}

	// TODO: some of these are integrity check errors and shall be logged as such
	// with a dedicated sentinel error kind.

	// Decrypt
	decryptedSPIFFEIDPattern, err := s.decrypt(encryptedSPIFFEIDPattern, nonce)
	if err != nil {
		failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf(
			"failed to decrypt SPIFFE ID pattern for policy %s", policy.ID,
		)
		return nil, failErr
	}
	decryptedPathPattern, err := s.decrypt(encryptedPathPattern, nonce)
	if err != nil {
		failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf(
			"failed to decrypt path pattern for policy %s", policy.ID,
		)
		return nil, failErr
	}
	decryptedPermissions, err := s.decrypt(encryptedPermissions, nonce)
	if err != nil {
		failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf(
			"failed to decrypt permissions for policy %s", policy.ID,
		)
		return nil, failErr
	}

	// Set decrypted values
	policy.SPIFFEIDPattern = string(decryptedSPIFFEIDPattern)
	policy.PathPattern = string(decryptedPathPattern)
	policy.CreatedAt = time.Unix(createdTime, 0)
	policy.UpdatedAt = time.Unix(updatedTime, 0)

	policy.Permissions = deserializePermissions(string(decryptedPermissions))

	// Compile regex
	if err := compileRegexPatterns(&policy); err != nil {
		return nil, err
	}

	return &policy, nil
}

// LoadAllPolicies retrieves all policies from the backend storage.
//
// The function loads all policy data and compiles regex patterns for SPIFFE ID
// and path matching. If any individual policy fails to load, decrypt, or
// compile (due to corruption or invalid data), the error is logged as a
// warning and that policy is skipped. This allows the system to continue
// operating with valid policies even when some policies are corrupted.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - map[string]*data.Policy: Map of policy IDs to successfully loaded
//     policies with compiled patterns. May be incomplete if some policies
//     failed to load (check logs for warnings).
//   - *sdkErrors.SDKError: nil on success, or an error if the database query
//     itself fails or if iterating over rows fails. Individual policy load
//     failures do not cause the function to return an error.
func (s *DataStore) LoadAllPolicies(
	ctx context.Context,
) (map[string]*data.Policy, *sdkErrors.SDKError) {
	const fName = "LoadAllPolicies"
	validateContext(ctx, fName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, ddl.QueryAllPolicies)
	if err != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			failErr := sdkErrors.ErrFSFileCloseFailed.Wrap(err)
			failErr.Msg = fmt.Sprintf("failed to close rows")
			log.WarnErr(fName, *failErr)
		}
	}(rows)

	policies := make(map[string]*data.Policy)

	for rows.Next() {
		var policy data.Policy
		var encryptedSPIFFEIDPattern []byte
		var encryptedPathPattern []byte
		var encryptedPermissions []byte
		var nonce []byte
		var createdTime int64
		var updatedTime int64

		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&encryptedSPIFFEIDPattern,
			&encryptedPathPattern,
			&encryptedPermissions,
			&nonce,
			&createdTime,
			&updatedTime,
		); err != nil {
			failErr := sdkErrors.ErrEntityQueryFailed.Wrap(err)
			failErr.Msg = "failed to scan policy row, skipping"
			log.WarnErr(fName, *failErr)
			continue
		}

		// Decrypt
		decryptedSPIFFEIDPattern, err := s.decrypt(encryptedSPIFFEIDPattern, nonce)
		if err != nil {
			failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
			failErr.Msg = fmt.Sprintf(
				"failed to decrypt SPIFFE ID pattern for policy %s, skipping",
				policy.ID,
			)
			log.WarnErr(fName, *failErr)
			continue
		}
		decryptedPathPattern, err := s.decrypt(encryptedPathPattern, nonce)
		if err != nil {
			failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
			failErr.Msg = fmt.Sprintf(
				"failed to decrypt path pattern for policy %s, skipping",
				policy.ID,
			)
			log.WarnErr(fName, *failErr)
			continue
		}
		decryptedPermissions, err := s.decrypt(encryptedPermissions, nonce)
		if err != nil {
			failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(err)
			failErr.Msg = fmt.Sprintf(
				"failed to decrypt permissions for policy %s, skipping",
				policy.ID,
			)
			log.WarnErr(fName, *failErr)
			continue
		}

		policy.SPIFFEIDPattern = string(decryptedSPIFFEIDPattern)
		policy.PathPattern = string(decryptedPathPattern)
		policy.CreatedAt = time.Unix(createdTime, 0)
		policy.UpdatedAt = time.Unix(updatedTime, 0)

		policy.Permissions = deserializePermissions(
			string(decryptedPermissions),
		)

		// Compile regex
		if err := compileRegexPatterns(&policy); err != nil {
			failErr := sdkErrors.ErrEntityInvalid.Wrap(err)
			failErr.Msg = fmt.Sprintf(
				"failed to compile regex patterns for policy %s, skipping",
				policy.ID,
			)
			log.WarnErr(fName, *failErr)
			continue
		}

		policies[policy.ID] = &policy
	}

	if err := rows.Err(); err != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(err)
	}

	return policies, nil
}
