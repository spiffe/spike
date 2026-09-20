//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\ SPDX-License-Identifier: Apache-2.0

package integration

import (
	"os"

	"github.com/spiffe/spike-sdk-go/log"
)

// removeTempDir deletes the per-run data directory.
//
// A failure here does not change the test verdict, but it must not pass
// silently either, so it is logged as a warning.
//
// Parameters:
//   - dir: The temporary data directory to remove.
func removeTempDir(dir string) {
	if rmErr := os.RemoveAll(dir); rmErr != nil {
		log.Warn("failed to remove the temporary data directory",
			"dir", dir, "err", rmErr)
	}
}
