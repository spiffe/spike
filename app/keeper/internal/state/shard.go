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

func zeroed(s *[32]byte) bool {
	for i := range s {
		if s[i] != 0 {
			return false
		}
	}
	return true
}

// SetShard safely updates the global shard value under a write lock.
//
// Parameters:
//   - s []byte: New shard value to store
//
// Thread-safe through shardMutex.
func SetShard(s *[32]byte) {
	shardMutex.Lock()
	defer shardMutex.Unlock()

	if zeroed(s) {
		return
	}

	for i := range s {
		shard[i] = s[i]
	}
}

// Shard safely retrieves the current global shard value under a read lock.
//
// Returns:
//   - []byte: Current shard value
//
// Thread-safe through shardMutex.
func Shard() *[32]byte {
	shardMutex.RLock()
	defer shardMutex.RUnlock()

	// Security: return a reference, not a copy.
	return &shard
}
