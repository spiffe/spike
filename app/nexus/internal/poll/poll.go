//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package poll

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	network "github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/pkg/crypto"
)

func Tick(
	ctx context.Context,
	source *workloadapi.X509Source,
	ticker *time.Ticker,
) {
	// Talk to all SPIKE Keeper endpoints and send their shards and get
	// acknowledgement that they received the shard.

	if source == nil {
		// If source is nil, nobody is going to recreate the source,
		// it's better to log and crash.
		log.FatalLn("Tick: source is nil. this should not happen.")
	}

	// TODO: keep a flag under ~/.spike/.nexus.bootstrap.tombstone
	// If the file exists; then it means that Nexus has successfully
	// bootstrapped.
	// (if the tombstone file is not there, then it means that Nexus has not
	// initialized the backing store.

	// If bootstrapped successfully and there is no backend;
	// then Nexus has not initialized yet. Try getting starting
	// material from the keepers.
	// If you fail to get that info; transition to error
	// state and wait for manual human intervention.

	// If not bootstrapped, then compute shards and bootstrap.

	//be := persist.Backend()
	//if be == nil {
	//	log.Log().Info("tick", "msg", "Backend not found")
	//	// TODO: check shards from keepers. If all keepers respond
	//	// with no shards, then keepers have not been initialized.
	//	//
	//}

	recoveryInfo := persist.ReadRecoveryInfo()
	if recoveryInfo != nil {
		// If SPIKE Nexus
		log.Log().Info("tick", "msg", "Recovery info found")
		return
	}

	// Create the root key and create shards out of the root key.
	rootKey, err := crypto.Aes256Seed()
	if err != nil {
		log.FatalLn("Tick: failed to create root key: " + err.Error())
	}
	decodedRootKey, err := hex.DecodeString(rootKey)
	if err != nil {
		log.FatalLn("Tick: failed to decode root key: " + err.Error())
	}
	rootSecret, rootShares := computeShares(decodedRootKey)
	sanityCheck(rootSecret, rootShares)

	// Initialize the backend store before sending shards to the keepers.
	// Keepers is our backup system, and they are not critical for system
	// operations. Initializing early allows SPIKE Nexus to serve before
	// keepers are hydrated.
	state.Initialize(rootKey)
	log.Log().Info("tick", "msg", "Initialized the backing store")

	successfulKeepers := make(map[string]bool)

	for {
		select {
		case <-ticker.C:
			keepers := env.Keepers()
			if len(keepers) < 3 {
				log.FatalLn("Tick: not enough keepers")
			}

			// Ensure to get a success response from ALL keepers eventually.
			for keeperId, keeperApiRoot := range keepers {
				u, err := url.JoinPath(
					keeperApiRoot,
					string(net.SpikeKeeperUrlContribute),
				)

				if err != nil {
					log.Log().Warn(
						"tick",
						"msg", "Failed to join path",
						"url", keeperApiRoot,
					)
					continue
				}

				client, err := network.CreateMtlsClientWithPredicate(
					source, auth.IsKeeper,
				)

				if err != nil {
					log.Log().Warn("tick",
						"msg", "Failed to create mTLS client",
						"err", err)
					continue
				}

				share := findShare(keeperId, keepers, rootShares)

				contribution, err := share.Value.MarshalBinary()
				if err != nil {
					log.Log().Warn("tick",
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
					log.Log().Warn("tick",
						"msg", "Failed to marshal request",
						"err", err, "keeper_id", keeperId)
					continue
				}

				data, err := net.Post(client, u, md)
				if err != nil {
					log.Log().Warn("tick", "msg",
						"Failed to post",
						"err", err, "keeper_id", keeperId)
				}

				if len(data) == 0 {
					log.Log().Info("tick", "msg", "No data")
					continue
				}

				var res reqres.ShardContributionResponse
				err = json.Unmarshal(data, &res)
				if err != nil {
					log.Log().Info("tick", "msg",
						"Failed to unmarshal response", "err", err)
					continue
				}

				successfulKeepers[keeperId] = true
				log.Log().Info("tick", "msg", "Success", "keeper_id", keeperId)

				if len(successfulKeepers) == 3 {
					log.Log().Info("tick", "msg", "All keepers initialized")
					return
				}
			}
		case <-ctx.Done():
			log.Log().Info("tick", "msg", "Context done")
			return
		}

		log.Log().Info("tick", "msg", "Waiting for keepers to initialize")
		time.Sleep(5 * time.Second)
	}
}
