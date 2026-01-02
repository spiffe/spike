//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"crypto/rand"
	"fmt"
	"strconv"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// RootShares generates a set of Shamir secret shares from a cryptographically
// secure random root key. It creates a 32-byte random seed, uses it to generate
// a root secret on the P256 elliptic curve group, and splits it into n shares
// using Shamir's Secret Sharing scheme with threshold t. The threshold t is
// set to (ShamirThreshold - 1), meaning t+1 shares are required for
// reconstruction. A deterministic reader seeded with the root key is used to
// ensure identical share generation across restarts, which is critical for
// synchronization after crashes. The function verifies that the generated
// shares can reconstruct the original secret before returning.
//
// Security behavior:
// The application will crash (via log.FatalErr) if:
//   - Called more than once per process (would generate different root keys)
//   - Random number generation fails
//   - Root secret unmarshaling fails
//   - Share reconstruction verification fails
//
// Returns:
//   - []shamir.Share: The generated Shamir secret shares
func RootShares() []shamir.Share {
	const fName = "rootShares"

	// Ensure this function is only called once per process.
	rootSharesGeneratedMu.Lock()
	if rootSharesGenerated {
		failErr := sdkErrors.ErrStateIntegrityCheck.Clone()
		failErr.Msg = "RootShares() called more than once"
		log.FatalErr(fName, *failErr)
	}
	rootSharesGenerated = true
	rootSharesGeneratedMu.Unlock()

	rootKeySeedMu.Lock()
	defer rootKeySeedMu.Unlock()

	if _, err := rand.Read(rootKeySeed[:]); err != nil {
		failErr := sdkErrors.ErrCryptoRandomGenerationFailed.Wrap(err)
		log.FatalErr(fName, *failErr)
	}

	// Initialize parameters
	g := group.P256
	t := uint(env.ShamirThresholdVal() - 1) // Need t+1 shares to reconstruct
	n := uint(env.ShamirSharesVal())        // Total number of shares

	// Create a secret from our 32-byte key:
	rootSecret := g.NewScalar()
	if err := rootSecret.UnmarshalBinary(rootKeySeed[:]); err != nil {
		failErr := sdkErrors.ErrDataUnmarshalFailure.Wrap(err)
		log.FatalErr(fName, *failErr)
	}

	// To compute identical shares, we need an identical seed for the random
	// reader. Using `finalKey` for seed is secure because Shamir Secret Sharing
	// algorithm's security does not depend on the random seed; it depends on
	// the shards being securely kept secret.
	// If we use `random.Read` instead, then synchronizing shards after Nexus
	// crashes will be cumbersome and prone to edge-case failures.
	reader := crypto.NewDeterministicReader(rootKeySeed[:])
	ss := shamir.New(reader, t, rootSecret)

	computedShares := ss.Share(n)

	// Verify the generated shares can reconstruct the original secret.
	// This crashes via log.FatalErr if reconstruction fails.
	crypto.VerifyShamirReconstruction(rootSecret, computedShares)

	return computedShares
}

// RootKey returns a pointer to the root key seed used for encryption.
// This key is generated when RootShares() is called and persists in memory
// for the duration of the bootstrap process. This function acquires a read
// lock to ensure thread-safe access to the root key seed.
//
// Returns:
//   - *[32]byte: Pointer to the root key seed
func RootKey() *[crypto.AES256KeySize]byte {
	rootKeySeedMu.RLock()
	defer rootKeySeedMu.RUnlock()
	return &rootKeySeed
}

// KeeperShare finds and returns the secret share corresponding to a specific
// Keeper ID. It searches through the provided root shares to locate the share
// with an ID matching the given keeperID (converted from string to integer).
// The function uses P256 scalar comparison to match share IDs with the Keeper
// identifier.
//
// Security behavior:
// The application will crash (via log.FatalErr) if:
//   - The keeperID cannot be converted to an integer
//   - No matching share is found for the specified keeper ID
//
// Parameters:
//   - rootShares: The Shamir secret shares to search through
//   - keeperID: The string identifier of the keeper (must be numeric)
//
// Returns:
//   - shamir.Share: The share corresponding to the keeper ID
func KeeperShare(
	rootShares []shamir.Share, keeperID string,
) shamir.Share {
	const fName = "keeperShare"

	var share shamir.Share
	for _, sr := range rootShares {
		kid, err := strconv.Atoi(keeperID)
		if err != nil {
			failErr := sdkErrors.ErrShamirInvalidIndex.Wrap(err)
			failErr.Msg = fmt.Sprintf(
				"failed to convert keeper ID to int: '%s'", keeperID,
			)
			log.FatalErr(fName, *failErr)
		}

		if sr.ID.IsEqual(group.P256.NewScalar().SetUint64(uint64(kid))) {
			share = sr
			break
		}
	}

	if share.ID.IsZero() {
		failErr := sdkErrors.ErrShamirInvalidIndex.Clone()
		failErr.Msg = fmt.Sprintf("no share found for keeper ID: '%s'", keeperID)
		log.FatalErr(fName, *failErr)
	}

	return share
}
