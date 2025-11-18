//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"os"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

func (s *DataStore) createDataDir() error {
	return os.MkdirAll(s.Opts.DataDir, 0750)
}

func (s *DataStore) createTables(ctx context.Context, db *sql.DB) error {
	const fName = "createTables"
	if ctx == nil {
		log.FatalLn(fName, "message", sdkErrors.ErrCodeNilContext)
	}

	_, err := db.ExecContext(ctx, ddl.QueryInitialize)
	return err
}
