//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"github.com/spiffe/spike/pkg/crypto"
	"sync"
)

// WaitForShards blocks until exactly 3 shards are collected in the global
// Shards map. Once collected, it computes the final key, generates shares,
// sets the internal shard, and performs validation checks.
//
// The function:
// - Polls the Shards map every 2 seconds until 3 shards are present
// - Panics if more than 3 shards are received
// - Processes the shards to generate the final distributed secret
//
// Panics:
//   - If more than 3 shards are received
//func WaitForShards() {
//	for {
//		shardCount := 0
//		Shards.Range(func(key, value any) bool {
//			shardCount++
//			return true
//		})
//
//		log.Log().Info(
//			"waitForShards", "msg", "Current shard count", "count", shardCount,
//		)
//
//		if shardCount < 3 {
//			time.Sleep(2 * time.Second)
//			continue
//		}
//
//		if shardCount > 3 {
//			// TODO: add an audit log, because this is a security incident likely.
//			log.FatalLn("waitForShards: Too many shards received")
//		}
//
//		finalKey := computeFinalKey()
//		secret, shares := computeShares(finalKey)
//		setInternalShard(shares)
//		sanityCheck(secret, shares)
//
//		break
//	}
//}

var myContribution []byte
var myContributionLock sync.Mutex

// RandomContribution generates and caches a random contribution for the
// distributed secret. The contribution is generated only once and reused for
// subsequent calls.
//
// Returns:
//   - []byte: Random contribution bytes from AES-256 seed
//
// Thread-safe through myContributionLock mutex.
func RandomContribution() []byte {
	myContributionLock.Lock()
	defer myContributionLock.Unlock()

	if len(myContribution) == 0 {
		mySeed, _ := crypto.Aes256Seed()
		myContribution = []byte(mySeed)

		return myContribution
	}

	return myContribution
}

var Shards sync.Map

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

// EraseIntermediateShards removes all entries from the global Shards map,
// cleaning up intermediate secret sharing data.
//
// Thread-safe through sync.Map operations.
func EraseIntermediateShards() {
	Shards.Range(func(key, value interface{}) bool {
		Shards.Delete(key)
		return true
	})
}
