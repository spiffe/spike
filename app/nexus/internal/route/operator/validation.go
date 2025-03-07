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

	hasSameShard := true
	for _, existingShard := range shards {
		for i := range shard {
			if existingShard[i] != shard[i] {
				hasSameShard = false
				break
			}
		}
	}
	if hasSameShard {
		return errors.ErrInvalidInput
	}

	shardsMutex.RUnlock()

	return nil
}
