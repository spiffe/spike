//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

// resetShard resets the shard to the zero state for testing.
func resetShard() {
	shardMutex.Lock()
	defer shardMutex.Unlock()
	for i := range shard {
		shard[i] = 0
	}
}

// lowByte returns the low eight bits of an integer.
//
// Test patterns only need a value that varies per index; masking keeps the
// conversion in range without a narrowing cast.
//
// Parameters:
//   - n: The integer to reduce.
//
// Returns:
//   - byte: The low eight bits of n.
func lowByte(n int) byte {
	return byte(n & 0xff)
}
