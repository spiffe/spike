//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/retry"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

var (
	ErrRecoveryRetry = errors.New("recovery failed; retrying")
)

// RecoverBackingStoreUsingKeeperShards iterates through keepers until
// you get two shards.
//
// Any 400 and 5xx response that a SPIKE Keeper gives is likely temporary.
// We should keep trying until we get a 200 or 404 response.
//
// This function attempts to recover the backing store by collecting shards
// from keeper nodes. It continuously polls the keepers until enough valid
// shards are collected to reconstruct the backing store. The function blocks
// until recovery is successful.
//
// The function maintains a map of successfully recovered shards from each
// keeper to avoid duplicate processing. On failure, it retries with an
// exponential backoff with a max retry delay of 5 seconds.
// The retry timeout is loaded from `env.RecoveryOperationTimeout` and
// defaults to 0 (unlimited; no timeout).
//
// Parameters:
//   - source *workloadapi.X509Source: An X509Source used for authenticating
//     with SPIKE Keeper nodes
func RecoverBackingStoreUsingKeeperShards(source *workloadapi.X509Source) {
	const fName = "RecoverBackingStoreUsingKeeperShards"

	log.Log().Info(fName, "msg", "Recovering backing store using keeper shards")

	successfulKeeperShards := make(map[string]*[32]byte)
	// Security: Ensure the shards are zeroed out after use.
	defer func() {
		log.Log().Info(fName, "msg", "Resetting successfulKeeperShards")
		for id := range successfulKeeperShards {
			// Note: you cannot simply use `mem.ClearRawBytes(successfulKeeperShards)`
			// because it will reset the pointer but not the data it points to.
			mem.ClearRawBytes(successfulKeeperShards[id])
		}
	}()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	_, err := retry.Do(ctx, func() (bool, error) {
		log.Log().Info(fName, "msg", "retry:"+time.Now().String())

		recoverySuccessful := iterateKeepersAndTryRecovery(
			source, successfulKeeperShards,
		)
		if recoverySuccessful {
			log.Log().Info(fName, "msg", "Recovery successful")
			return true, nil
		}

		log.Log().Warn(fName, "msg", "Recovery unsuccessful. Will retry.")
		log.Log().Warn(fName, "msg",
			fmt.Sprintf(
				"Successful keepers: %d", len(successfulKeeperShards),
			),
		)
		log.Log().Warn(fName, "msg", "!!! YOU MAY NEED TO MANUALLY BOOSTRAP !!!")
		log.Log().Info(fName, "msg", "Waiting for keepers to respond")
		return false, ErrRecoveryRetry
	},
		retry.WithBackOffOptions(
			retry.WithMaxInterval(env.RecoveryOperationMaxInterval()),
			retry.WithMaxElapsedTime(env.RecoveryOperationTimeout()),
		),
	)

	if err != nil {
		log.Log().Warn("Recovery failed; timed out")
		log.Log().Warn("You need to manually bootstrap SPIKE Nexus")
	}
}

// HydrateMemoryFromBackingStore loads all secrets from the persistent storage
// into the application's memory state. This function is typically called during
// application startup to restore the secret state from the previous session.
//
// The function reads all secrets from the backing store using
// persist.ReadAllSecrets() and imports them into the application state using
// state.ImportSecrets(). If no secrets are found in the backing store, the
// function returns without making any changes to the application state.
//
// Example usage:
//
//	func initializeApp() {
//		// Other initialization code...
//		memory.HydrateMemoryFromBackingStore()
//		// Continue with application startup
//	}
func HydrateMemoryFromBackingStore() {
	const fName = "HydrateMemoryFromBackingStore"

	log.Log().Info(fName, "msg", "HydrateMemoryFromBackingStore")

	secrets := persist.ReadAllSecrets()
	if len(secrets) > 0 {
		state.ImportSecrets(secrets)
	}
	log.Log().Info(fName, "msg", "HydrateMemoryFromBackingStore: secrets loaded")

	policies := persist.ReadAllPolicies()
	if len(policies) > 0 {
		state.ImportPolicies(policies)
	}

	log.Log().Info(fName, "msg", "HydrateMemoryFromBackingStore: policies loaded")
}

