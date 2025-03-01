//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"encoding/base64"

	"github.com/spiffe/spike-sdk-go/api/errors"
)

// validateShard checks if the shard is valid and not duplicate
func validateShard(shard string) error {
	// Check if shard is already stored
	shardsMutex.RLock()
	for _, existingShard := range shards {
		if existingShard == shard {
			shardsMutex.RUnlock()
			return errors.ErrInvalidInput
		}
	}
	shardsMutex.RUnlock()

	// Validate shard length
	decodedShard, err := base64.StdEncoding.DecodeString(shard)
	if err != nil {
		return errors.ErrInvalidInput
	}
	if len(decodedShard) != decodedShardSize {
		return errors.ErrInvalidInput
	}

	return nil
}
