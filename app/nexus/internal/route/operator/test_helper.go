//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
)

// Helper functions

func resetShards() {
	shardsMutex.Lock()
	defer shardsMutex.Unlock()
	shards = []recovery.ShamirShard{}
}

func createTestShardValue(id int) *[crypto.AES256KeySize]byte {
	value := &[crypto.AES256KeySize]byte{}
	// Fill with deterministic test data
	for i := range value {
		value[i] = byte((id*100 + i) % 256)
	}
	// Ensure the first byte is non-zero for validation
	value[0] = byte(id)
	return value
}