// RestoreBackingStoreUsingPilotShards restores the backing store using the
// provided Shamir secret sharing shards. It requires at least the threshold
// number of shards (as configured in the environment) to successfully
// recover the root key. Once the root key is recovered, it initializes the
// state and sends the shards to the keepers.
//
// Parameters:
//   - shards []*[32]byte: A slice of byte array pointers representing the shards
//
// The function will:
//   - Validate that enough shards are provided (at least the threshold amount)
//   - Recover the root key using the Shamir secret sharing algorithm
//   - Initialize the state with the recovered key
//   - Send the shards to the configured keepers
//
// It will return early with an error log if:
//   - There are not enough shards to meet the threshold
//   - The SPIFFE source cannot be created
func RestoreBackingStoreUsingPilotShards(shards []ShamirShard) {
	const fName = "RestoreBackingStoreUsingPilotShards"

	log.Log().Info(fName, "msg", "Restoring backing store using pilot shards")

	// Sanity check:
	for shard := range shards {
		value := shards[shard].Value
		id := shards[shard].Id

		if mem.Zeroed32(value) || id == 0 {
			log.Log().Error(
				fName,
				"msg", "Bad input: ID or Value of a shard is zero. Exiting recovery",
			)
			return
		}
	}

	log.Log().Info(fName,
		"msg", "Recovering backing store using pilot shards",
		"threshold", env.ShamirThreshold(),
		"len", len(shards),
	)

	// Ensure we have at least the threshold number of shards
	if len(shards) < env.ShamirThreshold() {
		log.Log().Error(fName, "msg", "Insufficient shards for recovery",
			"provided", len(shards), "required", env.ShamirThreshold())
		return
	}

	log.Log().Info(fName, "msg", "Recovering backing store using pilot shards")

	// Recover the root key using the threshold number of shards
	binaryRec := RecoverRootKey(shards)
	// Security: Ensure the root key is zeroed out after use.
	defer func() {
		mem.ClearRawBytes(binaryRec)
	}()

	log.Log().Info(fName, "msg", "Initializing state and root key")
	state.Initialize(binaryRec)
	state.SetRootKey(binaryRec)

	source, _, err := spiffe.Source(
		context.Background(), spiffe.EndpointSocket(),
	)
	if err != nil {
		log.Log().Info(fName, "msg", "Failed to create source", "err", err)
		return
	}
	defer spiffe.CloseSource(source)

	// Don't wait for the next cycle. Send the shards asap.
	sendShardsToKeepers(source, env.Keepers())
}

// SendShardsPeriodically distributes key shards to configured keeper nodes at
// regular intervals. It creates new shards from the current root key and sends
// them to each keeper using mTLS authentication. The function runs indefinitely
// until stopped.
//
// The function sends shards every 5 minutes. It requires a minimum number of keepers
// equal to the configured Shamir shares. If any operation fails for a keeper
// (URL creation, mTLS setup, marshaling, or network request), it logs a warning
// and continues with the next keeper.
//
// Parameters:
//   - source *workloadapi.X509Source: An X509Source used for creating mTLS
//     connections to keepers
func SendShardsPeriodically(source *workloadapi.X509Source) {
	const fName = "SendShardsPeriodically"

	log.Log().Info(fName, "msg", "Will send shards to keepers")

	ticker := time.NewTicker(env.RecoveryKeeperUpdateInterval())
	defer ticker.Stop()

	for range ticker.C {
		log.Log().Info(fName, "msg", "Sending shards to keepers")

		// if no root key, then skip.
		if state.RootKeyZero() {
			log.Log().Warn(fName, "msg", "No root key; skipping")
			continue
		}

		keepers := env.Keepers()
		if len(keepers) < env.ShamirShares() {
			log.FatalLn(fName + ": not enough keepers")
		}

		sendShardsToKeepers(source, keepers)
	}
}

