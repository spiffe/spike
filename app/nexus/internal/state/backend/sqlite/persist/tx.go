//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/validation"
)

// withSerializableTx executes fn within a serializable transaction.
// It handles begin, commit, and automatic rollback on error.
//
// The function acquires a write lock on the DataStore for the duration of the
// transaction to ensure thread safety. The transaction uses serializable
// isolation level for strict consistency.
//
// Parameters:
//   - ctx: Context for the database operation
//   - fName: Function name for logging purposes
//   - fn: The work to execute within the transaction
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or an error if transaction
//     operations fail or fn returns an error
func (s *DataStore) withSerializableTx(
	ctx context.Context,
	fName string,
	fn func(tx *sql.Tx) *sdkErrors.SDKError,
) *sdkErrors.SDKError {
	validation.NonNilContextOrDie(ctx, fName)

	s.mu.Lock()
	defer s.mu.Unlock()

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
