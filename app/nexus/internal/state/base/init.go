//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"errors"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/nexus/internal/net"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/pkg/crypto"
)

var ErrAlreadyInitialized = errors.New("already initialized")

// Initialize sets up the root key if it hasn't been initialized yet.
// If a root key already exists, it returns immediately.
// The root key is generated using AES-256 encryption.
//
// This function MUST be called ONCE during the application's startup.
//
// It is important to note that once the initialization is complete, the
// application is **guaranteed** to have a root key available for use.
//
// Returns:
//   - error: Any error encountered during initialization, nil on success
func Initialize(source *workloadapi.X509Source) error {
	// TODO: this logic is complicated, and can cause trouble in edge cases.
	// 1. Use a tombstone to indicate initialization.
	// 2. Use a keeper call too, to indicate initialization.
	// 3. Check for both, and it they both fail, initialize.

	existingRootKey := RootKey()
	if existingRootKey == "" {
		// Check if SPIKE Keeper has a cached root key first:
		key, err := net.FetchFromCache(source)
		if err != nil {
			return err
		}
		if key != "" {
			SetRootKey(key)
			existingRootKey = key
		}
	}

	if existingRootKey != "" {
		rootKeyMu.Lock()
		rootKey = existingRootKey
		rootKeyMu.Unlock()

		rootKeyMu.RLock()
		persist.InitializeBackend(existingRootKey)
		rootKeyMu.RUnlock()

		return ErrAlreadyInitialized
	}

	r, err := crypto.Aes256Seed()
	if err != nil {
		return err
	}

	rootKeyMu.Lock()
	rootKey = r
	rootKeyMu.Unlock()

	rootKeyMu.RLock()
	persist.InitializeBackend(rootKey)
	rootKeyMu.RUnlock()

	return nil
}
