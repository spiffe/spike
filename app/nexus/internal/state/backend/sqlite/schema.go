//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"database/sql"
	"os"
)

func (s *DataStore) createDataDir() error {
	return os.MkdirAll(s.opts.DataDir, 0750)
}

func (s *DataStore) createTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, queryInitialize)
	return err
}
