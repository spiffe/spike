//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package state provides thread-safe utilities for securely managing
// and accessing a global shard value. It ensures consistent access
// and updates to the shard using synchronization primitives.
package state

import (
	"sync"

	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
)

var shard [crypto.AES256KeySize]byte
var shardMutex sync.RWMutex

// SetShard safely updates the global shard value under a write lock.
// Although the value is a pointer type, it creates a copy. The value `s`
// can be safely erased after calling `SetShard()`.
//
// Parameters:
//   - s *[32]byte: Pointer to the new shard value to store
//
// Thread-safe through shardMutex.
func SetShard(s *[crypto.AES256KeySize]byte) {
	const fName = "SetShard"

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
		log.Log().Info(fName, "message", "zero value: skipping setting shard")
		return
	}

	copy(shard[:], s[:])
}

// ShardNoSync returns a pointer to the shard without acquiring any locks.
// Callers must ensure proper synchronization by using RLockShard and
// RUnlockShard when accessing the returned pointer.
func ShardNoSync() *[crypto.AES256KeySize]byte {
	return &shard
}

// RLockShard acquires a read lock on the shard mutex.
// This should be paired with a corresponding call to RUnlockShard,
// typically using `defer`.
func RLockShard() {
	shardMutex.RLock()
}

// RUnlockShard releases a read lock on the shard mutex.
// This should only be called after a corresponding call to RLockShard.
func RUnlockShard() {
	shardMutex.RUnlock()
}
