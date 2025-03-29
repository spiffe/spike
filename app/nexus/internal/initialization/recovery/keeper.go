//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"
	"net/url"
	"os"
	"strconv"

	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

func iterateKeepersToBootstrap(
	keepers map[string]string, rootShares []secretsharing.Share,
	successfulKeepers map[string]bool, source *workloadapi.X509Source,
) bool {
	const fName = "iterateKeepersToBootstrap"

	for keeperId, keeperApiRoot := range keepers {
		u, err := url.JoinPath(
			keeperApiRoot, string(apiUrl.SpikeKeeperUrlContribute),
		)
		if err != nil {
			log.Log().Warn(
				fName, "msg", "Failed to join path", "url", keeperApiRoot,
			)
			continue
		}

		var share secretsharing.Share

		for _, sr := range rootShares {
			kid, err := strconv.Atoi(keeperId)
			if err != nil {
				log.Log().Warn(
					fName, "msg", "Failed to convert keeper id to int", "err", err)
				continue
			}

			if sr.ID.IsEqual(group.P256.NewScalar().SetUint64(uint64(kid))) {
				share = sr
				break
			}
		}

		// If initialized, IDs start from 1. Zero means there is no match.
		if share.ID.IsZero() {
			log.Log().Info(fName, "msg",
				"Failed to find share for keeper", "keeper_id", keeperId)
			continue
		}

		contribution, err := share.Value.MarshalBinary()
		if err != nil {
			log.Log().Info(fName, "msg",
				"Failed to marshal share", "err", err, "keeper_id", keeperId)
			continue
		}

		data := shardContributionResponse(u, &contribution, source)
		if len(data) == 0 {
			// Security: Ensure that the share is zeroed out
			// before the function returns.
			mem.Clear(&contribution)

			log.Log().Info(fName, "msg", "No data; moving on...")
			continue
		}
		// Security: Ensure that the share is zeroed out
		// before the function returns.
		mem.Clear(&contribution)

		var res reqres.ShardContributionResponse
		err = json.Unmarshal(data, &res)
		if err != nil {
			log.Log().Info(fName, "msg", "Failed to unmarshal response", "err", err)
			continue
		}

		successfulKeepers[keeperId] = true
		log.Log().Info(fName, "msg", "Success", "keeper_id", keeperId)

		if len(successfulKeepers) == env.ShamirShares() {
			log.Log().Info(fName, "msg", "All keepers initialized")

			tombstone := config.SpikeNexusTombstonePath()

			// Create the tombstone file to mark SPIKE Nexus as bootstrapped.
			// 0600 to align with principle of least privilege. We can change the
			// permission fi it doesn't work out.
			err = os.WriteFile(
				tombstone,
				[]byte("spike.nexus.bootstrapped=true"), 0600,
			)
			if err != nil {
				// Although the tombstone file is just a marker, it's still important.
				// If SPIKE Nexus cannot create the tombstone file, or if someone
				// deletes it, this can slightly change system operations.
				// To be on the safe side, we let SPIKE Nexus crash because not being
				// able to write to the data volume (where the tombstone file would be)
				// can be a precursor of other problems that can affect the reliability
				// of the backing store.
				log.FatalLn(fName + ": failed to create tombstone file: " + err.Error())
			}

			log.Log().Info(fName, "msg", "Tombstone file created successfully")
			return true
		}
	}

	return false
}

func iterateKeepersAndTryRecovery(
	source *workloadapi.X509Source,
	successfulKeeperShards map[string]*[32]byte,
) bool {
	const fName = "iterateKeepersAndTryRecovery"

	for keeperId, keeperApiRoot := range env.Keepers() {
		log.Log().Info(fName, "id", keeperId, "url", keeperApiRoot)

		u := shardUrl(keeperApiRoot)
		if u == "" {
			continue
		}

		data := shardResponse(source, u)
		if len(data) == 0 {
			continue
		}

		res := unmarshalShardResponse(data)
		// Security: Reset data before the function exits.
		mem.Clear(&data)

		if res == nil {
			continue
		}

		if mem.Zeroed32(res.Shard) {
			log.Log().Info(fName, "msg", "Shard is zeroed")
			continue
		}

		successfulKeeperShards[keeperId] = res.Shard
		if len(successfulKeeperShards) != env.ShamirThreshold() {
			continue
		}

		// No need to erase `ss` because upon successful recovery,
		// `RecoverBackingStoreUsingKeeperShards()` resets `successfulKeeperShards`
		// which points to the same shards here. And until recovery, we will keep
		// a threshold number of shards in memory.
		ss := make([]ShamirShard, 0)
		for ix, shard := range successfulKeeperShards {
			id, err := strconv.Atoi(ix)
			if err != nil {
				// This is a configuration error; we cannot recover from it,
				// and it may cause further security issues. Crash immediately.
				log.FatalF(
					fName, "msg", "Failed to convert keeper Id to int", "err", err,
				)
				continue
			}

			ss = append(ss, ShamirShard{
				Id:    uint64(id),
				Value: shard,
			})
		}

		binaryRec := RecoverRootKey(ss)

		// Both of these methods directly or indirectly make a copy of `binaryRec`
		// It is okay to zero out `binaryRec` after calling these two functions.
		state.Initialize(binaryRec)
		state.SetRootKey(binaryRec)

		// Security: Zero out temporary variables before function exits.
		mem.Clear(binaryRec)
		// Security: Zero out temporary variables before function exits.
		// Note that `successfulKeeperShards` will be reset elsewhere.
		mem.Clear(res.Shard)

		// System initialized: Exit infinite loop.
		return true
	}

	return false
}
