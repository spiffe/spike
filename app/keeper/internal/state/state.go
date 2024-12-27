//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"fmt"
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike/app/keeper/internal/env"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/pkg/crypto"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var rootKey string
var rootKeyMutex sync.RWMutex

var Shards sync.Map

var shard []byte
var shardMutex sync.RWMutex

func SetShard(s []byte) {
	shardMutex.Lock()
	defer shardMutex.Unlock()
	shard = s
}

func Shard() []byte {
	shardMutex.RLock()
	defer shardMutex.RUnlock()
	return shard
}

func EraseIntermediateShards() {
	Shards.Range(func(key, value interface{}) bool {
		Shards.Delete(key)
		return true
	})
}

// RootKey returns the current root key value in a thread-safe manner.
// It uses a read lock to ensure concurrent read access is safe while
// preventing writes during the read operation.
func RootKey() string {
	rootKeyMutex.RLock()
	defer rootKeyMutex.RUnlock()
	return rootKey
}

// SetRootKey updates the root key value in a thread-safe manner.
// It acquires a write lock to ensure exclusive access during the update,
// preventing any concurrent reads or writes to the root key.
//
// Parameters:
//   - key: The new value to set as the root key
func SetRootKey(key string) {
	rootKeyMutex.Lock()
	defer rootKeyMutex.Unlock()
	rootKey = key
}

// TODO: move private methods to a separate file.

// TODO: move shard-related functionality to a separate file.

var myContribution []byte
var myContributionLock sync.Mutex

func RandomContribution() []byte {
	myContributionLock.Lock()
	defer myContributionLock.Unlock()

	if len(myContribution) == 0 {
		mySeed, _ := crypto.Aes256Seed()
		myContribution = []byte(mySeed)

		return myContribution
	}

	return myContribution
}

func setInternalShard(shares []secretsharing.Share) {
	// Sort the keys of env.Peers() alphabetically for deterministic
	// shard indexing.
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
			if val, ok := Shards.Load(myId); ok {
				myShard = val.([]byte)
				fmt.Printf("Saved shard for Keeper ID %s at index %d\n", myId, index)

				shareVal, _ := shares[index].Value.MarshalBinary()

				SetShard(shareVal)
				EraseIntermediateShards()

				break
			}
		}
	}

	// Ensure myShard is stored correctly in the state namespace
	if myShard == nil {
		panic(fmt.Sprintf("Shard for Keeper ID %s could not be found", myId))
	}
}

func computeFinalKey() []byte {
	finalKey := make([]byte, 32)

	counter := 0
	Shards.Range(func(key, value any) bool {
		counter++
		shard := value.([]byte)
		for i := 0; i < 32; i++ {
			finalKey[i] ^= shard[i]
		}
		return true
	})

	if counter != 3 {
		panic("computeFinalKey: Not all shards received")
	}

	if len(finalKey) != 32 {
		panic("computeFinalKey: FinalKey must be 32 bytes long")
	}

	return finalKey
}

func computeShares(finalKey []byte) (group.Scalar, []secretsharing.Share) {
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
	deterministicRand := crypto.NewDeterministicReader(
		[]byte(env.KeeperRandomSeed()),
	)

	// Create shares
	ss := secretsharing.New(deterministicRand, t, secret)
	return secret, ss.Share(n)
}

func sanityCheck(secret group.Scalar, shares []secretsharing.Share) {
	t := uint(1) // Need t+1 shares to reconstruct

	reconstructed, err := secretsharing.Recover(t, shares[:2])
	if err != nil {
		panic("Failed to reconstruct original: " + err.Error())
	}
	if !secret.IsEqual(reconstructed) {
		panic("Sanity Check Failure: Reconstructed secret does not match original")
	}
}

func WaitForShards() {
	for {
		shardCount := 0
		Shards.Range(func(key, value any) bool {
			shardCount++
			return true
		})

		log.Log().Info(
			"waitForShards", "msg", "Current shard count", "count", shardCount,
		)

		if shardCount < 3 {
			time.Sleep(2 * time.Second)
			continue
		}

		if shardCount > 3 {
			panic("waitForShards: Too many shards received")
		}

		finalKey := computeFinalKey()
		secret, shares := computeShares(finalKey)
		setInternalShard(shares)
		sanityCheck(secret, shares)

		break
	}
}

type AppState string

const AppStateNotReady AppState = "NOT_READY"
const AppStateReady AppState = "READY"
const AppStateError AppState = "ERROR"

func ReadAppState() AppState {
	data, err := os.ReadFile(env.StateFileName())
	if os.IsNotExist(err) {
		return AppStateNotReady
	}
	if err != nil {
		return AppStateError
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return AppStateNotReady
	}
	return AppState(strings.TrimSpace(string(data)))
}
