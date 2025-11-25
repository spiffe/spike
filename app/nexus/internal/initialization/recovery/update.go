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
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
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
			warnErr := sdkErrors.ErrAPIBadRequest.Wrap(err)
			warnErr.Msg = "failed to join path"
			log.WarnErr(fName, *warnErr)
			continue
		}

		if state.RootKeyZero() {
			log.Warn(fName, "message", "rootKey is zero: moving on")
			continue
		}

		rootSecret, rootShares := computeShares()
		sanityCheck(rootSecret, rootShares)

		var share secretsharing.Share
		for _, sr := range rootShares {
			kid, err := strconv.Atoi(keeperID)
			if err != nil {
				warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
				warnErr.Msg = "failed to convert keeper id to int"
				log.WarnErr(fName, *warnErr)
				continue
			}

			if sr.ID.IsEqual(group.P256.NewScalar().SetUint64(uint64(kid))) {
				share = sr
				break
			}
		}

		if share.ID.IsZero() {
			warnErr := *sdkErrors.ErrEntityNotFound // copy
			warnErr.Msg = "failed to find share for keeper"
			log.WarnErr(fName, warnErr)
			continue
		}

		rootSecret.SetUint64(0)

		contribution, err := share.Value.MarshalBinary()
		if err != nil {
			warnErr := sdkErrors.ErrDataMarshalFailure.Wrap(err)
			warnErr.Msg = "failed to marshal share"
			log.WarnErr(fName, *warnErr)

			// Security: Ensure sensitive data is zeroed out.
			mem.ClearBytes(contribution)
			share.Value.SetUint64(0)
			for i := range rootShares {
				rootShares[i].Value.SetUint64(0)
			}
			continue
		}

		if len(contribution) != crypto.AES256KeySize {
			// Log before clearing (contribution length is needed for logging).
			warnErr := *sdkErrors.ErrDataInvalidInput // copy
			warnErr.Msg = "invalid contribution length"
			log.WarnErr(fName, warnErr)

			// Security: Ensure sensitive data is zeroed out.
			// Note: use mem.ClearBytes() for slices, not mem.ClearRawBytes().
			mem.ClearBytes(contribution)
			share.Value.SetUint64(0)
			for i := range rootShares {
				rootShares[i].Value.SetUint64(0)
			}
			continue
		}

		scr := reqres.ShardPutRequest{}

		shard := new([crypto.AES256KeySize]byte)
		// Security: shard is intentionally binary (instead of string) for
		// better memory management. Do not change its data type.
		copy(shard[:], contribution)
		scr.Shard = shard

		md, err := json.Marshal(scr)

		// Security: Erase sensitive data when no longer in use.
		mem.ClearRawBytes(scr.Shard)
		mem.ClearBytes(contribution)
		share.Value.SetUint64(0)
		for i := range rootShares {
			rootShares[i].Value.SetUint64(0)
		}

		if err != nil {
			warnErr := sdkErrors.ErrDataMarshalFailure.Wrap(err)
			warnErr.Msg = "failed to marshal request"
			log.WarnErr(fName, *warnErr)
			continue
		}

		// Security: Only SPIKE Keeper can send shards to SPIKE Nexus.
		// Create the client just before use to avoid unnecessary allocation
		// if earlier checks fail.
		client := network.CreateMTLSClientWithPredicate(
			source, predicate.AllowKeeper,
		)

		_, err = net.Post(client, u, md)

		// Security: Ensure that md is zeroed out.
		mem.ClearBytes(md)

		if err != nil {
			warnErr := sdkErrors.ErrAPIPostFailed.Wrap(err)
			warnErr.Msg = "failed to post shard to keeper"
			log.WarnErr(fName, *warnErr)
			continue
		}
	}
}
