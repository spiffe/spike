//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"os"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// createDataDir creates the data directory for the SQLite database if it
// does not already exist. The directory path is determined by the
// s.Opts.DataDir field. The directory is created with 0750 permissions,
// allowing `read`, `write`, and `execute` for the owner, and `read` and
// `execute` for the group.
//
// Returns:
//   - *sdkErrors.SDKError: An error if the directory creation fails, wrapped
//     in ErrFSDirectoryCreationFailed. Returns nil on success.
func (s *DataStore) createDataDir() *sdkErrors.SDKError {
	err := os.MkdirAll(s.Opts.DataDir, 0750)
	if err != nil {
		return sdkErrors.ErrFSDirectoryCreationFailed.Wrap(err)
	}

	return nil
}
