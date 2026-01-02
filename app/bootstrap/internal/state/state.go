//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"fmt"
	"strconv"

	"github.com/cloudflare/circl/group"
	shamir "github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// RootKeySeed returns a pointer to the root key seed used for encryption.
// This key is generated when RootShares() is called and persists in memory
// for the duration of the bootstrap process. This function acquires a read
// lock to ensure thread-safe access to the root key seed.
//
// Returns:
//   - *[32]byte: Pointer to the root key seed
func RootKeySeed() *[crypto.AES256KeySize]byte {
	rootKeySeedMu.RLock()
	defer rootKeySeedMu.RUnlock()
	return &rootKeySeed
}

// RootKeySeedNoLock returns a pointer to the root key seed without acquiring
// any lock. The caller must hold the lock via LockRootKeySeed before calling
// this function and release it via UnlockRootKeySeed when done.
//
// This function exists to support patterns where the caller needs to perform
// multiple operations on the root key seed atomically, or when using defer
// in a loop would cause resource leaks.
//
// Returns:
//   - *[32]byte: Pointer to the root key seed
func RootKeySeedNoLock() *[crypto.AES256KeySize]byte {
	return &rootKeySeed
}

// LockRootKeySeed acquires an exclusive write lock on the root key seed.
// This must be paired with UnlockRootKeySeed to release the lock.
//
// Use this in combination with RootKeySeedNoLock when you need explicit lock
// control, such as avoiding defer in loops or performing multiple operations
// atomically.
func LockRootKeySeed() {
	rootKeySeedMu.Lock()
}

// UnlockRootKeySeed releases the exclusive write lock on the root key seed.
// This must be called after LockRootKeySeed to avoid deadlocks.
func UnlockRootKeySeed() {
	rootKeySeedMu.Unlock()
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
