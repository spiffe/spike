//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"sync"

	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike/app/nexus/internal/env"
)

// TODO: add documentation.

var (
	secretStore = kv.New(kv.Config{
		MaxSecretVersions: env.MaxSecretVersions(),
	})
	secretStoreMu sync.RWMutex
)

var policies sync.Map

var (
	// An array of 32 bytes initialized to zeroes.
	rootKey   [32]byte
	rootKeyMu sync.RWMutex
)

func RootKey() *[32]byte {
	rootKeyMu.RLock()
	defer rootKeyMu.RUnlock()
	return &rootKey
}

func RootKeyNoLock() [32]byte {
	return rootKey
}

func LockRootKey() {
	rootKeyMu.Lock()
}

func UnlockRootKey() {
	rootKeyMu.Unlock()
}

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

func ResetRootKey() {
	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()

	// Explicitly reset the root key bytes to zeroes
	for i := range rootKey {
		rootKey[i] = 0
	}
}

func SetRootKey(rk *[32]byte) {
	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()

	for i := range rootKey {
		rootKey[i] = rk[i]
	}
}
