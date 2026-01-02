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

	"github.com/spiffe/spike-sdk-go/validation"
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

	validation.NonNilContextOrDie(ctx, fName)

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, beginErr := s.db.BeginTx(
		ctx, &sql.TxOptions{Isolation: sql.LevelSerializable},
	)
	if beginErr != nil {
		failErr := sdkErrors.ErrTransactionBeginFailed.Wrap(beginErr)
		return failErr
	}

	committed := false
	defer func(tx *sql.Tx) {
		if !committed {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				failErr := sdkErrors.ErrTransactionRollbackFailed.Wrap(rollbackErr)
				log.WarnErr(fName, *failErr)
			}
		}
	}(tx)

	_, execErr := tx.ExecContext(ctx, ddl.QueryDeletePolicy, id)
	if execErr != nil {
		failErr := sdkErrors.ErrEntityQueryFailed.Wrap(execErr)
		return failErr
	}

	if commitErr := tx.Commit(); commitErr != nil {
		failErr := sdkErrors.ErrTransactionCommitFailed.Wrap(commitErr)
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

	validation.NonNilContextOrDie(ctx, fName)

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, beginErr := s.db.BeginTx(
		ctx, &sql.TxOptions{Isolation: sql.LevelSerializable},
	)
	if beginErr != nil {
		failErr := sdkErrors.ErrTransactionBeginFailed.Wrap(beginErr)
		return failErr
	}

	committed := false
	defer func(tx *sql.Tx) {
		if !committed {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				failErr := sdkErrors.ErrTransactionRollbackFailed.Wrap(rollbackErr)
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
	nonce, nonceErr := generateNonce(s)
	if nonceErr != nil {
		return sdkErrors.ErrCryptoNonceGenerationFailed.Wrap(nonceErr)
	}
	encryptedSpiffeID, encErr := encryptWithNonce(
		s, nonce, []byte(policy.SPIFFEIDPattern),
	)
	if encErr != nil {
		failErr := sdkErrors.ErrCryptoEncryptionFailed.Wrap(encErr)
		failErr.Msg = fmt.Sprintf(
			"failed to encrypt SPIFFE ID pattern for policy %s", policy.ID,
		)
		return failErr
	}

	encryptedPathPattern, pathErr := encryptWithNonce(
		s, nonce, []byte(policy.PathPattern),
	)

	if pathErr != nil {
		failErr := sdkErrors.ErrCryptoEncryptionFailed.Wrap(pathErr)
		failErr.Msg = fmt.Sprintf(
			"failed to encrypt path pattern for policy %s", policy.ID,
		)
		return failErr
	}
	encryptedPermissions, permErr := encryptWithNonce(
		s, nonce, []byte(permissionsStr),
	)
	if permErr != nil {
		failErr := sdkErrors.ErrCryptoEncryptionFailed.Wrap(permErr)
		failErr.Msg = fmt.Sprintf(
			"failed to encrypt permissions for policy %s", policy.ID,
		)
		return failErr
	}

	_, execErr := tx.ExecContext(ctx, ddl.QueryUpsertPolicy,
		policy.ID,
		policy.Name,
		nonce,
		encryptedSpiffeID,
		encryptedPathPattern,
		encryptedPermissions,
		policy.CreatedAt.Unix(),
		policy.UpdatedAt.Unix(),
	)

	if execErr != nil {
		failErr := sdkErrors.ErrEntityQueryFailed.Wrap(execErr)
		failErr.Msg = fmt.Sprintf("failed to upsert policy %s", policy.ID)
		return failErr
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return sdkErrors.ErrTransactionCommitFailed.Wrap(commitErr)
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

	validation.NonNilContextOrDie(ctx, fName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var policy data.Policy
	var encryptedSPIFFEIDPattern []byte
	var encryptedPathPattern []byte
	var encryptedPermissions []byte
	var nonce []byte
	var createdTime int64
	var updatedTime int64

	scanErr := s.db.QueryRowContext(ctx, ddl.QueryLoadPolicy, id).Scan(
		&policy.ID,
		&policy.Name,
		&encryptedSPIFFEIDPattern,
		&encryptedPathPattern,
		&encryptedPermissions,
		&nonce,
		&createdTime,
		&updatedTime,
	)
	if scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil, sdkErrors.ErrEntityNotFound
		}
		failErr := sdkErrors.ErrEntityLoadFailed.Wrap(scanErr)
		return nil, failErr
	}

	// Decrypt
	decryptedSPIFFEIDPattern, spiffeDecryptErr := s.decrypt(
		encryptedSPIFFEIDPattern, nonce,
	)
	if spiffeDecryptErr != nil {
		failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(spiffeDecryptErr)
		failErr.Msg = fmt.Sprintf(
			"failed to decrypt SPIFFE ID pattern for policy %s", policy.ID,
		)
		return nil, failErr
	}
	decryptedPathPattern, pathDecryptErr := s.decrypt(encryptedPathPattern, nonce)
	if pathDecryptErr != nil {
		failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(pathDecryptErr)
		failErr.Msg = fmt.Sprintf(
			"failed to decrypt path pattern for policy %s", policy.ID,
		)
		return nil, failErr
	}
	decryptedPermissions, permDecryptErr := s.decrypt(encryptedPermissions, nonce)
	if permDecryptErr != nil {
		failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(permDecryptErr)
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
	if compileErr := compileRegexPatterns(&policy); compileErr != nil {
		return nil, compileErr
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

	validation.NonNilContextOrDie(ctx, fName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, queryErr := s.db.QueryContext(ctx, ddl.QueryAllPolicies)
	if queryErr != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(queryErr)
	}
	defer func(rows *sql.Rows) {
		closeErr := rows.Close()
		if closeErr != nil {
			failErr := sdkErrors.ErrFSFileCloseFailed.Wrap(closeErr)
			failErr.Msg = "failed to close rows"
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

		if scanErr := rows.Scan(
			&policy.ID,
			&policy.Name,
			&encryptedSPIFFEIDPattern,
			&encryptedPathPattern,
			&encryptedPermissions,
			&nonce,
			&createdTime,
			&updatedTime,
		); scanErr != nil {
			failErr := sdkErrors.ErrEntityQueryFailed.Wrap(scanErr)
			failErr.Msg = "failed to scan policy row, skipping"
			log.WarnErr(fName, *failErr)
			continue
		}

		// Decrypt
		decryptedSPIFFEIDPattern, spiffeDecryptErr := s.decrypt(
			encryptedSPIFFEIDPattern, nonce,
		)
		if spiffeDecryptErr != nil {
			failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(spiffeDecryptErr)
			failErr.Msg = fmt.Sprintf(
				"failed to decrypt SPIFFE ID pattern for policy %s, skipping",
				policy.ID,
			)
			log.WarnErr(fName, *failErr)
			continue
		}
		decryptedPathPattern, pathDecryptErr := s.decrypt(
			encryptedPathPattern, nonce,
		)
		if pathDecryptErr != nil {
			failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(pathDecryptErr)
			failErr.Msg = fmt.Sprintf(
				"failed to decrypt path pattern for policy %s, skipping",
				policy.ID,
			)
			log.WarnErr(fName, *failErr)
			continue
		}
		decryptedPermissions, permDecryptErr := s.decrypt(
			encryptedPermissions, nonce,
		)
		if permDecryptErr != nil {
			failErr := sdkErrors.ErrCryptoDecryptionFailed.Wrap(permDecryptErr)
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
		if compileErr := compileRegexPatterns(&policy); compileErr != nil {
			failErr := sdkErrors.ErrEntityInvalid.Wrap(compileErr)
			failErr.Msg = fmt.Sprintf(
				"failed to compile regex patterns for policy %s, skipping",
				policy.ID,
			)
			log.WarnErr(fName, *failErr)
			continue
		}

		policies[policy.ID] = &policy
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, sdkErrors.ErrEntityQueryFailed.Wrap(rowsErr)
	}

	return policies, nil
}
