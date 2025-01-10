//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package poll

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	fmt.Println("in compute shares")
	fmt.Println("final key is: ", finalKey, "len: ", len(finalKey))

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
	// TODO: use deterministic random that uses the root key as the seed
	// otherwise every keeper crash will create a different shard and the
	// system will not be able to recover the root key.
	// or it will have to re-seed all SPIKE Keeper instances (which will be
	// extra work)
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
	// TODO: immediate fix list based on the mini demo:
	//
	// 2. The counter does not work as expected. For example, if there is a
	//    a single keeper live, it will increment until 3 (instead of
	//    remainig at 1.
	//
	// 3. spike policy list gives `null` for no policies instead of a message
	//    also the response is json rather than a more human readable output.
	//    also `createdBy` is emppy.
	//    we can create "good first issue"s for these.

	// TODO: Plan for inverting keeper flow
	// 1. disable all keeper shard creation
	// 2. move shard creation here
	// 3. send shards to keepers (as if Nexus is a keeper; need to adjust some predicates)
	// 4. when keeper crashes, it can ask its only shard from nexus
	// 5. when nexus crashes, it can call keepers.
	// 6. when both crash it's the doomsday break-the-glass scenario, that will be implemented later.

	// Talk to all SPIKE Keeper endpoints and send their shards and get
	// acknowledgement that they received the shard.

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

	decodedRootKey, err := hex.DecodeString(rootKey)
	if err != nil {
		log.FatalLn("Tick: failed to decode root key: " + err.Error())
	}

	rootSecret, rootShares := computeShares(decodedRootKey)
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

	fmt.Println("initialized the backing store with the root key")

	successfulKeepers := make(map[string]bool)

	for {
		select {
		case <-ticker.C:
			keepers := env.Keepers()
			// TODO: this check will change once we make #keepers configurable.
			if len(keepers) < 3 {
				log.FatalLn("Tick: not enough keepers")
			}

			// TODO: exit loop if all keepers initialized.

			// Ensure to get a success response from ALL keepers in one go.
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
				if err != nil {
					log.Log().Warn("tick", "msg", "Failed to marshal request",
						"err", err, "keeper_id", keeperId)
					continue
				}

				data, err := net.Post(client, u, md)
				if err != nil {
					log.Log().Warn("tick", "msg", "Failed to post",
						"err", err, "keeper_id", keeperId)
				}
				var res reqres.ShardContributionResponse

				if len(data) == 0 {
					log.Log().Info("tick", "msg", "No data")
					continue
				}

				fmt.Println("data is: ", string(data))

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

// TODO: disallow policy creation with the same name.
