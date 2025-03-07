//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"github.com/spiffe/spike-sdk-go/api/errors"
)

// validateShard checks if the shard is valid and not duplicate
func validateShard(shard *[32]byte) error {
	// Check if shard is already stored
	shardsMutex.RLock()
	defer shardsMutex.RUnlock() // Ensure lock is always released

	// Check for duplicates
	for _, existingShard := range shards {
		// Assume this is a duplicate until we find a difference
		isDuplicate := true

		// Check each byte - if any byte differs, it's not a duplicate
		for i := range shard {
			if existingShard[i] != shard[i] {
				isDuplicate = false
				break
			}
		}

		// If we went through all bytes and found no differences, it's a duplicate
		if isDuplicate {
			return errors.ErrInvalidInput
		}
	}

	return nil
}
