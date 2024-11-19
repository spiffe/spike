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
// Returns:
//   - error: Any error encountered during initialization, nil on success
func Initialize(source *workloadapi.X509Source) error {
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

		persist.InitializeBackend(existingRootKey)

		return ErrAlreadyInitialized
	}

	r, err := crypto.Aes256Seed()
	if err != nil {
		return err
	}

	rootKeyMu.Lock()
	rootKey = r
	rootKeyMu.Unlock()

	persist.InitializeBackend(rootKey)

	return nil
}
