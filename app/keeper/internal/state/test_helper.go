//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

// Helper function to reset shard to zero state for testing
func resetShard() {
	shardMutex.Lock()
	defer shardMutex.Unlock()
	for i := range shard {
		shard[i] = 0
	}
}
