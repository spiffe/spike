//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"fmt"
	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/log"
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

	t := uint(env.ShamirThreshold() - 1) // Need t+1 shares to reconstruct

	reconstructed, err := shamir.Recover(t, shares[:env.ShamirThreshold()])
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

		log.FatalLn(fName + ": Failed to recover: " + err.Error())
	}
	if !secret.IsEqual(reconstructed) {
		// deferred will not run in a fatal crash.
		reconstructed.SetUint64(0)

		log.FatalLn(fName + ": Recovered secret does not match original")
	}
}

// computeShares generates a set of Shamir secret shares from the root key.
// The function uses a deterministic random reader seeded with the root key,
// which ensures that the same shares are always generated for a given root key.
// This deterministic behavior is crucial for the system's reliability, allowing
// shares to be recomputed as needed while maintaining consistency.
func computeShares() (group.Scalar, []shamir.Share) {
	const fName = "computeShares"

	log.Log().Info(fName, "msg", "Computing Shamir shares")

	state.LockRootKey()
	defer state.UnlockRootKey()
	rk := state.RootKeyNoLock()

	// Initialize parameters
	g := group.P256
	t := uint(env.ShamirThreshold() - 1) // Need t+1 shares to reconstruct
	n := uint(env.ShamirShares())        // Total number of shares

	log.Log().Info(fName, "t", t, "n", n)

	// Create secret from our 32 byte key
	secret := g.NewScalar()

	if err := secret.UnmarshalBinary(rk[:]); err != nil {
		log.FatalLn(fName + ": Failed to unmarshal key: %v" + err.Error())
	}

	// To compute identical shares, we need an identical seed for the random
	// reader. Using `finalKey` for seed is secure because Shamir Secret Sharing
	// algorithm's security does not depend on the random seed; it depends on
	// the shards being securely kept secret.
	// If we use `random.Read` instead, then synchronizing shards after Nexus
	// crashes will be cumbersome and prone to edge-case failures.
	reader := crypto.NewDeterministicReader(rk[:])
	ss := shamir.New(reader, t, secret)

	log.Log().Info(fName, "msg", "Generated Shamir shares")

	fmt.Printf("rk: %x\n", *rk)
	fmt.Printf("Secret: %v\n", secret)

	computedShares := ss.Share(n)

	for i, share := range computedShares {
		fmt.Printf("Share %d: %v\n", i, share)
		fmt.Printf("ID: %v\n", share.ID)
		fmt.Printf("Value: %v\n", share.Value)
	}

	// secret is a pointer type; ss.Share(n) is a slice
	// shares will have monotonically increasing IDs, starting from 1.
	return secret, computedShares
}
