//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"context"
	"math/big"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/retry"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffe"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
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
//   - source: An X509Source used for SPIFFE-based mTLS authentication with
//     SPIKE Keeper nodes. Can be nil. If source is nil during a retry
//     iteration, the function will log a warning and retry. This graceful
//     handling allows recovery from transient workload API failures where
//     the source may be temporarily unavailable but can be restored in
//     subsequent retry attempts.
func InitializeBackingStoreFromKeepers(source *workloadapi.X509Source) {
	const fName = "InitializeBackingStoreFromKeepers"

	log.Info(
		fName,
		"message", "recovering backing store using keeper shards",
	)

	successfulKeeperShards := make(map[string]*[crypto.AES256KeySize]byte)
	// Security: Ensure the shards are zeroed out after use.
	defer func() {
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

	_, err := retry.Forever(ctx, func() (bool, *sdkErrors.SDKError) {
		log.Info(fName, "message", "retry:"+time.Now().String())

		// Early check: avoid unnecessary function call if source is nil
		if source == nil {
			log.Warn(fName, "message", "X509 source is nil, will retry")
			return false, sdkErrors.ErrRecoveryRetryFailed
		}

		initSuccessful := iterateKeepersAndInitializeState(
			source, successfulKeeperShards,
		)
		if initSuccessful {
			log.Info(fName, "message", "initialization successful")
			return true, nil
		}

		log.Warn(
			fName,
			"message", "initialization unsuccessful: will retry",
			"keepers_so_far", len(successfulKeeperShards),
		)
		return false, sdkErrors.ErrRecoveryRetryFailed
	})

	// This should never happen since the above loop retries forever:
	if err != nil {
		log.FatalLn(
			fName,
			"message", "initialization failed",
			"err", err.Error(),
		)
	}
}

// RestoreBackingStoreFromPilotShards restores the backing store using the
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
func RestoreBackingStoreFromPilotShards(shards []ShamirShard) {
	const fName = "RestoreBackingStoreFromPilotShards"

	log.Info(
		fName,
		"message", "restoring backing store using pilot shards",
	)

	// Sanity check:
	for shard := range shards {
		value := shards[shard].Value
		id := shards[shard].ID

		// Security: Crash immediately if data is corrupt.
		if value == nil || mem.Zeroed32(value) || id == 0 {
			log.FatalLn(
				fName,
				"message",
				"bad input: ID or Value of a shard is zero: exiting recovery",
			)
			return
		}
	}

	log.Info(
		fName,
		"message", "recovering backing store using pilot shards",
		"threshold", env.ShamirThresholdVal(),
		"len", len(shards),
	)

	// Ensure we have at least the threshold number of shards
	if len(shards) < env.ShamirThresholdVal() {
		log.Error(
			fName,
			"message", "insufficient shards for recovery",
			"provided", len(shards),
			"required", env.ShamirThresholdVal(),
		)
		return
	}

	log.Info(
		fName,
		"message", "recovering backing store using pilot shards",
	)

	// Recover the root key using the threshold number of shards
	rk := ComputeRootKeyFromShards(shards)

	if rk == nil || mem.Zeroed32(rk) {
		log.FatalLn(fName, "message", "failed to recover the root key")
	}

	// Security: Ensure the root key is zeroed out after use.
	defer func() {
		mem.ClearRawBytes(rk)
	}()

	log.Info(fName, "message", "initializing state and root key")
	state.Initialize(rk)

	source, _, err := spiffe.Source(
		context.Background(), spiffe.EndpointSocket(),
	)
	if err != nil {
		log.Error(
			fName,
			"message", "failed to create source",
			"err", err.Error(),
		)
		return
	}
	defer spiffe.CloseSource(source)

	// Don't wait for the next cycle in `SendShardsPeriodically`.
	// Send the shards asap.
	sendShardsToKeepers(source, env.KeepersVal())
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
//   - source: An X509Source used for creating SPIFFE-based mTLS connections to
//     keepers. Can be nil. If source is nil during any iteration, the function
//     performs an early check and skips shard distribution for that iteration,
//     logging a warning and waiting for the next scheduled interval. This
//     graceful handling allows recovery from transient workload API failures.
func SendShardsPeriodically(source *workloadapi.X509Source) {
	const fName = "SendShardsPeriodically"

	log.Info(fName, "message", "will send shards to keepers")

	ticker := time.NewTicker(env.RecoveryKeeperUpdateIntervalVal())
	defer ticker.Stop()

	for range ticker.C {
		log.Info(fName, "message", "sending shards to keepers")

		// Early check: skip if source is nil
		if source == nil {
			log.Warn(fName, "message", "X509 source is nil: skipping shard send")
			continue
		}

		// if no root key, then skip.
		if state.RootKeyZero() {
			log.Warn(fName, "message", "no root key: skipping shard send")
			continue
		}

		keepers := env.KeepersVal()
		if len(keepers) < env.ShamirSharesVal() {
			log.FatalLn(fName, "message", "not enough keepers")
		}

		sendShardsToKeepers(source, keepers)
	}
}

// NewPilotRecoveryShards generates a set of recovery shards from the root key
// using Shamir's Secret Sharing scheme. These shards can be used to reconstruct
// the root key in a recovery scenario.
//
// The function first retrieves the root key from the system state. If no root
// key exists, it returns nil. Otherwise, it splits the root key into shards
// using a secret sharing scheme, performs validation checks, and converts the
// shares into byte arrays.
//
// Each shard in the returned map (keyed by shard ID) represents a portion of
// the secret needed to reconstruct the root key. The shares are generated in a
// way that requires a specific threshold of shards to be combined to recover
// the original secret.
//
// Security and Error Handling:
//
// This function employs a fail-fast strategy with log.FatalLn for any errors
// during shard generation. This is intentional and critical for security:
//   - Shard generation failures indicate memory corruption, crypto library bugs,
//     or corrupted internal state
//   - Continuing to operate with corrupted shards could propagate invalid
//     recovery data to operators
//   - An operator storing broken shards would discover they are useless only
//     during an actual recovery scenario
//   - Crashing immediately ensures the system fails securely rather than
//     silently generating invalid recovery material
//
// Returns:
//   - map[int]*[32]byte: A map of shard IDs to byte array pointers representing
//     the recovery shards. Returns nil if the root key is not available.
//
// Example:
//
//	shards := NewPilotRecoveryShards()
//	for id, shard := range shards {
//	    // Store each shard securely
//	    storeShard(id, shard)
//	}
func NewPilotRecoveryShards() map[int]*[crypto.AES256KeySize]byte {
	const fName = "NewPilotRecoveryShards"
	log.Info(fName, "message", "generating pilot recovery shards")

	if state.RootKeyZero() {
		log.Warn(fName, "message", "no root key: skipping generation")
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
		log.Info(fName, "message", "generating shard", "shard_id", shard.ID)

		contribution, err := shard.Value.MarshalBinary()
		if err != nil {
			log.FatalLn(
				fName,
				"message", "failed to marshal shard",
				"err", err.Error(),
			)
			return nil
		}

		if len(contribution) != crypto.AES256KeySize {
			log.FatalLn(fName, "message", "length of shard is unexpected")
			return nil
		}

		bb, err := shard.ID.MarshalBinary()
		if err != nil {
			log.FatalLn(
				fName,
				"message", "failed to marshal shard ID",
				"err", err.Error(),
			)
			return nil
		}

		bigInt := new(big.Int).SetBytes(bb)
		ii := bigInt.Uint64()

		if len(contribution) != crypto.AES256KeySize {
			log.FatalLn(fName, "message", "length of shard is unexpected")
			return nil
		}

		var rs [crypto.AES256KeySize]byte
		copy(rs[:], contribution)

		log.Info(fName, "message", "generated shards", "len", len(rs))

		result[int(ii)] = &rs
	}

	log.Info(fName,
		"message", "successfully generated pilot recovery shards")
	return result
}
