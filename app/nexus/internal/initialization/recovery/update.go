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
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/predicate"
	"github.com/spiffe/spike-sdk-go/security/mem"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// sendShardsToKeepers distributes shares of the root key to all keeper nodes.
// Shares are recomputed for each keeper rather than computed once and
// distributed. This is safe because:
//  1. computeShares() uses a deterministic random reader seeded with the
//     root key
//  2. Given the same root key, it will always produce identical shares
//  3. Each keeper receives its designated share based on keeper ID
//
// This approach simplifies the code flow and maintains consistency across
// potential system restarts or failures.
//
// The function optimistically moves on to the next SPIKE Keeper in the list on
// error. This is acceptable because SPIKE Nexus does not need all keepers to be
// healthy simultaneously. Since shards are sent periodically, all SPIKE Keepers
// will eventually receive their shards provided there is no configuration
// error.
//
// Parameters:
//   - source: X509Source for mTLS authentication with keepers
//   - keepers: Map of keeper IDs to their API root URLs
func sendShardsToKeepers(
	source *workloadapi.X509Source, keepers map[string]string,
) {
	const fName = "sendShardsToKeepers"

	for keeperID, keeperAPIRoot := range keepers {
		u, urlErr := url.JoinPath(
			keeperAPIRoot, string(apiUrl.KeeperContribute),
		)
		if urlErr != nil {
			warnErr := sdkErrors.ErrAPIBadRequest.Wrap(urlErr)
			warnErr.Msg = "failed to join path"
			log.WarnErr(fName, *warnErr)
			continue
		}

		if state.RootKeyZero() {
			log.Warn(fName, "message", "rootKey is zero: moving on")
			continue
		}

		state.LockRootKey()
		rootSecret, rootShares := crypto.ComputeShares(state.RootKeyNoLock())
		// not using `defer` because ComputeShare is deterministic, it does not
		// return an error, and `root key` is not nil. -- using defer in a loop
		// can potentially leak resources. An alternative approach could be to
		// use a closure or create a copy of the root key; both of the approaches
		// complicate the code further.
		state.UnlockRootKey()

		var share secretsharing.Share
		for _, sr := range rootShares {
			kid, atoiErr := strconv.Atoi(keeperID)
			if atoiErr != nil {
				warnErr := sdkErrors.ErrDataInvalidInput.Wrap(atoiErr)
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
			warnErr := *sdkErrors.ErrEntityNotFound.Clone()
			warnErr.Msg = "failed to find share for keeper"
			log.WarnErr(fName, warnErr)
			continue
		}

		rootSecret.SetUint64(0)

		contribution, marshalErr := share.Value.MarshalBinary()
		if marshalErr != nil {
			warnErr := sdkErrors.ErrDataMarshalFailure.Wrap(marshalErr)
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
			warnErr := *sdkErrors.ErrDataInvalidInput.Clone()
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

		md, jsonErr := json.Marshal(scr)

		// Security: Erase sensitive data when no longer in use.
		mem.ClearRawBytes(scr.Shard)
		mem.ClearBytes(contribution)
		share.Value.SetUint64(0)
		for i := range rootShares {
			rootShares[i].Value.SetUint64(0)
		}

		if jsonErr != nil {
			warnErr := sdkErrors.ErrDataMarshalFailure.Wrap(jsonErr)
			warnErr.Msg = "failed to marshal request"
			log.WarnErr(fName, *warnErr)
			continue
		}

		// Security: Only SPIKE Keeper can send shards to SPIKE Nexus.
		// Create the client just before use to avoid unnecessary allocation
		// if earlier checks fail.
		client := net.CreateMTLSClientWithPredicate(
			source, predicate.AllowKeeper,
		)

		_, postErr := net.Post(client, u, md)

		// Security: Ensure that md is zeroed out.
		mem.ClearBytes(md)

		if postErr != nil {
			warnErr := sdkErrors.ErrAPIPostFailed.Wrap(postErr)
			warnErr.Msg = "failed to post shard to keeper"
			log.WarnErr(fName, *warnErr)
			continue
		}
	}
}
