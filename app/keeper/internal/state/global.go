//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"sync"

	"github.com/spiffe/spike-sdk-go/crypto"
)

// shard holds the Shamir secret share assigned to this SPIKE Keeper instance.
// SPIKE Nexus distributes shards of the root key to multiple Keeper instances
// using Shamir's Secret Sharing scheme. Each Keeper stores exactly one shard.
// When SPIKE Nexus needs to recover its root key, it collects shards from
// multiple Keepers and reconstructs the original secret.
//
// This variable must only be accessed through the exported functions in this
// package (SetShard, ShardNoSync, RLockShard, RUnlockShard) to ensure thread
// safety.
var shard [crypto.AES256KeySize]byte

// shardMutex protects concurrent access to the shard variable. Use RLockShard
// and RUnlockShard for read access, and the mutex is locked internally by
// SetShard for write access.
var shardMutex sync.RWMutex
