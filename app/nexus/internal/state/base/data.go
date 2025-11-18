//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"sync"

	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike-sdk-go/log"
)

// Global variables related to the root key with thread-safety protection.
var (
	// rootKey is a 32-byte array that stores the cryptographic root key.
	// It is initialized to zeroes by default.
	rootKey [crypto.AES256KeySize]byte
	// rootKeyMu provides mutual exclusion for access to the root key.
	rootKeyMu sync.RWMutex
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
// This function does not own its parameter; the `rk` argument can
// be (and should be) cleaned up after calling this function without
// impacting the saved root key.
//
// To ensure the system always has a legitimate root key, the operation is a
// no-op if rk is nil or zeroed out. When that happens, the function logs
// a warning.
//
// Parameters:
//   - rk: Pointer to a 32-byte array containing the new root key value
func SetRootKey(rk *[crypto.AES256KeySize]byte) {
	fName := "SetRootKey"

	log.Log().Info(fName, "message", "setting root key")

	if rk == nil {
		log.Log().Warn(fName, "message", sdkErrors.ErrCodeRootKeyEmpty)
		return
	}

	if mem.Zeroed32(rk) {
		log.Log().Warn(fName, "message", sdkErrors.ErrCodeRootKeyEmpty)
		return
	}

	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()

	for i := range rootKey {
		rootKey[i] = rk[i]
	}

	log.Log().Info(fName, "message", "root key set")
}
