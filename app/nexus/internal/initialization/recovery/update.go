//    \\ SPIKE: Secure your secrets with SPIFFE.
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
	network "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/security/mem"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// mustUpdateRecoveryInfo updates the recovery information by setting a new root
// key and computing new shares. It returns the computed shares.
//
// The function sets the provided root key in the state, computes shares from
// the root secret, performs a sanity check on the computed shares, and ensures
// that temporary variables containing sensitive information are zeroed out
// after use.
//
// This is a critical security function that handles sensitive key material.
//
// Parameters:
//   - rk: A pointer to a 32-byte array containing the new root key
//
// Returns:
//   - []secretsharing.Share: The computed shares for the root secret
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

		// TODO: The sendShardsToKeepers function continues to the next keeper
		// on error. This is reasonable, but consider if all keepers must receive
		// shares for safety.
		// For example, maybe configuration problems should cause a fatal error
		// instead of just bypassing the keeper.
		// Since this is done periodically in `SendShardsPeriodically()`, we
		// can overlook temporary issues.

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

		if share.ID.IsZero() {
			log.Log().Warn(fName,
				"msg", "Failed to find share for keeper", "keeper_id", keeperId)
			continue
		}

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
			//
			// Note that you cannot do `mem.Clear(contribution)` because
			// contribution is a slice, not a struct.
			// When we pass a byte slice s to the function Clear[T any](s *T),
			// we are passing a pointer to the slice header, not a pointer to the
			// underlying array. The slice header contains three fields:
			// * A pointer to the underlying array
			// * The length of the slice
			// * The capacity of the slice
			// mem.Clear(s) will zero out this slice header structure, but not the
			// actual array data the slice points to
			for i := range contribution {
				contribution[i] = 0
			}
			// TODO: maybe a helper function for this too.

			log.Log().Warn(fName,
				"msg", "invalid contribution length",
				"len", len(contribution), "keeper_id", keeperId)
			continue
		}

		scr := reqres.ShardContributionRequest{}

		shard := new([32]byte)
		// Security: shard is intentionally binary (instead of string) for
		// better memory management. Do not change its data type.
		copy(shard[:], contribution)
		scr.Shard = shard

		// Security: Ensure that the contribution is zeroed out before
		// the next iteration.
		for i := range contribution {
			contribution[i] = 0
		}

		md, err := json.Marshal(scr)

		// Security: Erase scr.Shard when no longer in use.
		mem.Clear(scr.Shard)

		if err != nil {
			log.Log().Warn(fName,
				"msg", "Failed to marshal request",
				"err", err, "keeper_id", keeperId)
			continue
		}

		_, err = net.Post(client, u, md)

		// Security: Ensure that the md is zeroed out before
		// the next iteration.
		mem.Clear(&md)

		if err != nil {
			log.Log().Warn(fName, "msg",
				"Failed to post",
				"err", err, "keeper_id", keeperId)
			continue
		}
	}
}
