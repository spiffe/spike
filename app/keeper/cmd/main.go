//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/net"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike/app/keeper/internal/env"
	"github.com/spiffe/spike/app/keeper/internal/route/handle"
	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/app/keeper/internal/trust"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/pkg/crypto"
	"net/url"
	"os"
	"sort"
	"time"
)

const appName = "SPIKE Keeper"

const stateNotReady = "NOT_READY"
const stateStarted = "STARTED"
const stateReady = "READY"
const stateError = "ERROR"

func readState() string {
	data, err := os.ReadFile(env.StateFileName())
	if os.IsNotExist(err) {
		return stateNotReady
	}
	if err != nil {
		return stateError
	}
	return string(data)
}

// deterministicReader creates a deterministic reader from a seed
type deterministicReader struct {
	data []byte
	pos  int
}

func (r *deterministicReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		// Generate more deterministic data if needed
		hash := sha256.Sum256(r.data)
		r.data = hash[:]
		r.pos = 0
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func newDeterministicReader(seed []byte) *deterministicReader {
	hash := sha256.Sum256(seed)
	return &deterministicReader{
		data: hash[:],
		pos:  0,
	}
}

func waitForShards() {

	for {
		// Check the current shard count
		shardCount := 0
		state.Shards.Range(func(key, value any) bool {
			shardCount++
			return true
		})

		fmt.Printf("Waiting for shards... Current count: %d\n", shardCount)

		if shardCount >= 3 {
			fmt.Println("Required number of shards received")

			state.Shards.Range(func(key, value any) bool {
				shard := value.([]byte)
				fmt.Printf("Key: %s Shard: %s\n", key, shard)
				return true
			})

			// return
		}

		// Break the loop if we have at least 3 shards
		if shardCount >= 3 {
			fmt.Println("Required number of shards received. Proceeding...")

			finalKey := make([]byte, 32)
			state.Shards.Range(func(key, value any) bool {
				shard := value.([]byte)

				fmt.Printf("Key: %s Shard: %s\n", key, shard)

				for i := 0; i < 32; i++ {
					finalKey[i] ^= shard[i]
				}
				return true
			})

			if len(finalKey) != 32 {
				panic("finalKey must be 32 bytes long")
			}

			// Initialize parameters
			g := group.P256
			t := uint(1) // Need t+1 shares to reconstruct
			n := uint(3) // Total number of shares

			// Create secret from your 32 byte key
			secret := g.NewScalar()
			if err := secret.UnmarshalBinary(finalKey); err != nil {
				panic("Failed to convert key to scalar: %v" + err.Error())
			}

			// Create deterministic random source using the key itself as seed
			// You could use any other seed value for consistency
			deterministicRand := newDeterministicReader([]byte("42"))

			fmt.Println("Final key: >>>>>>", finalKey)

			// Create shares
			ss := secretsharing.New(deterministicRand, t, secret)
			shares := ss.Share(n)

			clonedShares := make([]secretsharing.Share, 0)

			// Print shares
			fmt.Println("Created shares:")
			for i, share := range shares {
				shareData, _ := share.Value.MarshalBinary()
				fmt.Printf("Share %d: %x\n", i+1, shareData)
				fmt.Println("Share ID: ", share.ID)

				encodedShareData := base64.StdEncoding.EncodeToString(shareData)
				fmt.Printf("Base64 Encoded Share %d: %s\n", i+1, encodedShareData)

				decodedShareData, err := base64.StdEncoding.DecodeString(encodedShareData)
				if err != nil {
					panic("Failed to decode share data: " + err.Error())
				}
				fmt.Printf("Decoded Share %d: %x\n", i+1, decodedShareData)

				shareId, _ := share.ID.MarshalBinary()

				encodedShareID := base64.StdEncoding.EncodeToString(shareId)
				fmt.Printf("Base64 Encoded Share ID %d: %s\n", i+1, encodedShareID)

				decodedShareID, err := base64.StdEncoding.DecodeString(encodedShareID)
				if err != nil {
					panic("Failed to decode share ID: " + err.Error())
				}
				fmt.Printf("Decoded Share ID %d: %x\n", i+1, decodedShareID)

				clonedShare := secretsharing.Share{
					ID:    g.NewScalar(),
					Value: g.NewScalar(),
				}

				err = clonedShare.ID.UnmarshalBinary(decodedShareID)
				if err != nil {
					panic("Failed to unmarshal share ID: " + err.Error())
				}

				err = clonedShare.Value.UnmarshalBinary(decodedShareData)
				if err != nil {
					panic("Failed to unmarshal share: " + err.Error())
				}

				clonedShares = append(clonedShares, clonedShare)

				//e := share.Value.UnmarshalBinary(shareData)
				//if e != nil {
				//	panic("Failed to unmarshal share: " + e.Error())
				//}
			}

			// Sort the keys of env.Peers() alphabetically
			peers := env.Peers()
			peerKeys := make([]string, 0, len(peers))
			for id := range peers {
				peerKeys = append(peerKeys, id)
			}
			sort.Strings(peerKeys)

			myId := env.KeeperId()

			// Find the index of the current Keeper's ID
			var myShard []byte
			for index, id := range peerKeys {
				if id == myId {
					// Save the shard corresponding to this Keeper
					if val, ok := state.Shards.Load(myId); ok {
						myShard = val.([]byte)
						fmt.Printf("Saved shard for Keeper ID %s at index %d\n", myId, index)

						shareVal, _ := shares[index].Value.MarshalBinary()

						// TODO: maybe it should be share data instead of shard.
						state.SetShard(shareVal)
						// state.EraseShards()

						break
					}
				}
			}

			// Ensure myShard is stored correctly in the state namespace
			if myShard == nil {
				panic(fmt.Sprintf("Shard for Keeper ID %s could not be found", myId))
			}

			//firstShard, _ := state.Shards.Load("1")
			//secondShard, _ := state.Shards.Load("2")
			//
			//firstId := g.NewScalar()
			//firstId.SetUint64(1)
			//firstValue := g.NewScalar()
			//e := firstValue.UnmarshalBinary(firstShard.([]byte))
			//if e != nil {
			//	panic("Failed to unmarshal share: " + e.Error())
			//}
			//firstShare := secretsharing.Share{
			//	ID:    firstId,
			//	Value: firstValue,
			//}
			//
			//secondId := g.NewScalar()
			//secondId.SetUint64(2)
			//secondValue := g.NewScalar()
			//_ = secondValue.UnmarshalBinary(secondShard.([]byte))
			//secondShare := secretsharing.Share{
			//	ID:    secondId,
			//	Value: secondValue,
			//}
			//
			//newShares := make([]secretsharing.Share, 0)
			//newShares = append(newShares, firstShare)
			//newShares = append(newShares, secondShare)

			fmt.Println("using shares:")

			// Reconstruct using 2 shares
			reconstructed, err := secretsharing.Recover(t, shares[:2])
			if err != nil {
				panic("Failed to reconstruct: " + err.Error())
			}
			fmt.Printf("\nReconstruction successful: %v\n", secret.IsEqual(reconstructed))
			fmt.Println("Original key:    :", finalKey)
			binaryKey, _ := reconstructed.MarshalBinary()
			fmt.Println("Reconstructed key:", binaryKey)

			fmt.Println("using cloned shares:")

			// Reconstruct using 2 cloned shares
			reconstructed, err = secretsharing.Recover(t, clonedShares[:2])
			if err != nil {
				panic("Failed to reconstruct using cloned shares: " + err.Error())
			}
			fmt.Printf("\nReconstruction successful with cloned shares: %v\n", secret.IsEqual(reconstructed))
			fmt.Println("Original key:	:", finalKey)
			binaryKey, _ = reconstructed.MarshalBinary()
			fmt.Println("Reconstructed key using cloned shares:", binaryKey)

			//
			//fmt.Println("using other shares:")
			//
			//// Reconstruct using 2 shares
			//reconstructed, err = secretsharing.Recover(t, newShares)
			//if err != nil {
			//	panic("Failed to reconstruct: " + err.Error())
			//}
			//
			//fmt.Printf("\nReconstruction successful: %v\n", secret.IsEqual(reconstructed))
			//fmt.Println("Original key:    :", finalKey)
			//binaryKey, _ := reconstructed.MarshalBinary()
			//fmt.Println("Reconstructed key:", binaryKey)

			break
		}

		// TODO: if shard count is > 3 then there is a problem.

		// Sleep for a bit before checking again
		time.Sleep(2 * time.Second)
	}
}

var myContribution []byte

func randomContribution() []byte {
	// TODO: check if I need locking.

	if len(myContribution) == 0 {
		mySeed, _ := crypto.Aes256Seed()
		myContribution = []byte(mySeed)

		return myContribution
	}

	return myContribution
}

func contribute(source *workloadapi.X509Source) {
	fmt.Println("I will contribute")

	peers := env.Peers()

	myId := env.KeeperId()

	for id, peer := range peers {
		if id == myId {
			continue
		}

		contributeUrl, _ := url.JoinPath(peer, "v1/store/contribute")

		fmt.Printf("Processing peer %s: %s\n", id, contributeUrl)
		// Add your logic here to interact with each peer

		if source == nil {
			panic("No source")
		}

		client, err := net.CreateMtlsClientWithPredicate(
			source,
			auth.IsKeeper,
		)
		if err != nil {
			panic(err)
		}

		// TODO: maybe an override that returns a byte array instead.
		contribution := randomContribution()

		//if _, err := rand.Read([]byte(contribution)); err != nil {
		//	panic(err)
		//}

		state.Shards.Store(myId, []byte(contribution))

		fmt.Println("---------------------")
		fmt.Println("Sending contribution")
		fmt.Println("contribution:", contribution)
		fmt.Println("contribution:", base64.StdEncoding.EncodeToString([]byte(contribution)))
		fmt.Println("keeper id", myId)
		fmt.Println("peer id", id)
		fmt.Println("---------------------")

		md, err := json.Marshal(
			reqres.ShardContributionRequest{
				KeeperId: myId,
				Shard:    base64.StdEncoding.EncodeToString([]byte(contribution)),
				Version:  0,
			},
		)

		// TODO: this is temporary; we need a more robust handling.
		_, err = net.Post(client, contributeUrl, md)
		for err != nil {
			time.Sleep(5 * time.Second)
			_, err = net.Post(client, contributeUrl, md)
			if err != nil {
				fmt.Println("Retrying in 5 seconds due to error:", err)
				time.Sleep(5 * time.Second)
			}
		}

		// TODO: if error, retry intelligently;
		// TODO: if for exits with not all shards gathered, restart

		fmt.Println("Sent >>>>")
		fmt.Println("err:", err)
	}

}

func main() {
	log.Log().Info(appName, "msg", appName, "version", config.KeeperVersion)

	fmt.Println("IN KEEPER >>>>>>")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, spiffeid, err := spiffe.Source(ctx, spiffe.EndpointSocket())
	if err != nil {
		log.FatalLn(err.Error())
	}
	defer spiffe.CloseSource(source)

	trust.Authenticate(spiffeid)

	// 1. Load State
	keeperState := readState()
	if keeperState == stateError {
		panic("Error reading state file")
	}
	if keeperState == stateNotReady {
		fmt.Println("Not ready. Will send shards")
		// 3. goroutine: Create shard and send to peers.
		go contribute(source)
		go waitForShards()
	}
	if keeperState == stateStarted {
		// 2. If state STARTED but no shard then crashed; try recovery.
		panic("I started, but I don't know what to do.")
	}

	// 4. collect shards
	// 5. if all shards are collected create root key compute your shard
	//    transition to started.

	log.Log().Info(appName,
		"msg", fmt.Sprintf("Started service: %s v%s",
			appName, config.KeeperVersion))
	if err := net.ServeWithPredicate(
		source, handle.InitializeRoutes,
		auth.CanTalkToKeeper,
		env.TlsPort(),
	); err != nil {
		log.FatalF("%s: Failed to serve: %s\n", appName, err.Error())
	}
}
