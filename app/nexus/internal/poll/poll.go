//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package poll

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	network "github.com/spiffe/spike-sdk-go/net"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/config"
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

	tombstone := path.Join(config.SpikeNexusDataFolder(), "bootstrap.tombstone")

	_, err := os.Stat(tombstone)

	if err == nil {
		log.Log().Info("tick",
			"msg", "Tombstone file exists, SPIKE Nexus is bootstrapped",
		)

		panic("Implement me: Recover root key from SPIKE Keepers.")

		return
	}

	if !os.IsNotExist(err) {
		log.FatalLn("Tick: failed to check tombstone file: " + err.Error())
	}

	log.Log().Info("tick", "msg",
		"Tombstone file does not exist. Bootstrapping SPIKE Nexus...")

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

					// Create the tombstone file to mark SPIKE Nexus as bootstrapped.
					err = os.WriteFile(tombstone, []byte("spike.nexus.bootstrapped=true"), 0644)
					if err != nil {
						log.FatalLn("Tick: failed to create tombstone file: " + err.Error())
					}

					log.Log().Info("tick", "msg", "Tombstone file created successfully")

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
