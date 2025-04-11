//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"fmt"
	"github.com/spiffe/spike/internal/log"
	"sync"

	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike/app/nexus/internal/env"
)

// Global variables for storing secrets and policies with thread-safety.
var (
	// secretStore is a key-value store for managing secrets with version control.
	secretStore = kv.New(kv.Config{
		MaxSecretVersions: env.MaxSecretVersions(),
	})
	// secretStoreMu provides mutual exclusion for access to the secret store.
	secretStoreMu sync.RWMutex
)

// policies is a thread-safe map used to store policy information.
var policies sync.Map

// Global variables related to the root key with thread-safety protection.
var (
	// rootKey is a 32-byte array that stores the cryptographic root key.
	// It is initialized to zeroes by default.
	rootKey [32]byte
	// rootKeyMu provides mutual exclusion for access to the root key.
	rootKeyMu sync.RWMutex
)

// RootKeyNoLock returns a copy of the root key without acquiring the lock.
// This should only be used in contexts where the lock is already held
// or thread safety is managed externally.
//
// Returns:
//   - *[32]byte: Pointer to the root key
func RootKeyNoLock() *[32]byte {
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

//// ResetRootKey resets the root key to all zeroes.
//// This is typically used when clearing sensitive cryptographic material.
//func ResetRootKey() {
//	rootKeyMu.Lock()
//	defer rootKeyMu.Unlock()
//
//	// Explicitly reset the root key bytes to zeroes
//	mem.ClearRawBytes(&rootKey)
//}

// SetRootKey updates the root key with the provided value.
//
// Parameters:
//   - rk: Pointer to a 32-byte array containing the new root key value
func SetRootKey(rk *[32]byte) {
	fName := "SetRootKey"
	log.Log().Info(fName, "msg", "Setting root key")

	fmt.Printf("rk: %x\n", rk)
	fmt.Printf("Root key: %x\n", rootKey)

	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()

	for i := range rootKey {
		rootKey[i] = rk[i]
	}

	fmt.Printf("Root key: %x\n", rootKey)

	log.Log().Info(fName, "msg", "Root key set")
}
