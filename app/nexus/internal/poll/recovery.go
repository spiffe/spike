package poll

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
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
	"net/url"
	"os"
	"path"
	"time"
)

func recoverUsingKeeperShards(source *workloadapi.X509Source) {
	successfulKeeperShards := make(map[string]string)

	// Iterate through keepers until you get two shards.
	//
	// Any 400 and 5xx response that a SPIKE Keeper gives is likely
	// temporary. We should keep trying until we get a 200 or 404
	// response.
	for {
		for keeperId, keeperApiRoot := range env.Keepers() {
			log.Log().Info("keeper",
				"id", keeperId,
				"url", keeperApiRoot,
			)

			u, err := url.JoinPath(
				keeperApiRoot,
				string(net.SpikeKeeperUrlShard),
			)

			if err != nil {
				log.Log().Warn(
					"tick",
					"msg", "Failed to join path",
					"url", keeperApiRoot,
				)
				continue
			}

			shardRequest := reqres.ShardRequest{}
			md, err := json.Marshal(shardRequest)
			if err != nil {
				log.Log().Warn("tick",
					"msg", "Failed to marshal request",
					"err", err, "keeper_id", keeperId)
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

			var res reqres.ShardResponse
			err = json.Unmarshal(data, &res)
			if err != nil {
				log.Log().Info("tick", "msg",
					"Failed to unmarshal response", "err", err)
				continue
			}

			shard := res.Shard

			if len(shard) == 0 {
				log.Log().Info("tick", "msg", "No shard")
				continue
			}

			successfulKeeperShards[keeperId] = shard
			shrds := make([][]byte, 0)

			if len(successfulKeeperShards) == 2 {
				// TODO: combine shards to create a root key.

				fmt.Println("---------------------------------------------")
				// Print the contents of each shard
				for keeperId, shard := range successfulKeeperShards {
					log.Log().Info("shard",
						"keeper_id", keeperId,
						"contents", shard,
					)

					decodedShard, err := base64.StdEncoding.DecodeString(shard)
					if err != nil {
						log.Log().Warn("tick",
							"msg", "Failed to decode shard from base64",
							"err", err, "keeper_id", keeperId)
						continue
					}
					shrds = append(shrds, decodedShard)

				}
				fmt.Println("---------------------------------------------")

				decodedShard, err := base64.StdEncoding.DecodeString(shard)
				if err != nil {
					log.Log().Warn("tick",
						"msg", "Failed to decode shard from base64",
						"err", err, "keeper_id", keeperId)
					continue
				}

				fmt.Println("decodedShard", decodedShard)

				fmt.Println("my job is done here.")
				g := group.P256
				firstShard := shrds[0]
				secondShard := shrds[1]
				firstShare := secretsharing.Share{
					ID:    g.NewScalar(),
					Value: g.NewScalar(),
				}
				firstShare.ID.SetUint64(1)
				err = firstShare.Value.UnmarshalBinary(firstShard)
				if err != nil {
					log.FatalLn("Failed to unmarshal share: " + err.Error())
				}
				secondShare := secretsharing.Share{
					ID:    g.NewScalar(),
					Value: g.NewScalar(),
				}
				secondShare.ID.SetUint64(2)
				err = secondShare.Value.UnmarshalBinary(secondShard)
				if err != nil {
					log.FatalLn("Failed to unmarshal share: " + err.Error())
				}

				var shares []secretsharing.Share
				shares = append(shares, firstShare)
				shares = append(shares, secondShare)

				reconstructed, err := secretsharing.Recover(1, shares)
				if err != nil {
					log.FatalLn("Failed to recover: " + err.Error())
				}

				// TODO: check for errors.
				binaryRec, _ := reconstructed.MarshalBinary()

				// TODO: check size 32bytes.

				encoded := hex.EncodeToString(binaryRec)
				state.Initialize(encoded)

				// TODO: verify that the root key is identical.
				fmt.Println("initialized state with root key ", encoded)

				return
			}
		}

		log.Log().Info("tick", "msg", "Waiting for keepers to respond")
		time.Sleep(5 * time.Second)
	}
}

func bootstrapBackingStore(source *workloadapi.X509Source) {
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

	fmt.Println("here is the root key:", rootKey)

	decodedRootKey, err := hex.DecodeString(rootKey)
	if err != nil {
		log.FatalLn("Tick: failed to decode root key: " + err.Error())
	}
	rootSecret, rootShares := computeShares(decodedRootKey)
	sanityCheck(rootSecret, rootShares)

	// Initialize the backend store before sending shards to the keepers.
	// SPIKE Keepers are our backup system, and they are not critical for system
	// operations. Initializing early allows SPIKE Nexus to serve before
	// keepers are hydrated.
	state.Initialize(rootKey)
	log.Log().Info("tick", "msg", "Initialized the backing store")

	successfulKeepers := make(map[string]bool)

	for {
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

				tombstone := path.Join(config.SpikeNexusDataFolder(), "bootstrap.tombstone")

				// Create the tombstone file to mark SPIKE Nexus as bootstrapped.
				err = os.WriteFile(tombstone, []byte("spike.nexus.bootstrapped=true"), 0644)
				if err != nil {
					log.FatalLn("Tick: failed to create tombstone file: " + err.Error())
				}

				log.Log().Info("tick", "msg", "Tombstone file created successfully")

				return
			}
		}

		log.Log().Info("tick", "msg", "Waiting for keepers to initialize")
		time.Sleep(5 * time.Second)
	}
}
