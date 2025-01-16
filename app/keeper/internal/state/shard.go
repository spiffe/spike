//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"sync"
)

var shard []byte
var shardMutex sync.RWMutex

// SetShard safely updates the global shard value under a write lock.
//
// Parameters:
//   - s []byte: New shard value to store
//
// Thread-safe through shardMutex.
func SetShard(s []byte) {
	shardMutex.Lock()
	defer shardMutex.Unlock()
	shard = s
}

// Shard safely retrieves the current global shard value under a read lock.
//
// Returns:
//   - []byte: Current shard value
//
// Thread-safe through shardMutex.
func Shard() []byte {
	shardMutex.RLock()
	defer shardMutex.RUnlock()
	return shard
}
