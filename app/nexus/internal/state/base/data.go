//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike-sdk-go/log"
)

// RootKeyNoLock returns a copy of the root key without acquiring the lock.
// This should only be used in contexts where the lock is already held
// or thread safety is managed externally.
//
// Returns:
//   - *[32]byte: Pointer to the root key
func RootKeyNoLock() *[crypto.AES256KeySize]byte {
	return &rootKey
}

// LockRootKey acquires an exclusive lock on the root key.
// This must be paired with a corresponding call to UnlockRootKey.
func LockRootKey() {
	rootKeyMu.Lock()
}

// UnlockRootKey releases an exclusive lock on the root key previously
// acquired with LockRootKey.
func UnlockRootKey() {
	rootKeyMu.Unlock()
}

// RootKeyZero checks if the root key contains only zero bytes.
//
// If the rot key is zero and SPIKE Nexus is not in "in memory" mode,
// then it means SPIKE Nexus has not been initialized yet, and any secret
// and policy management operation should be denied at the API level.
//
// Returns:
//   - bool: true if the root key contains only zeroes, false otherwise
func RootKeyZero() bool {
	rootKeyMu.RLock()
	defer rootKeyMu.RUnlock()

	for _, b := range rootKey[:] {
		if b != 0 {
			return false
		}
	}
	return true
}

// SetRootKey updates the root key with the provided value.
//
// This function does not own its parameter; the `rk` argument can be (and
// should be) cleaned up after calling this function without impacting the
// saved root key.
//
// Security behavior:
// The application will crash (via log.FatalErr) if rk is nil or contains only
// zero bytes. This is a defense-in-depth measure: the caller (Initialize)
// already validates the key, but if somehow an invalid key reaches this
// function, crashing is the correct response. Operating with a nil or zero
// root key would mean secrets are unencrypted or encrypted with a predictable
// key, which is a critical security failure.
//
// Note: For in-memory backends, this function should not be called at all.
// The Initialize function handles this by returning early for memory backends.
//
// Parameters:
//   - rk: Pointer to a 32-byte array containing the new root key value.
//     Must be non-nil and non-zero.
func SetRootKey(rk *[crypto.AES256KeySize]byte) {
	fName := "SetRootKey"

	log.Info(fName, "message", "setting root key")

	if rk == nil {
		failErr := *sdkErrors.ErrRootKeyMissing.Clone()
		log.FatalErr(fName, failErr)
		return
	}

	if mem.Zeroed32(rk) {
		failErr := *sdkErrors.ErrRootKeyEmpty.Clone()
		log.FatalErr(fName, failErr)
	}

	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()

	for i := range rootKey {
		rootKey[i] = rk[i]
	}

	log.Info(fName, "message", "root key set")
}
