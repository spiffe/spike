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

	cipher "github.com/spiffe/spike-sdk-go/crypto"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// computeShares generates a set of Shamir secret shares from the root key.
// The function uses a deterministic random reader seeded with the root key,
// which ensures that the same shares are always generated for a given root key.
// This deterministic behavior is crucial for the system's reliability, allowing
// shares to be recomputed as needed while maintaining consistency.
//
// Returns:
//   - group.Scalar: The root secret as a P256 scalar (caller must zero after
//     use)
//   - []shamir.Share: The computed shares with monotonically increasing IDs
//     starting from 1 (caller must zero after use)
//
// The function will log a fatal error and exit if:
//   - The root key is nil or zeroed
//   - The root key fails to unmarshal into a scalar
//   - The generated shares fail reconstruction verification
func computeShares() (group.Scalar, []shamir.Share) {
	const fName = "computeShares"

	state.LockRootKey()
	defer state.UnlockRootKey()
	rk := state.RootKeyNoLock()

	if rk == nil || mem.Zeroed32(rk) {
		failErr := sdkErrors.ErrRootKeyEmpty.Clone()
		log.FatalErr(fName, *failErr)
	}

	g := group.P256
	t := uint(env.ShamirThresholdVal() - 1) // Need t+1 shares to reconstruct
	n := uint(env.ShamirSharesVal())        // Total number of shares

	rootSecret := g.NewScalar()
	if err := rootSecret.UnmarshalBinary(rk[:]); err != nil {
		failErr := sdkErrors.ErrDataUnmarshalFailure.Wrap(err)
		log.FatalErr(fName, *failErr)
	}

	// Using the root key as the seed is secure because Shamir Secret Sharing
	// security does not depend on the random seed; it depends on the shards
	// being kept secret. Using a deterministic reader ensures identical shares
	// are generated for the same root key, which simplifies synchronization
	// after Nexus restarts.
	reader := crypto.NewDeterministicReader(rk[:])
	ss := shamir.New(reader, t, rootSecret)
	shares := ss.Share(n)

	// Verify the generated shares can reconstruct the original secret.
	// This crashes via log.FatalErr if reconstruction fails.
	cipher.VerifyShamirReconstruction(rootSecret, shares)

	return rootSecret, shares
}
