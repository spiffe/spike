//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike-sdk-go/validation"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

// createTables initializes the database schema by executing the DDL
// statements to create all required tables for secret and policy storage.
// This function is idempotent and can be called multiple times safely.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - db: The SQLite database connection on which to create the tables
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or ErrEntityQueryFailed if the
//     schema creation fails
func (s *DataStore) createTables(
	ctx context.Context, db *sql.DB,
) *sdkErrors.SDKError {
	const fName = "createTables"

	validation.NonNilContextOrDie(ctx, fName)

	_, err := db.ExecContext(ctx, ddl.QueryInitialize)
	if err != nil {
		return sdkErrors.ErrEntityQueryFailed.Wrap(err)
	}

	return nil
}
