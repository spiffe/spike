//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/retry"
	"github.com/spiffe/spike-sdk-go/validation"
)

// withSerializableTx executes fn within a serializable transaction.
// It handles begin, commit, and automatic rollback on error.
//
// The function acquires a write lock on the DataStore for the duration of
// the transaction to ensure thread safety. The transaction uses
// serializable isolation level for strict consistency.
//
// SQLite can fail transiently (SQLITE_BUSY, SQLITE_LOCKED) when a
// competing transaction holds the database lock for longer than the
// driver's busy timeout, so the whole begin-execute-commit cycle is
// retried with exponential backoff for those errors. All other errors
// surface immediately. The overall operation is bounded by the
// SPIKE_NEXUS_DB_OPERATION_TIMEOUT deadline.
//
// Parameters:
//   - ctx: Context for the database operation
//   - fName: Function name for logging purposes
//   - fn: The work to execute within the transaction
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or an error if transaction
//     operations fail, fn returns a permanent error, or the retries are
//     exhausted
func (s *DataStore) withSerializableTx(
	ctx context.Context,
	fName string,
	fn func(tx *sql.Tx) *sdkErrors.SDKError,
) *sdkErrors.SDKError {
	validation.NonNilContextOrDie(ctx, fName)

	s.mu.Lock()
	defer s.mu.Unlock()

	opCtx, cancel := operationContext(ctx)
	defer cancel()

	var permanentErr *sdkErrors.SDKError
	_, retryErr := retry.Do(opCtx, func() (bool, *sdkErrors.SDKError) {
		txErr := s.attemptSerializableTx(opCtx, fName, fn)
		if txErr == nil {
			return true, nil
		}

		if transientDBError(txErr) {
			warnErr := *txErr.Clone()
			warnErr.Msg = "transient SQLite error: will retry"
			log.WarnErr(fName, warnErr)
			return false, txErr
		}

		// A permanent error: capture it and stop the retrier. Returning
		// success here only ends the retry loop; the captured error is
		// what the caller receives.
		permanentErr = txErr
		return true, nil
	},
		retry.WithBackOffOptions(
			retry.WithInitialInterval(dbRetryInitialInterval),
			retry.WithMaxInterval(dbRetryMaxInterval),
		),
	)

	if permanentErr != nil {
		return permanentErr
	}
	return retryErr
}

// attemptSerializableTx runs a single begin-execute-commit attempt. The
// caller must hold the DataStore write lock.
//
// Parameters:
//   - ctx: Context for the database operation
//   - fName: Function name for logging purposes
//   - fn: The work to execute within the transaction
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or an error if transaction
//     operations fail or fn returns an error
func (s *DataStore) attemptSerializableTx(
	ctx context.Context,
	fName string,
	fn func(tx *sql.Tx) *sdkErrors.SDKError,
) *sdkErrors.SDKError {
	tx, beginErr := s.db.BeginTx(
		ctx, &sql.TxOptions{Isolation: sql.LevelSerializable},
	)
	if beginErr != nil {
		return sdkErrors.ErrTransactionBeginFailed.Wrap(beginErr)
	}

	committed := false
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				failErr := sdkErrors.ErrTransactionRollbackFailed.Wrap(rollbackErr)
				log.WarnErr(fName, *failErr)
			}
		}
	}()

	if execErr := fn(tx); execErr != nil {
		return execErr
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return sdkErrors.ErrTransactionCommitFailed.Wrap(commitErr)
	}

	committed = true
	return nil
}
