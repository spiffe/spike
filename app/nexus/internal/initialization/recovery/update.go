//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/base64"
	"encoding/hex"
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

func mustUpdateRecoveryInfo(rk string) []secretsharing.Share {
	const fName = "mustUpdateRecoveryInfo"
	log.Log().Info(fName, "msg", "Updating recovery info")

	decodedRootKey, err := hex.DecodeString(rk)
	if err != nil {
		log.FatalLn(fName + ": failed to decode root key: " + err.Error())
	}
	rootSecret, rootShares := computeShares(decodedRootKey)
	sanityCheck(rootSecret, rootShares)

	// Save recovery information.
	state.SetRootKey(decodedRootKey)

	return rootShares
}

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

		rk := state.RootKey()
		if rk == nil {
			log.Log().Info(fName, "msg", "rootKey is nil; moving on...")
			continue
		}

		rootSecret, rootShares := computeShares(rk)

		sanityCheck(rootSecret, rootShares)

		share := findShare(keeperId, keepers, rootShares)

		contribution, err := share.Value.MarshalBinary()
		if err != nil {
			log.Log().Warn(fName,
				"msg", "Failed to marshal share",
				"err", err, "keeper_id", keeperId)
			continue
		}

		scr := reqres.ShardContributionRequest{
			KeeperId: keeperId,
			Shard:    base64.StdEncoding.EncodeToString(contribution),
		}
		md, err := json.Marshal(scr)
		if err != nil {
			log.Log().Warn(fName,
				"msg", "Failed to marshal request",
				"err", err, "keeper_id", keeperId)
			continue
		}

		_, err = net.Post(client, u, md)
		if err != nil {
			log.Log().Warn(fName, "msg",
				"Failed to post",
				"err", err, "keeper_id", keeperId)
		}
	}
}
