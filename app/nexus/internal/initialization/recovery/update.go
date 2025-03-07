//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/json"
	"net/url"

	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"
	network "github.com/spiffe/spike-sdk-go/net"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

func mustUpdateRecoveryInfo(rk *[32]byte) []secretsharing.Share {
	const fName = "mustUpdateRecoveryInfo"
	log.Log().Info(fName, "msg", "Updating recovery info")

	// Save recovery information.
	state.SetRootKey(rk)

	rootSecret, rootShares := computeShares()
	sanityCheck(rootSecret, rootShares)
	// Security: Ensure that temporary variables are zeroed out.
	defer func() {
		rootSecret.SetUint64(0)
	}()

	return rootShares
}

// sendShardsToKeepers distributes shares of the root key to all keeper nodes.
// Note that we recompute shares for each keeper rather than computing them once
// and distributing them. This is safe because:
//  1. computeShares() uses a deterministic random reader seeded with the
//     root key
//  2. Given the same root key, it will always produce identical shares
//  3. findShare() ensures each keeper receives its designated share
//     This approach simplifies the code flow and maintains consistency across
//     potential system restarts or failures.
func sendShardsToKeepers(
	source *workloadapi.X509Source, keepers map[string]string,
) {
	const fName = "sendShardsToKeepers"

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

		client, err := network.CreateMtlsClientWithPredicate(
			source, auth.IsKeeper,
		)

		if err != nil {
			log.Log().Warn(fName,
				"msg", "Failed to create mTLS client",
				"err", err)
			continue
		}

		if state.RootKeyZero() {
			log.Log().Info(fName, "msg", "rootKey is zero; moving on...")
			continue
		}

		rootSecret, rootShares := computeShares()
		sanityCheck(rootSecret, rootShares)

		share := findShare(keeperId, keepers, rootShares)

		rootSecret.SetUint64(0)
		// Security: Ensure that the rootShares are zeroed out before
		// the function returns.
		for i := range rootShares {
			rootShares[i].Value.SetUint64(0)
		}

		contribution, err := share.Value.MarshalBinary()

		// Security: Ensure that the share is zeroed out before
		// the next iteration.
		share.Value.SetUint64(0)

		if err != nil {
			// Security: Ensure that the contribution is zeroed out before
			// the next iteration.
			for i := range contribution {
				contribution[i] = 0
			}

			log.Log().Warn(fName,
				"msg", "Failed to marshal share",
				"err", err, "keeper_id", keeperId)
			continue
		}

		if len(contribution) != 32 {
			// Security: Ensure that the contribution is zeroed out before
			// the next iteration.
			for i := range contribution {
				contribution[i] = 0
			}

			log.Log().Warn(fName,
				"msg", "invalid contribution length",
				"len", len(contribution), "keeper_id", keeperId)
			continue
		}

		scr := reqres.ShardContributionRequest{
			KeeperId: keeperId,
		}

		// Security: shard is intentionally binary (instead of string) for
		// better memory management. Do not change its data type.
		for i, b := range contribution {
			scr.Shard[i] = b
		}

		// Security: Ensure that the contribution is zeroed out before
		// the next iteration.
		for i := range contribution {
			contribution[i] = 0
		}

		md, err := json.Marshal(scr)

		// Security: Erase scr.Shard when no longer in use.
		for i := range scr.Shard {
			scr.Shard[i] = 0
		}

		if err != nil {
			log.Log().Warn(fName,
				"msg", "Failed to marshal request",
				"err", err, "keeper_id", keeperId)
			continue
		}

		_, err = net.Post(client, u, md)

		// Security: Ensure that the md is zeroed out before
		// the next iteration.
		for i := range md {
			md[i] = 0
		}

		if err != nil {
			log.Log().Warn(fName, "msg",
				"Failed to post",
				"err", err, "keeper_id", keeperId)
			continue
		}
	}
}
