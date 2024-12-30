//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package poll

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/auth"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/internal/log"
)

func Tick(
	ctx context.Context,
	source *workloadapi.X509Source,
	ticker *time.Ticker,
) {
	// Talk to all keeper endpoints until we get the minimum number of shards
	// to reconstruct the root key. Once the root key is reconstructed,
	// initialize the backing store with the root key and exit the ticker.

	for {
		if source == nil {
			log.Log().Info("tick", "msg", "source is nil")
			time.Sleep(time.Second * 5)
			continue
		}

		select {
		case <-ticker.C:
			keepers := env.Keepers()

			shardsNeeded := 2
			var shardsCollected [][]byte

			for _, keeperApiRoot := range keepers {
				u, _ := url.JoinPath(keeperApiRoot, "/v1/store/shard")

				client, err := net.CreateMtlsClientWithPredicate(
					source, auth.IsKeeper,
				)
				if err != nil {
					log.Log().Info("tick", "msg",
						"Failed to create mTLS client", "err", err)
					continue
				}

				md, err := json.Marshal(reqres.ShardRequest{})
				if err != nil {
					log.Log().Info("tick", "msg",
						"Failed to marshal request", "err", err)
					continue
				}

				data, err := net.Post(client, u, md)
				var res reqres.ShardResponse

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

				if len(shardsCollected) < shardsNeeded {
					decodedShard, err := base64.StdEncoding.DecodeString(res.Shard)
					if err != nil {
						log.Log().Info("tick", "msg", "Failed to decode shard")
						continue
					}

					// Check if the shard already exists in shardsCollected
					shardExists := false
					for _, existingShard := range shardsCollected {
						if bytes.Equal(existingShard, decodedShard) {
							shardExists = true
							break
						}
					}
					if shardExists {
						continue
					}

					shardsCollected = append(shardsCollected, decodedShard)
				}

				if len(shardsCollected) >= shardsNeeded {
					log.Log().Info("tick",
						"msg", "Collected required shards",
						"shards_collected", len(shardsCollected))

					g := group.P256

					firstShard := shardsCollected[0]
					firstShare := secretsharing.Share{
						ID:    g.NewScalar(),
						Value: g.NewScalar(),
					}
					firstShare.ID.SetUint64(1)
					err := firstShare.Value.UnmarshalBinary(firstShard)
					if err != nil {
						log.FatalLn("Failed to unmarshal share: " + err.Error())
					}

					secondShard := shardsCollected[1]
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

					log.Log().Info("tick", "msg", "Initialized backing store")
					return
				}
			}

			log.Log().Info("tick",
				"msg", "Failed to collect shards... will retry",
			)
		case <-ctx.Done():
			return
		}
	}
}
