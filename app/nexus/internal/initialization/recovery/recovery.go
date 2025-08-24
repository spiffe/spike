//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/retry"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffe"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

var (
	ErrRecoveryRetry = errors.New("recovery failed; retrying")
)

// InitializeBackingStoreFromKeepers iterates through keepers until
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
func InitializeBackingStoreFromKeepers(source *workloadapi.X509Source) {
	const fName = "InitializeBackingStoreFromKeepers"

	log.Log().Info(fName,
		"message", "Recovering backing store using keeper shards")

	successfulKeeperShards := make(map[string]*[crypto.AES256KeySize]byte)
	// Security: Ensure the shards are zeroed out after use.
	defer func() {
		log.Log().Info(fName, "message", "Resetting successfulKeeperShards")
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
		log.Log().Info(fName, "message", "retry:"+time.Now().String())

		initSuccessful := iterateKeepersAndInitializeState(
			source, successfulKeeperShards,
		)
		if initSuccessful {
			log.Log().Info(fName, "message", "Initialization successful.")
			return true, nil
		}

		log.Log().Warn(fName,
			"message", "Initialization unsuccessful. Will retry.",
			"keepersSoFar", len(successfulKeeperShards),
		)
		return false, ErrRecoveryRetry
	},
		retry.WithBackOffOptions(
			retry.WithMaxInterval(env.RecoveryOperationMaxInterval()),
			retry.WithMaxElapsedTime(0), // Retry forever.
		),
	)

	// This should never happen since the above loop retries forever:
	if err != nil {
		log.FatalLn(fName, "message", "Initialization failed", "err", err)
	}
}

// RestoreBackingStoreUsingPilotShards restores the backing store using the
// provided Shamir secret sharing shards. It requires at least the threshold
// number of shards (as configured in the environment) to successfully
// recover the root key. Once the root key is recovered, it initializes the
// state and sends the shards to the keepers.
//
// Parameters:
//   - shards []*[32]byte: A slice of byte array pointers representing the
//     shards
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

	log.Log().Info(fName, "message", "Restoring backing store using pilot shards")

	// Sanity check:
	for shard := range shards {
		value := shards[shard].Value
		id := shards[shard].ID

		// Security: Crash immediately if data is corrupt.
		if value == nil || mem.Zeroed32(value) || id == 0 {
			log.FatalLn(
				fName,
				"message",
				"Bad input: ID or Value of a shard is zero. Exiting recovery",
			)
		}
	}

	log.Log().Info(fName,
		"message", "Recovering backing store using pilot shards",
		"threshold", env.ShamirThreshold(),
		"len", len(shards),
	)

	// Ensure we have at least the threshold number of shards
	if len(shards) < env.ShamirThreshold() {
		log.Log().Error(fName, "message", "Insufficient shards for recovery",
			"provided", len(shards), "required", env.ShamirThreshold())
		return
	}

	log.Log().Info(fName,
		"message", "Recovering backing store using pilot shards")

	// Recover the root key using the threshold number of shards
	rk := RecoverRootKey(shards)

	if rk == nil || mem.Zeroed32(rk) {
		log.FatalLn(fName, "message", "Failed to recover root key")
	}

	// Security: Ensure the root key is zeroed out after use.
	defer func() {
		mem.ClearRawBytes(rk)
	}()

	log.Log().Info(fName, "message", "Initializing state and root key")
	state.Initialize(rk)

	source, _, err := spiffe.Source(
		context.Background(), spiffe.EndpointSocket(),
	)
	if err != nil {
		log.Log().Info(fName, "message", "Failed to create source", "err", err)
		return
	}
	defer spiffe.CloseSource(source)

	// Don't wait for the next cycle in `SendShardsPeriodically`.
	// Send the shards asap.
	sendShardsToKeepers(source, env.Keepers())
}

// SendShardsPeriodically distributes key shards to configured keeper nodes at
// regular intervals. It creates new shards from the current root key and sends
// them to each keeper using mTLS authentication. The function runs indefinitely
// until stopped.
//
// The function sends shards every 5 minutes. It requires a minimum number of
// keepers equal to the configured Shamir shares. If any operation fails for a
// keeper (URL creation, mTLS setup, marshaling, or network request), it logs a
// warning and continues with the next keeper.
//
// Parameters:
//   - source *workloadapi.X509Source: An X509Source used for creating mTLS
//     connections to keepers
func SendShardsPeriodically(source *workloadapi.X509Source) {
	const fName = "SendShardsPeriodically"

	log.Log().Info(fName, "message", "Will send shards to keepers")

	ticker := time.NewTicker(env.RecoveryKeeperUpdateInterval())
	defer ticker.Stop()

	for range ticker.C {
		log.Log().Info(fName, "message", "Sending shards to keepers")

		// if no root key, then skip.
		if state.RootKeyZero() {
			log.Log().Warn(fName, "message", "No root key; skipping")
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
// shards using a secret sharing scheme, performs validation checks, and
// converts the shares into byte arrays.
//
// Each shard in the returned slice represents a portion of the secret needed to
// reconstruct the root key. The shares are generated in a way that requires a
// specific threshold of shards to be combined to recover the original secret.
//
// Returns:
//   - []*[32]byte: A slice of byte array pointers representing the recovery
//     shards. Returns an empty slice if the root key is not available or if
//     shard generation fails.
//
// Example:
//
//	shards := NewPilotRecoveryShards()
//	for _, shard := range shards {
//	    // Store each shard securely
//	    storeShard(shard)
//	}
func NewPilotRecoveryShards() map[int]*[crypto.AES256KeySize]byte {
	const fName = "NewPilotRecoveryShards"
	log.Log().Info(fName, "message", "Generating pilot recovery shards")

	if state.RootKeyZero() {
		log.Log().Warn(fName, "message", "No root key; skipping")
		return nil
	}

	rootSecret, rootShards := computeShares()
	// sanityCheck crashes the app if shards are corrupted.
	sanityCheck(rootSecret, rootShards)
	// Security: Ensure the root key and shards are zeroed out after use.
	defer func() {
		rootSecret.SetUint64(0)
		for i := range rootShards {
			rootShards[i].Value.SetUint64(0)
		}
	}()

	var result = make(map[int]*[crypto.AES256KeySize]byte)

	for _, shard := range rootShards {
		log.Log().Info(fName, "message", "Generating shard", "shard.id", shard.ID)

		contribution, err := shard.Value.MarshalBinary()
		if err != nil {
			log.Log().Error(fName, "message", "Failed to marshal shard")
			return nil
		}

		if len(contribution) != crypto.AES256KeySize {
			log.Log().Error(fName, "message", "Length of shard is unexpected")
			return nil
		}

		bb, err := shard.ID.MarshalBinary()
		if err != nil {
			log.Log().Error(fName, "message", "Failed to unmarshal shard ID")
			return nil
		}

		bigInt := new(big.Int).SetBytes(bb)
		ii := bigInt.Uint64()

		if len(contribution) != crypto.AES256KeySize {
			log.Log().Error(fName, "message", "Length of shard is unexpected")
			return nil
		}

		var rs [crypto.AES256KeySize]byte
		copy(rs[:], contribution)

		log.Log().Info(fName, "message", "Generated shares", "len", len(rs))

		result[int(ii)] = &rs
	}

	log.Log().Info(fName,
		"message", "Successfully generated pilot recovery shards.")
	return result
}
