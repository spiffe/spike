//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"strconv"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

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
//     32.
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
	successfulKeeperShards map[string]*[crypto.AES256KeySize]byte,
) bool {
	const fName = "iterateKeepersAndInitializeState"

	// In memory mode, no recovery is needed regardless of source availability
	if env.BackendStoreTypeVal() == env.Memory {
		log.Warn(fName, "message", "in memory mode: skipping recovery")
		return true
	}

	// For persistent backends, X509 source is required for mTLS with keepers.
	// We warn and return false (triggering retry) rather than crashing because:
	// 1. This function runs in retry.Forever() (designed for transient failures)
	// 2. Workload API may temporarily lose source and recover
	// 3. Returning false allows the system to retry and recover gracefully
	if source == nil {
		failErr := sdkErrors.ErrSPIFFENilX509Source.Clone()
		failErr.Msg = "X509 source is nil, cannot perform mTLS with keepers"
		log.WarnErr(fName, *failErr)
		return false
	}

	for keeperID, keeperAPIRoot := range env.KeepersVal() {
		log.Info(
			fName,
			"message", "iterating keepers",
			"id", keeperID, "url", keeperAPIRoot,
		)

		u := url.ShardFromKeeperAPIRoot(keeperAPIRoot)
		data, err := shardGetResponse(source, u)
		if err != nil {
			warnErr := sdkErrors.ErrNetPeerConnection.Wrap(err)
			warnErr.Msg = "failed to get shard from keeper: " + u
			log.WarnErr(fName, *warnErr) // just log: will retry
			continue
		}

		res, unmarshalErr := unmarshalShardResponse(data)
		// Security: Reset data before the function exits.
		mem.ClearBytes(data)
		if unmarshalErr != nil {
			failErr := unmarshalErr.Clone()
			failErr.Msg = "failed to unmarshal shard response"
			log.WarnErr(fName, *failErr) // just log: will retry
			continue
		}

		if mem.Zeroed32(res.Shard) {
			warnErr := *sdkErrors.ErrShamirEmptyShard.Clone()
			warnErr.Msg = "shard is zeroed"
			log.WarnErr(fName, warnErr) // just log: will retry
			continue
		}

		successfulKeeperShards[keeperID] = res.Shard
		if len(successfulKeeperShards) != env.ShamirThresholdVal() {
			log.Info(
				fName,
				"message", "still shards remaining",
				"id", keeperID,
				"url", keeperAPIRoot,
				"has", len(successfulKeeperShards),
				"needs", env.ShamirThresholdVal(),
			)
			continue
		}
		log.Info(fName, "message", "threshold reached")

		// No need to erase `ss` because upon successful recovery,
		// `InitializeBackingStoreFromKeepers()` resets `successfulKeeperShards`
		// which points to the same shards here. And until recovery, we will keep
		// a threshold number of shards in memory.
		ss := make([]crypto.ShamirShard, 0)
		for ix, shard := range successfulKeeperShards {
			id, err := strconv.Atoi(ix)
			if err != nil {
				// Unlike URL misconfiguration (which we tolerate above), an
				// unparseable keeper ID is fatal because:
				// 1. We've already collected threshold shards. Skipping one now
				//    means we'd need to re-fetch, but the same ID will fail
				//    again.
				// 2. The keeper ID is used as the Shamir shard index. Using a
				//    wrong index produces an incorrect root key, which is worse
				//    than crashing.
				// 3. This same ID was used during bootstrap to store the shard.
				//    If it was valid then but invalid now, the configuration
				//    has been corrupted.
				failErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
				failErr.Msg = "failed to convert keeper ID to int"
				log.FatalErr(fName, *failErr)
				return false
			}

			ss = append(ss, crypto.ShamirShard{
				ID:    uint64(id),
				Value: shard,
			})
		}

		rk := crypto.ComputeRootKeyFromShards(ss)

		// Security: Crash if there is a problem with root key recovery.
		if rk == nil || mem.Zeroed32(rk) {
			failErr := *sdkErrors.ErrShamirReconstructionFailed.Clone()
			failErr.Msg = "failed to recover the root key"
			log.FatalErr(fName, failErr)
		}

		// It is okay to zero out `rk` after calling this function because we
		// make a copy of rk.
		state.Initialize(rk)

		// Security: Zero out temporary variables before the function exits.
		mem.ClearRawBytes(rk)
		// Security: Zero out temporary variables before the function exits.
		// Note that `successfulKeeperShards` will be reset elsewhere.
		mem.ClearRawBytes(res.Shard)

		// System initialized: Exit loop.
		return true
	}

	// Failed to initialize.
	return false
}
