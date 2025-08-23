//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"strconv"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// TODO: move ShardSize const to SDK so we can refer it from a central location.

// iterateKeepersAndInitializeState retrieves Shamir secret shards from multiple
// SPIKE Keeper instances and attempts to reconstruct the root key when a
// threshold number of shards is collected.
//
// The function iterates through all configured keepers, requesting their shards
// via SPIFFE mTLS. Once the Shamir threshold is reached, it reconstructs the
// root key and initializes the system state. This function implements secure
// memory handling, ensuring sensitive data is cleared after use.
//
// Parameters:
//   - source: An X.509 source for mTLS authentication when communicating with
//     keeper services
//   - successfulKeeperShards: A map storing successfully retrieved shards
//     indexed by keeper ID. Each shard is a fixed-size byte array of size
//     shardSize
//
// Returns:
//   - bool: true if the system was successfully initialized with the
//     reconstructed root key, false if initialization failed or insufficient
//     shards were collected
//
// Security considerations:
//   - All sensitive data (shards, root key) is securely erased from memory
//     after use
//   - The function will fatal log and terminate if keeper IDs cannot be
//     converted to integers
//   - Shards are validated to ensure they are not zeroed before being accepted
//
// The function performs the following steps:
//  1. Iterates through all configured keepers from env.Keepers()
//  2. For each keeper, requests its shard via HTTP using mTLS authentication
//  3. Validates and stores successful shard responses
//  4. When the threshold is reached, reconstructs the root key using Shamir's
//     Secret Sharing
//  5. Initializes the system state with the recovered root key
//  6. Securely clears all sensitive data from memory
func iterateKeepersAndInitializeState(
	source *workloadapi.X509Source,
	successfulKeeperShards map[string]*[shardSize]byte,
) bool {
	const fName = "iterateKeepersAndInitializeState"

	// TODO: log and exit if "in memory" mode.

	for keeperID, keeperAPIRoot := range env.Keepers() {
		log.Log().Info(fName, "id", keeperID, "url", keeperAPIRoot)

		u := shardURL(keeperAPIRoot)
		if u == "" {
			continue
		}

		data := shardResponse(source, u)
		if len(data) == 0 {
			continue
		}

		res := unmarshalShardResponse(data)
		// Security: Reset data before the function exits.
		mem.ClearBytes(data)

		if res == nil {
			continue
		}

		if mem.Zeroed32(res.Shard) {
			log.Log().Info(fName, "message", "Shard is zeroed")
			continue
		}

		successfulKeeperShards[keeperID] = res.Shard
		if len(successfulKeeperShards) != env.ShamirThreshold() {
			continue
		}

		// No need to erase `ss` because upon successful recovery,
		// `InitializeBackingStoreFromKeepers()` resets `successfulKeeperShards`
		// which points to the same shards here. And until recovery, we will keep
		// a threshold number of shards in memory.
		ss := make([]ShamirShard, 0)
		for ix, shard := range successfulKeeperShards {
			id, err := strconv.Atoi(ix)
			if err != nil {
				// This is a configuration error; we cannot recover from it,
				// and it may cause further security issues. Crash immediately.
				log.FatalLn(
					fName, "message", "Failed to convert keeper ID to int", "err", err,
				)
				continue
			}

			ss = append(ss, ShamirShard{
				ID:    uint64(id),
				Value: shard,
			})
		}

		rk := RecoverRootKey(ss)

		// TODO: bail if rk is nil or empty.

		// Both of these methods directly or indirectly make a copy of `rk`
		// It is okay to zero out `rk` after calling these two functions.
		state.Initialize(rk)

		// Security: Zero out temporary variables before the function exits.
		mem.ClearRawBytes(rk)
		// Security: Zero out temporary variables before the function exits.
		// Note that `successfulKeeperShards` will be reset elsewhere.
		mem.ClearRawBytes(res.Shard)

		// System initialized: Exit infinite loop.
		return true
	}

	return false
}
