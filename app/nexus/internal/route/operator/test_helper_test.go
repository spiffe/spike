//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"github.com/spiffe/spike-sdk-go/crypto"
)

// Helper functions

func resetShards() {
	shardsMutex.Lock()
	defer shardsMutex.Unlock()
	shards = []crypto.ShamirShard{}
}

// createTestShardValue fills a shard with deterministic test data derived
// from a shard ID.
//
// The arithmetic is done in bytes on purpose: wrapping is fine for test data,
// and it avoids narrowing conversions.
//
// Parameters:
//   - id: The shard ID that seeds the pattern and becomes the first byte.
//
// Returns:
//   - *[crypto.AES256KeySize]byte: A shard whose first byte is non-zero.
func createTestShardValue(id uint8) *[crypto.AES256KeySize]byte {
	value := &[crypto.AES256KeySize]byte{}
	var offset uint8
	for i := range value {
		value[i] = id*100 + offset
		offset++
	}
	// Ensure the first byte is non-zero for validation
	value[0] = id
	return value
}
