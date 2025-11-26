//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// sanityCheck verifies that a set of secret shares can correctly reconstruct
// the original secret. It performs this verification by attempting to recover
// the secret using the minimum required number of shares and comparing the
// result with the original secret.
//
// Parameters:
//   - secret group.Scalar: The original secret to verify against
//   - shares []shamir.Share: The generated secret shares to verify
//
// The function will:
//   - Calculate the threshold (t) from the environment configuration
//   - Attempt to reconstruct the secret using exactly t+1 shares
//   - Compare the reconstructed secret with the original
//   - Zero out the reconstructed secret regardless of success or failure
//
// If the verification fails, the function will:
//   - Log a fatal error and exit if recovery fails
//   - Log a fatal error and exit if the recovered secret doesn't match the
//     original
//
// Security:
//   - The reconstructed secret is always zeroed out to prevent memory leaks
//   - In case of fatal errors, the reconstructed secret is explicitly zeroed
//     before logging since deferred functions won't run after log.FatalLn
func sanityCheck(secret group.Scalar, shares []shamir.Share) {
	const fName = "sanityCheck"

	t := uint(env.ShamirThresholdVal() - 1) // Need t+1 shares to reconstruct

	reconstructed, err := shamir.Recover(t, shares[:env.ShamirThresholdVal()])
	// Security: Ensure that the secret is zeroed out if the check fails.
	defer func() {
		if reconstructed == nil {
			return
		}
		reconstructed.SetUint64(0)
	}()

	if err != nil {
		// deferred will not run in a fatal crash.
		reconstructed.SetUint64(0)

		log.FatalLn(
			fName,
			"message", "failed to recover",
			"err", err.Error(),
		)
	}
	if !secret.IsEqual(reconstructed) {
		// deferred will not run in a fatal crash.
		reconstructed.SetUint64(0)

		log.FatalLn(fName, "message", "recovered secret does not match original")
	}
}

// computeShares generates a set of Shamir secret shares from the root key.
// The function uses a deterministic random reader seeded with the root key,
// which ensures that the same shares are always generated for a given root key.
// This deterministic behavior is crucial for the system's reliability, allowing
// shares to be recomputed as needed while maintaining consistency.
func computeShares() (group.Scalar, []shamir.Share) {
	const fName = "computeShares"

	state.LockRootKey()
	defer state.UnlockRootKey()
	rk := state.RootKeyNoLock()

	if rk == nil || mem.Zeroed32(rk) {
		failErr := sdkErrors.ErrRootKeyEmpty.Clone()
		log.FatalErr(fName, *failErr)
	}

	// TODO: this one and bootstrap/internal/state.go:70 do almost identical
	// things.

	// Initialize parameters
	g := group.P256
	t := uint(env.ShamirThresholdVal() - 1) // Need t+1 shares to reconstruct
	n := uint(env.ShamirSharesVal())        // Total number of shares

	// Create a secret from our 32-byte key:
	rootSecret := g.NewScalar()
	if err := rootSecret.UnmarshalBinary(rk[:]); err != nil {
		failErr := sdkErrors.ErrDataUnmarshalFailure.Wrap(err)
		log.FatalErr(fName, *failErr)
	}

	// To compute identical shares, we need an identical seed for the random
	// reader. Using `finalKey` for seed is secure because Shamir Secret Sharing
	// algorithm's security does not depend on the random seed; it depends on
	// the shards being securely kept secret.
	// If we use `random.Read` instead, then synchronizing shards after Nexus
	// crashes will be cumbersome and prone to edge-case failures.
	reader := crypto.NewDeterministicReader(rk[:])
	ss := shamir.New(reader, t, rootSecret)

	computedShares := ss.Share(n)

	// Security: Ensure the root key and shares are zeroed out after use.
	// validation.VerifyShamirReconstruction(rootSecret, computedShares

	// secret is a pointer type; ss.Share(n) is a slice
	// shares will have monotonically increasing IDs, starting from 1.
	return rootSecret, computedShares
}
