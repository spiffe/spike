//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"github.com/spiffe/spike-sdk-go/api/errors"
)

// validateShard checks if the shard is valid and not duplicate
func validateShard(shard *[32]byte) error {
	// TODO: in a valid Shamir Secret Sharing scheme, what matters is both the
	// shard ID and its value.
	// Two different shards (with different IDs) could legitimately have the same
	// value in some cases. Similarly, an attacker could submit the same value
	// with different claimed IDs to potentially manipulate the reconstruction.
	// TODO: first check for dupe id (most common)
	// TODO: then check for the content duplicity.
	// TODO: additionally ensure that shard is not all zeroes

	// Check if shard is already stored
	shardsMutex.RLock()
	defer shardsMutex.RUnlock() // Ensure lock is always released

	// Check for duplicates
	for _, existingShard := range shards {
		// Assume this is a duplicate until we find a difference
		isDuplicate := true

		// Check each byte - if any byte differs, it's not a duplicate
		for i := range shard {
			if existingShard.Value[i] != shard[i] {
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