// NewPilotRecoveryShards generates a set of recovery shards from the root key
// using Shamir's Secret Sharing scheme. These shards can be used to reconstruct
// the root key in a recovery scenario.
//
// The function first retrieves the root key from the system state. If no root
// key exists, it returns an empty slice. Otherwise, it splits the root key into
// shares using a secret sharing scheme, performs validation checks, and
// converts the shares into byte arrays.
//
// Each shard in the returned slice represents a portion of the secret needed to
// reconstruct the root key. The shares are generated in a way that requires a
// specific threshold of shards to be combined to recover the original secret.
//
// Returns:
//   - []*[32]byte: A slice of byte array pointers representing the recovery
//     shards. Returns an empty slice if the root key is not available or if
//     share generation fails.
//
// Example:
//
//	shards := NewPilotRecoveryShards()
//	for _, shard := range shards {
//	    // Store each shard securely
//	    storeShard(shard)
//	}
func NewPilotRecoveryShards() map[int]*[32]byte {
	const fName = "NewPilotRecoveryShards"
	log.Log().Info(fName, "msg", "Generating pilot recovery shards")

	if state.RootKeyZero() {
		log.Log().Warn(fName, "msg", "No root key; skipping")
		return nil
	}

	rootSecret, rootShares := computeShares()
	// Security: Ensure the root key and shares are zeroed out after use.
	sanityCheck(rootSecret, rootShares)
	defer func() {
		rootSecret.SetUint64(0)
		for i := range rootShares {
			rootShares[i].Value.SetUint64(0)
		}
	}()

	var result = make(map[int]*[32]byte)

	for _, share := range rootShares {
		log.Log().Info(fName, "msg", "Generating share", "share.id", share.ID)

		contribution, err := share.Value.MarshalBinary()
		if err != nil {
			log.Log().Error(fName, "msg", "Failed to marshal share")
			return nil
		}

		if len(contribution) != 32 {
			log.Log().Error(fName, "msg", "Length of share is unexpected")
			return nil
		}

		bb, err := share.ID.MarshalBinary()
		if err != nil {
			log.Log().Error(fName, "msg", "Failed to unmarshal share Id")
			return nil
		}

		bigInt := new(big.Int).SetBytes(bb)
		ii := bigInt.Uint64()

		if len(contribution) != 32 {
			log.Log().Error(fName, "msg", "Length of share is unexpected")
			return nil
		}

		var rs [32]byte
		copy(rs[:], contribution)

		log.Log().Info(fName, "msg", "Generated shares", "len", len(rs))

		result[int(ii)] = &rs
	}

	log.Log().Info(fName, "msg", "Successfully generated pilot recovery shards.")
	return result
}

// BootstrapBackingStoreWithNewRootKey initializes the backing store with a new
// root key if it hasn't been bootstrapped already. It generates a new AES-256
// root key, initializes the state with this key, and distributes key shards
// to all configured keepers.
//
// The function requires the number of keepers to match the configured Shamir
// shares. It continuously attempts to distribute shards to all keepers until
// successful, waiting 5 seconds between retry attempts. The backing store is
// initialized before keeper distribution to allow immediate operation.
//
// Parameters:
//   - source *workloadapi.X509Source: An X509Source used for authenticating
//     with keeper nodes
//
// The function will crash fatally if:
//   - Root key creation fails
//   - The number of keepers doesn't match the configured Shamir shares
func BootstrapBackingStoreWithNewRootKey(source *workloadapi.X509Source) {
	const fName = "BootstrapBackingStoreWithNewRootKey"

	log.Log().Info(fName, "msg",
		"Tombstone file does not exist. Bootstrapping SPIKE Nexus...")

	if !state.RootKeyZero() {
		log.Log().Info(fName, "msg",
			"Recovery info found. Backing store already bootstrapped.",
		)
		return
	}

	// Initialize the backend store before sending shards to the keepers.
	// SPIKE Keepers are our backup system, and they are not critical for system
	// operations. Initializing early allows SPIKE Nexus to serve before
	// keepers are hydrated.
	//
	// Security: Use a static byte array and pass it as a pointer to avoid
	// inadvertent copying / pass-by-value / memory allocation.
	var seed [32]byte
	// Security: Ensure the seed is zeroed out after use.
	defer func() {
		mem.ClearRawBytes(&seed)
	}()

	if _, err := rand.Read(seed[:]); err != nil {
		log.Fatal(err.Error())
	}

	state.Initialize(&seed)
	log.Log().Info(fName, "msg", "Initialized the backing store")

	// Compute Shamir shares out of the root key.
	rootShares := mustUpdateRecoveryInfo(&seed)
	// Security: Ensure the seed is zeroed out after use.
	defer func() {
		for _, share := range rootShares {
			share.Value.SetUint64(0)
		}
	}()

	successfulKeepers := make(map[string]bool)
	keepers := env.Keepers()

	shamirShareCount := env.ShamirShares()
	if len(keepers) != shamirShareCount {
		log.FatalLn(
			fName+": Keepers not configured correctly.",
			"Share count:", shamirShareCount, "Keepers:", len(keepers),
		)
	}

	for {
		// Ensure to get a success response from ALL keepers eventually.
		exit := iterateKeepersToBootstrap(
			keepers, rootShares, successfulKeepers, source,
		)
		if exit {
			return
		}

		log.Log().Info(fName, "msg", "Waiting for keepers to initialize")
		time.Sleep(env.RecoveryOperationPollInterval())
	}
}
