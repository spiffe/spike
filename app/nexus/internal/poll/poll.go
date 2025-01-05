//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package poll

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	net2 "github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
	"github.com/spiffe/spike/pkg/crypto"
	"net/url"
	"sort"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func sanityCheck(secret group.Scalar, shares []secretsharing.Share) {
	t := uint(1) // Need t+1 shares to reconstruct

	reconstructed, err := secretsharing.Recover(t, shares[:2])
	if err != nil {
		log.FatalLn("computeShares: Failed to recover: " + err.Error())
	}
	if !secret.IsEqual(reconstructed) {
		log.FatalLn("computeShares: Recovered secret does not match original")
	}
}

func computeShares(finalKey []byte) (group.Scalar, []secretsharing.Share) {
	// Initialize parameters
	g := group.P256
	// TODO: these will be configurable
	t := uint(1) // Need t+1 shares to reconstruct
	n := uint(3) // Total number of shares

	// Create secret from your 32 byte key
	secret := g.NewScalar()
	if err := secret.UnmarshalBinary(finalKey); err != nil {
		log.FatalLn("computeShares: Failed to unmarshal key: %v" + err.Error())
	}

	// Create shares
	ss := secretsharing.New(rand.Reader, t, secret)
	return secret, ss.Share(n)
}

func findShare(id string, keepers map[string]string, shares []secretsharing.Share) *secretsharing.Share {

	sortedKeys := make([]string, 0, len(keepers))
	for k := range keepers {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	matchingIndex := -1
	for i, key := range sortedKeys {
		if key == id {
			matchingIndex = i
			break
		}
	}

	if matchingIndex == -1 {
		return nil
	}

	if matchingIndex < 0 || matchingIndex >= len(shares) {
		return nil
	}

	return &shares[matchingIndex]
}

func Tick(
	ctx context.Context,
	source *workloadapi.X509Source,
	ticker *time.Ticker,
) {
	// TODO: Plan for inverting keeper flow
	// 1. disable all keeper shard creation
	// 2. move shard creation here
	// 3. send shards to keepers (as if Nexus is a keeper; need to adjust some predicates)
	// 4. when keeper crashes, it can ask its only shard from nexus
	// 5. when nexus crashes, it can call keepers.
	// 6. when both crash it's the doomsday break-the-glass scenario, that will be implemented later.

	// Talk to all SPIKE Keeper endpoints and send their shards and get
	// acknowledgement that they received the shard.
	for {
		if source == nil {
			// If source is nil, nobody is going to recreate the source,
			// it's better to log and crash.
			log.FatalLn("Tick: source is nil. this should not happen.")
		}

		// Create the root key and create shards out of the root key.
		// TODO: Securely erase intermediate material

		rootKey, err := crypto.Aes256Seed()
		if err != nil {
			log.FatalLn("Tick: failed to create root key: " + err.Error())
		}

		rootSecret, rootShares := computeShares([]byte(rootKey))
		sanityCheck(rootSecret, rootShares)

		// TODO: implement this logic:
		// If already initialized, and no root key, ask keepers and assemble the root key.
		// already initialized -> use a flag in the file system
		// root key check -> if dbBackend is not nil, then it has a root key.
		// The flow below assumes that the keeper is not initialized, and it's day zero.

		// Initialize the backend store before sending shards to the keepers.
		// Keepers is our backup system, and they are not critical for system
		// operations. Initializing early allows SPIKE Nexus to serve before
		// keepers are hydrated.
		state.Initialize(rootKey)

		select {
		case <-ticker.C:
			keepers := env.Keepers()
			// TODO: this check will change once we make #keepers configurable.
			if len(keepers) < 3 {
				log.FatalLn("Tick: not enough keepers")
			}

			for keeperId, keeperApiRoot := range keepers {
				u, err := url.JoinPath(
					keeperApiRoot,
					string(net.SpikeKeeperUrlContribute),
				)

				if err != nil {
					log.Log().Warn(
						"tick", "msg",
						"Failed to join path", keeperApiRoot,
					)
					continue
				}

				client, err := net2.CreateMtlsClientWithPredicate(
					source, auth.IsKeeper,
				)

				if err != nil {
					log.Log().Warn("tick", "msg",
						"Failed to create mTLS client", "err", err)
					continue
				}

				share := findShare(keeperId, keepers, rootShares)

				contribution, err := share.Value.MarshalBinary()
				if err != nil {
					log.Log().Warn("tick", "msg", "Failed to marshal share",
						"err", err, "keeper_id", keeperId)
					continue
				}

				scr := reqres.ShardContributionRequest{
					KeeperId: keeperId,
					Shard:    base64.StdEncoding.EncodeToString(contribution),
				}
				md, err := json.Marshal(scr)

				data, err := net.Post(client, u, md)
				var res reqres.ShardContributionResponse

				if len(data) == 0 {
					log.Log().Info("tick", "msg", "No data")
					continue
				}

				err = json.Unmarshal(data, &res)
				if err != nil {
					log.Log().Info("tick", "msg",
						"Failed to unmarshal response", "err", err)
					continue
				}
			}
		case <-ctx.Done():
			return
		}

		time.Sleep(5 * time.Second)
	}
}

//
//select {
//case <-ticker.C:
//	keepers := env.Keepers()
//
//	shardsNeeded := 2
//	var shardsCollected [][]byte
//
//	for _, keeperApiRoot := range keepers {
//		u, _ := url.JoinPath(keeperApiRoot, "/v1/store/shard")
//
//		client, err := net.CreateMtlsClientWithPredicate(
//			source, auth.IsKeeper,
//		)
//		if err != nil {
//			log.Log().Info("tick", "msg",
//				"Failed to create mTLS client", "err", err)
//			continue
//		}
//
//		md, err := json.Marshal(reqres.ShardRequest{})
//		if err != nil {
//			log.Log().Info("tick", "msg",
//				"Failed to marshal request", "err", err)
//			continue
//		}
//
//		data, err := net.Post(client, u, md)
//		var res reqres.ShardResponse
//
//		if len(data) == 0 {
//			log.Log().Info("tick", "msg", "No data")
//			continue
//		}
//
//		err = json.Unmarshal(data, &res)
//		if err != nil {
//			log.Log().Info("tick", "msg",
//				"Failed to unmarshal response", "err", err)
//			continue
//		}
//
//		if len(shardsCollected) < shardsNeeded {
//			decodedShard, err := base64.StdEncoding.DecodeString(res.Shard)
//			if err != nil {
//				log.Log().Info("tick", "msg", "Failed to decode shard")
//				continue
//			}
//
//			// Check if the shard already exists in shardsCollected
//			shardExists := false
//			for _, existingShard := range shardsCollected {
//				if bytes.Equal(existingShard, decodedShard) {
//					shardExists = true
//					break
//				}
//			}
//			if shardExists {
//				continue
//			}
//
//			shardsCollected = append(shardsCollected, decodedShard)
//		}
//
//		if len(shardsCollected) >= shardsNeeded {
//			log.Log().Info("tick",
//				"msg", "Collected required shards",
//				"shards_collected", len(shardsCollected))
//
//			g := group.P256
//
//			firstShard := shardsCollected[0]
//			firstShare := secretsharing.Share{
//				ID:    g.NewScalar(),
//				Value: g.NewScalar(),
//			}
//			firstShare.ID.SetUint64(1)
//			err := firstShare.Value.UnmarshalBinary(firstShard)
//			if err != nil {
//				log.FatalLn("Failed to unmarshal share: " + err.Error())
//			}
//
//			secondShard := shardsCollected[1] secondShare := secretsharing.Share{
//				ID:    g.NewScalar(),
//				Value: g.NewScalar(),
//			}
//			secondShare.ID.SetUint64(2)
//			err = secondShare.Value.UnmarshalBinary(secondShard)
//			if err != nil {
//				log.FatalLn("Failed to unmarshal share: " + err.Error())
//			}
//
//			var shares []secretsharing.Share
//			shares = append(shares, firstShare)
//			shares = append(shares, secondShare)
//
//			reconstructed, err := secretsharing.Recover(1, shares)
//			if err != nil {
//				log.FatalLn("Failed to recover: " + err.Error())
//			}
//
//			// TODO: check for errors.
//			binaryRec, _ := reconstructed.MarshalBinary()
//
//			// TODO: check size 32bytes.
//
//			encoded := hex.EncodeToString(binaryRec)
//			state.Initialize(encoded)
//
//			log.Log().Info("tick", "msg", "Initialized backing store")
//			return
//		}
//
//		log.Log().Info("tick",
//			"msg", "Failed to collect shards... will retry",
//		)
//	case <-ctx.Done():
//		return
//	}
//}
