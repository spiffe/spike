//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package state provides thread-safe utilities for securely managing
// and accessing a global shard value. It ensures consistent access
// and updates to the shard using synchronization primitives.
package state

import (
	"sync"
)

var shard [32]byte
var shardMutex sync.RWMutex

// SetShard safely updates the global shard value under a write lock.
// Although the value is a pointer type, it creates a copy. The value `s`
// can be safely erased after calling `SetShard()`.
//
// Parameters:
//   - s *[32]byte: Pointer to the new shard value to store
//
// Thread-safe through shardMutex.
func SetShard(s *[32]byte) {
	shardMutex.Lock()
	defer shardMutex.Unlock()

	zeroed := true
	for i := range s {
		if s[i] != 0 {
			zeroed = false
			break
		}
	}

	// Do not reset the shard if the new value is zero.
	if zeroed {
		return
	}

	copy(shard[:], s[:])
}

// Shard safely retrieves the current global shard value under a read lock.
// Although this function returns a pointer, it is intended to be used for
// read-only access. Do not mutate the value that this function's return value
// points at. If you want to change the shard, use `SetShard()` instead.
//
// Returns:
//   - *[32]byte: Pointer to the current shard value
//
// Thread-safe through shardMutex.
func Shard() *[32]byte {
	shardMutex.RLock()
	defer shardMutex.RUnlock()

	// Security: return a reference, not a copy.
	return &shard
}
