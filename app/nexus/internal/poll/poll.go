//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package poll

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/auth"
	"net/url"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/internal/log"
)

//func BytesToShares(shards [][]byte) ([]secretsharing.Share, error) {
//	if len(shards) == 0 {
//		return nil, fmt.Errorf("no shards provided")
//	}
//
//	shares := make([]secretsharing.Share, len(shards))
//	g := group.P256
//
//	for i, shardBytes := range shards {
//		// Decode base64
//		//decodedBytes, err := base64.StdEncoding.DecodeString(string(shardBytes))
//		decodedBytes := shardBytes
//		// if err != nil {
//		//	return nil, fmt.Errorf("failed to decode base64 for shard %d: %v", i, err)
//		//}
//
//		// Create ID (starting from 1)
//		id := g.NewScalar()
//		id.SetUint64(uint64(i + 1))
//
//		// Create Value and unmarshal directly from decoded bytes
//		value := g.NewScalar()
//		if err := value.UnmarshalBinary(decodedBytes); err != nil {
//			return nil, fmt.Errorf("failed to unmarshal shard %d: %v", i, err)
//		}
//
//		shares[i] = secretsharing.Share{
//			ID:    id,
//			Value: value,
//		}
//	}
//
//	return shares, nil
//}

// Tick continuously updates SPIKE Keeper, sending the root key to be backed up
// in memory.
//
// It runs until the provided context is cancelled.
//
// The function uses a select statement to either:
// 1. Update the cache when the ticker signals, or
// 2. Exit when the context is done
//
// Parameters:
//   - ctx: A context.Context for cancellation control
//   - source: A pointer to workloadapi.X509Source that provides the source data
//   - ticker: A time.Ticker that determines the update interval
//
// The function will log any errors that occur during cache updates but
// continue running.
//
// To stop the function, cancel the provided context.
//
// Example usage:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	source := &workloadapi.X509Source{...}
//	ticker := time.NewTicker(5 * time.Minute)
//	defer ticker.Stop()
//
//	go Tick(ctx, source, ticker)
func Tick(
	ctx context.Context,
	source *workloadapi.X509Source,
	ticker *time.Ticker,
) {
	//// Wait for the root key to be initialized
	//key := ""
	//for {
	//	key = state.RootKey()
	//	if key == "" {
	//		time.Sleep(time.Second)
	//		continue
	//	}
	//	break
	//}

	// Root key is set only once during initialization, and it is never reset
	// to an empty string, so we can safely assume that if we have a root key
	// here, it will not be empty.

	for {
		select {
		case <-ticker.C:
			// TODO:
			// 1. talk to all keeper endpoints.
			// 2. get at least two shards
			// 3. recompute root key
			// 4. keep root key in cache for 15 mins (configurable)
			// 5. if root key not there, try asking keepers until you get the key.

			keepers := env.Keepers()

			shardsNeeded := 2
			var shardsCollected [][]byte

			for _, keeper := range keepers {
				u, _ := url.JoinPath(keeper, "/v1/store/shard")

				fmt.Println("will try to read from: ", u)

				if source == nil {
					panic("source is nil")
				}

				client, err := net.CreateMtlsClientWithPredicate(
					source, auth.IsKeeper,
				)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}

				md, err := json.Marshal(
					reqres.ShardRequest{},
				)
				if err != nil {
					panic(err)
				}

				data, err := net.Post(client, u, md)
				var res reqres.ShardResponse

				if len(data) == 0 {
					fmt.Println("no data")
				}

				err = json.Unmarshal(data, &res)
				if err != nil {
					fmt.Println("failed to unmarshal")
				}

				fmt.Println(">>>>>> Incoming Shard:", res.Shard)

				if len(shardsCollected) < shardsNeeded {
					decodedShard, err := base64.StdEncoding.DecodeString(res.Shard)
					if err != nil {
						fmt.Println("failed to decode shard:", err)
						continue
					}
					shardsCollected = append(shardsCollected, decodedShard)
				}

				if len(shardsCollected) >= shardsNeeded {
					log.Log().Info("tick", "msg", "Collected required shards", "shards_collected", len(shardsCollected))

					g := group.P256

					firstShard := shardsCollected[0]
					firstShare := secretsharing.Share{
						ID:    g.NewScalar(),
						Value: g.NewScalar(),
					}
					firstShare.ID.SetUint64(1)
					err := firstShare.Value.UnmarshalBinary(firstShard)
					if err != nil {
						panic(err)
					}

					secondShard := shardsCollected[1]
					secondShare := secretsharing.Share{
						ID:    g.NewScalar(),
						Value: g.NewScalar(),
					}
					secondShare.ID.SetUint64(2)
					err = secondShare.Value.UnmarshalBinary(secondShard)
					if err != nil {
						panic(err)
					}

					var shares []secretsharing.Share
					shares = append(shares, firstShare)
					shares = append(shares, secondShare)

					reconstructed, err := secretsharing.Recover(1, shares)
					if err != nil {
						panic(err)
					}

					binaryRec, _ := reconstructed.MarshalBinary()

					fmt.Println("reconstructed", reconstructed)
					fmt.Println("reconstructed", binaryRec)

					//// shardsCollected is a slice of slice of bytes.
					//shares, err := BytesToShares(shardsCollected)
					//if err != nil {
					//	log.Log().Error("tick",
					//		"msg", "Failed to convert shards to shares",
					//		"hint", "Make sure SPIKE Keeper is up and running",
					//		"err", err.Error())
					//}
					//
					//reconstructed, err := secretsharing.Recover(1, shares)
					//
					//if err != nil {
					//	log.Log().Error("tick",
					//		"msg", "Failed to reconstruct root key",
					//		"hint", "Make sure SPIKE Keeper is up and running",
					//	)
					//}

					fmt.Println("reconstructed:>>>>>>", reconstructed)

				}

			}

			fmt.Println("---------------------------------------")

			//err := cache.UpdateCache(source, key)
			//if err != nil {
			//	log.Log().Error("tick",
			//		"msg", "Failed to update the cache",
			//		"hint", "Make sure SPIKE Keeper is up and running",
			//		"err", err.Error())
			//
			//	continue
			//}

			log.Log().Info("tick", "msg", "Successfully updated the cache")
		case <-ctx.Done():
			return
		}
	}
}
