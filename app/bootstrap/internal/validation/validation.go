//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/log"
)

// VerifyShamirReconstruction verifies that a set of secret shares can
// correctly reconstruct the original secret. It performs this verification by
// attempting to recover the secret using the minimum required number of shares
// and comparing the result with the original secret.
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
func VerifyShamirReconstruction(secret group.Scalar, shares []shamir.Share) {
	const fName = "VerifyShamirReconstruction"

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

		log.FatalLn(fName + ": Failed to recover: " + err.Error())
	}
	if !secret.IsEqual(reconstructed) {
		// deferred will not run in a fatal crash.
		reconstructed.SetUint64(0)

		log.FatalLn(fName + ": Recovered secret does not match original")
	}
}
