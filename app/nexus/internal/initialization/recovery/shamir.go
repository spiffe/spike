//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/log"
)

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

	state.LockRootKey()
	defer state.UnlockRootKey()
	rk := state.RootKeyNoLock()

	// Initialize parameters
	g := group.P256
	t := uint(env.ShamirThreshold() - 1) // Need t+1 shares to reconstruct
	n := uint(env.ShamirShares())        // Total number of shares

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

	// secret is a pointer type; ss.Share(n) is a slice
	// shares will have monotonically increasing IDs, starting from 1.
	return secret, ss.Share(n)
}
