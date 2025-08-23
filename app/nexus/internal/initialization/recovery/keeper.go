//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

func iterateKeepersToBootstrap(
	keepers map[string]string, rootShares []secretsharing.Share,
	successfulKeepers map[string]bool, source *workloadapi.X509Source,
) bool {
	const fName = "iterateKeepersToBootstrap"

	for keeperID, keeperAPIRoot := range keepers {
		u, err := url.JoinPath(
			keeperAPIRoot, string(apiUrl.KeeperContribute),
		)
		if err != nil {
			log.Log().Warn(
				fName, "message", "Failed to join path", "url", keeperAPIRoot,
			)
			continue
		}

		var share secretsharing.Share

		for _, sr := range rootShares {
			kid, err := strconv.Atoi(keeperID)
			if err != nil {
				log.Log().Warn(
					fName, "message", "Failed to convert keeper id to int", "err", err)
				continue
			}

			if sr.ID.IsEqual(group.P256.NewScalar().SetUint64(uint64(kid))) {
				share = sr
				break
			}
		}

		// If initialized, IDs start from 1. Zero means there is no match.
		if share.ID.IsZero() {
			log.Log().Info(fName, "message",
				"Failed to find share for keeper", "keeper_id", keeperID)
			continue
		}

		contribution, err := share.Value.MarshalBinary()
		if err != nil {
			log.Log().Info(fName, "message",
				"Failed to marshal share", "err", err, "keeper_id", keeperID)
			continue
		}

		data := shardContributionResponse(u, &contribution, source)
		if len(data) == 0 {
			// Security: Ensure that the share is zeroed out
			// before the function returns.
			mem.ClearBytes(contribution)

			log.Log().Info(fName, "message", "No data; moving on...")
			continue
		}

		// Security: Ensure that the share is zeroed out
		// before the function returns.
		mem.ClearBytes(contribution)

		var res reqres.ShardContributionResponse
		err = json.Unmarshal(data, &res)
		if err != nil {
			log.Log().Info(fName,
				"message", "Failed to unmarshal response", "err", err)
			continue
		}

		successfulKeepers[keeperID] = true
		log.Log().Info(fName, "message", "Success", "keeper_id", keeperID)

		if len(successfulKeepers) == env.ShamirShares() {
			log.Log().Info(fName, "message", "All keepers initialized")
			return true
		}
	}

	return false
}

func iterateKeepersAndInitializeState(
	source *workloadapi.X509Source,
	successfulKeeperShards map[string]*[shardSize]byte,
) bool {
	const fName = "iterateKeepersAndInitializeState"

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

		binaryRec := RecoverRootKey(ss)

		// Both of these methods directly or indirectly make a copy of `binaryRec`
		// It is okay to zero out `binaryRec` after calling these two functions.
		state.Initialize(binaryRec)

		// Security: Zero out temporary variables before the function exits.
		mem.ClearRawBytes(binaryRec)
		// Security: Zero out temporary variables before the function exits.
		// Note that `successfulKeeperShards` will be reset elsewhere.
		mem.ClearRawBytes(res.Shard)

		// System initialized: Exit infinite loop.
		return true
	}

	return false
}
