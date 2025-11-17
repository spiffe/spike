//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"crypto/rand"
	"strconv"
	"sync"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/bootstrap/internal/validation"
)

var (
	// rootKeySeed stores the root key seed generated during initialization.
	// It is kept in memory to allow encryption operations during bootstrap.
	rootKeySeed [crypto.AES256KeySize]byte
	// rootKeySeedMu provides mutual exclusion for access to the root key seed.
	rootKeySeedMu sync.RWMutex

	// rootSharesGenerated tracks whether RootShares() has been called.
	rootSharesGenerated bool
	// rootSharesGeneratedMu protects the rootSharesGenerated flag.
	rootSharesGeneratedMu sync.Mutex
)

// RootShares generates a set of Shamir secret shares from a cryptographically
// secure random root key. It creates a 32-byte random seed, uses it to generate
// a root secret on the P256 elliptic curve group, and splits it into n shares
// using Shamir's Secret Sharing scheme with threshold t. The threshold t is
// set to (ShamirThreshold - 1), meaning t+1 shares are required for
// reconstruction. A deterministic reader seeded with the root key is used to
// ensure identical share generation across restarts, which is critical for
// synchronization after crashes. The function performs security validation and
// zeroing of sensitive data after use.
//
// CRITICAL: This function MUST be called exactly once per process. Calling it
// multiple times will generate different root keys, breaking the cryptographic
// guarantees of the system. The function will terminate the application if
// called more than once.
func RootShares() []shamir.Share {
	const fName = "rootShares"

	// Ensure this function is only called once per process.
	rootSharesGeneratedMu.Lock()
	if rootSharesGenerated {
		log.FatalLn(
			fName,
			"message", "RootShares() called more than once",
			"err", "This is a critical error that would "+
				"generate different root keys and break the system",
		)
	}
	rootSharesGenerated = true
	rootSharesGeneratedMu.Unlock()

	rootKeySeedMu.Lock()
	defer rootKeySeedMu.Unlock()

	if _, err := rand.Read(rootKeySeed[:]); err != nil {
		log.FatalLn(fName, "message", "key seed failure", "err", err.Error())
	}

	// Initialize parameters
	g := group.P256
	t := uint(env.ShamirThresholdVal() - 1) // Need t+1 shares to reconstruct
	n := uint(env.ShamirSharesVal())        // Total number of shares

	log.Log().Info(
		fName,
		"message", "generating Shamir shares",
		"t", t, "n", n,
	)

	// Create a secret from our 32-byte key:
	rootSecret := g.NewScalar()

	if err := rootSecret.UnmarshalBinary(rootKeySeed[:]); err != nil {
		log.FatalLn(
			fName,
			"message", "key unmarshal failure",
			"err", err.Error(),
		)
	}

	// To compute identical shares, we need an identical seed for the random
	// reader. Using `finalKey` for seed is secure because Shamir Secret Sharing
	// algorithm's security does not depend on the random seed; it depends on
	// the shards being securely kept secret.
	// If we use `random.Read` instead, then synchronizing shards after Nexus
	// crashes will be cumbersome and prone to edge-case failures.
	reader := crypto.NewDeterministicReader(rootKeySeed[:])
	ss := shamir.New(reader, t, rootSecret)

	log.Log().Info(fName, "message", "generated Shamir shares")

	rs := ss.Share(n)

	// Security: Ensure the root key and shares are zeroed out after use.
	validation.VerifyShamirReconstruction(rootSecret, rs)

	log.Log().Info(fName, "message", "successfully generated shards")
	return rs
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
// identifier. The function will terminate the program with exit code 1 if the
// Keeper ID cannot be converted to an integer or if no matching share is found
// for the specified keeper.
func KeeperShare(
	rootShares []shamir.Share, keeperID string,
) shamir.Share {
	const fName = "keeperShare"

	var share shamir.Share
	for _, sr := range rootShares {
		kid, err := strconv.Atoi(keeperID)
		if err != nil {
			log.FatalLn(
				fName,
				"message", "failed to convert keeper id to int",
				"keeper_id", keeperID,
				"err", err.Error(),
			)
		}

		if sr.ID.IsEqual(group.P256.NewScalar().SetUint64(uint64(kid))) {
			share = sr
			break
		}
	}

	if share.ID.IsZero() {
		log.FatalLn(
			fName,
			"message", "failed to find share for keeper",
			"keeper_id", keeperID,
		)
	}

	return share
}
