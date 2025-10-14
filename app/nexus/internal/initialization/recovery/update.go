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
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
	network "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"
	"github.com/spiffe/spike-sdk-go/security/mem"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

// sendShardsToKeepers distributes shares of the root key to all keeper nodes.
// Note that we recompute shares for each keeper rather than computing them once
// and distributing them. This is safe because:
//  1. computeShares() uses a deterministic random reader seeded with the
//     root key
//  2. Given the same root key, it will always produce identical shares
//  3. findShare() ensures each keeper receives its designated share
//     This approach simplifies the code flow and maintains consistency across
//     potential system restarts or failures.
//
// Note that sendSharesToKeepers optimistically moves on to the next SPIKE
// Keeper in the list on error. This is okay, because SPIKE Nexus may not
// need all keepers to be healthy all at once, and since we periodically
// send shards to keepers, provided there is no configuration mistake,
// all SPIKE Keepers will get their shards eventually.
func sendShardsToKeepers(
	source *workloadapi.X509Source, keepers map[string]string,
) {
	const fName = "sendShardsToKeepers"

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

		// Security: Only SPIKE Keeper can send shards to SPIKE Nexus
		client, err := network.CreateMTLSClientWithPredicate(
			source, predicate.AllowKeeper,
		)

		if err != nil {
			log.Log().Warn(fName,
				"message", "Failed to create mTLS client",
				"err", err)
			continue
		}

		if state.RootKeyZero() {
			log.Log().Warn(fName, "message", "rootKey is zero; moving on...")
			continue
		}

		rootSecret, rootShares := computeShares()
		sanityCheck(rootSecret, rootShares)

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

		if share.ID.IsZero() {
			log.Log().Warn(fName,
				"message", "Failed to find share for keeper", "keeper_id", keeperID)
			continue
		}

		rootSecret.SetUint64(0)

		contribution, err := share.Value.MarshalBinary()
		if err != nil {
			log.Log().Warn(fName, "message", "Failed to marshal share",
				"err", err, "keeper_id", keeperID)

			// Security: Ensure that the contribution is zeroed out before
			// the next iteration.
			mem.ClearBytes(contribution)

			// Security: Ensure that the share is zeroed out before
			// the next iteration.
			share.Value.SetUint64(0)

			// Security: Ensure that the rootShares are zeroed out before
			// the function returns.
			for i := range rootShares {
				rootShares[i].Value.SetUint64(0)
			}

			log.Log().Warn(fName,
				"message", "Failed to marshal share",
				"err", err, "keeper_id", keeperID)
			continue
		}

		if len(contribution) != crypto.AES256KeySize {
			// Security: Ensure that the contribution is zeroed out before
			// the next iteration.
			//
			// Note that you cannot do `mem.ClearRawBytes(contribution)` because
			// the contribution is a slice, not a struct; we use `mem.ClearBytes()`
			// instead.
			mem.ClearBytes(contribution)

			// Security: Ensure that the share is zeroed out before
			// the next iteration.
			share.Value.SetUint64(0)

			// Security: Ensure that the rootShares are zeroed out before
			// the function returns.
			for i := range rootShares {
				rootShares[i].Value.SetUint64(0)
			}

			log.Log().Warn(fName,
				"message", "invalid contribution length",
				"len", len(contribution), "keeper_id", keeperID)
			continue
		}

		scr := reqres.ShardPutRequest{}

		shard := new([crypto.AES256KeySize]byte)
		// Security: shard is intentionally binary (instead of string) for
		// better memory management. Do not change its data type.
		copy(shard[:], contribution)
		scr.Shard = shard

		md, err := json.Marshal(scr)

		// Security: Erase scr.Shard when no longer in use.
		mem.ClearRawBytes(scr.Shard)

		// Security: Ensure that the contribution is zeroed out before
		// the next iteration.
		mem.ClearBytes(contribution)

		// Security: Ensure that the share is zeroed out before
		// the next iteration.
		share.Value.SetUint64(0)

		// Security: Ensure that the rootShares are zeroed out before
		// the function returns.
		for i := range rootShares {
			rootShares[i].Value.SetUint64(0)
		}

		if err != nil {
			log.Log().Warn(fName,
				"message", "Failed to marshal request",
				"err", err, "keeper_id", keeperID)
			continue
		}

		_, err = net.Post(client, u, md)
		// Security: Ensure that the md is zeroed out before
		// the next iteration.
		mem.ClearBytes(md)

		if err != nil {
			log.Log().Warn(fName, "message",
				"Failed to post",
				"err", err, "keeper_id", keeperID)
			continue
		}
	}
}
