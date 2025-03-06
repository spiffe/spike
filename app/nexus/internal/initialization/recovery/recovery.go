//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/retry"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/log"
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
// shards are collected to reconstruct the backing store. The function blocks\
// until recovery is successful.
//
// The function maintains a map of successfully recovered shards from each
// keeper to avoid duplicate processing. On failure, it retries with an
// exponential backoff with a max retry delay of 5 seconds.
// The retry timeout is loaded from `env.RecoveryOperationTimeout` and
// defaults to 0 (unlimited; no timeout).
//
// Parameters:
//   - source: An X509Source used for authenticating with keeper nodes
func RecoverBackingStoreUsingKeeperShards(source *workloadapi.X509Source) {
	const fName = "RecoverBackingStoreUsingKeeperShards"

	log.Log().Info(fName, "msg", "Recovering backing store using keeper shards")

	successfulKeeperShards := make(map[string]*[32]byte)

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
			retry.WithMaxInterval(60*time.Second),
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
	if len(secrets) == 0 {
		return
	}

	state.ImportSecrets(secrets)

	log.Log().Info(fName, "msg", "HydrateMemoryFromBackingStore: secrets loaded")
}

// RestoreBackingStoreUsingPilotShards restores the backing store using the
// provided Shamir secret sharing shards. It requires at least the threshold
// number of shards (as configured in the environment) to successfully
// recover the root key. Once the root key is recovered, it initializes the
// state and sends the shards to the keepers.
//
// Parameters:
//   - shards: A slice of base64-encoded string shards
//
// The function will:
//   - Validate that enough shards are provided (at least the threshold amount)
//   - Decode the required number of shards from base64 format
//   - Recover the root key using the Shamir secret sharing algorithm
//   - Initialize the state with the recovered key
//   - Send the shards to the configured keepers
//
// It will return early with an error log if:
//   - There are insufficient shards to meet the threshold
//   - Any shard fails to decode properly
//   - The SPIFFE source cannot be created
func RestoreBackingStoreUsingPilotShards(shards []*[32]byte) {
	const fName = "RestoreBackingStoreUsingPilotShards"

	// Ensure we have at least the threshold number of shards
	if len(shards) < env.ShamirThreshold() {
		log.Log().Error(fName, "msg", "Insufficient shards for recovery",
			"provided", len(shards), "required", env.ShamirThreshold())
		return
	}

	// Recover the root key using the threshold number of shards
	binaryRec := RecoverRootKey(shards)
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
// The function sends shards every 13 seconds (configurable in future). It
// requires a minimum of 3 keepers to be configured. If any operation fails for
// a keeper (URL creation, mTLS setup, marshaling, or network request), it logs
// a warning and continues with the next keeper.
//
// Parameters:
//   - source: An X509Source used for creating mTLS connections to keepers
func SendShardsPeriodically(source *workloadapi.X509Source) {
	const fName = "SendShardsPeriodically"

	log.Log().Info(fName, "msg", "Will send shards to keepers")

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Log().Info(fName, "msg", "Sending shards to keepers")

		// if no root key skip.
		if state.RootKeyZero() {
			log.Log().Info(fName, "msg", "No root key; skipping")
			continue
		}

		keepers := env.Keepers()
		if len(keepers) < env.ShamirShares() {
			log.FatalLn(fName + ": not enough keepers")
		}

		sendShardsToKeepers(source, keepers)
	}
}

// PilotRecoveryShards generates a set of recovery shards from the root key
// using Shamir's Secret Sharing scheme. These shards can be used to reconstruct
// the root key in a recovery scenario.
//
// The function first retrieves the root key from the system state. If no root
// key exists, it returns an empty slice. Otherwise, it splits the root key into
// shares using a secret sharing scheme, performs validation checks, and
// converts the shares into base64-encoded strings.
//
// Each shard in the returned slice represents a portion of the secret needed to
// reconstruct the root key. The shares are generated in a way that requires a
// specific threshold of shards to be combined to recover the original secret.
//
// Returns:
//   - []string: A slice of base64-encoded recovery shards. Returns empty slice
//     if the root key is not available or if share generation fails.
//
// Example:
//
//	shards := PilotRecoveryShards()
//	for _, shard := range shards {
//	    // Store each shard securely
//	    storeShard(shard)
//	}
func PilotRecoveryShards() []string {
	const fName = "PilotRecoveryShards"
	log.Log().Info(fName, "msg", "Generating pilot recovery shards")

	if state.RootKeyZero() {
		return []string{}
	}

	rootSecret, rootShares := computeShares()

	sanityCheck(rootSecret, rootShares)

	result := make([]string, 0, len(rootShares))
	for _, share := range rootShares {
		contribution, err := share.Value.MarshalBinary()
		if err != nil {
			continue
		}
		shard := base64.StdEncoding.EncodeToString(contribution)
		result = append(result, shard)
	}
	return result
}

// BootstrapBackingStoreWithNewRootKey initializes the backing store with a new
// root key if it hasn't been bootstrapped already. It generates a new AES-256
// root key, initializes the state with this key, and distributes key shards
// to all configured keepers.
//
// The function requires a minimum of 3 keepers to be configured. It
// continuously attempts to distribute shards to all keepers until successful,
// waiting 5 seconds between retry attempts. The backing store is initialized
// before keeper distribution to allow immediate operation.
//
// Parameters:
//   - source: An X509Source used for authenticating with keeper nodes
//
// The function will fatal if:
//   - Root key creation fails
//   - Fewer than 3 keepers are configured
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

	//// Create the root key and create shards out of the root key.
	//rk, err := crypto.Aes256Seed()
	//if err != nil {
	//	log.FatalLn("Bootstrap: failed to create root key: " + err.Error())
	//}

	// Initialize the backend store before sending shards to the keepers.
	// SPIKE Keepers are our backup system, and they are not critical for system
	// operations. Initializing early allows SPIKE Nexus to serve before
	// keepers are hydrated.
	// Use a static byte array and pass it as pointer to avoid inadvertent
	// copying / memory allocation.
	var seed [32]byte
	_, err := rand.Read(seed[:])
	if err != nil {
		log.Fatal(err.Error())
	}
	state.Initialize(&seed)
	log.Log().Info(fName, "msg", "Initialized the backing store")

	// Compute Shamir shares out of the root key.
	// TODO: pass seed instead!
	rootShares := mustUpdateRecoveryInfo(&seed)

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
		// TODO: make the time configurable.
		time.Sleep(5 * time.Second)
	}
}
