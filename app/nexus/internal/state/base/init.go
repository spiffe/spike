//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"errors"
	"github.com/spiffe/spike/internal/log"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/nexus/internal/net"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/pkg/crypto"
)

var ErrAlreadyInitialized = errors.New("already initialized")

// Bootstrap sets up the root key if it hasn't been initialized yet.
// If a root key already exists, it returns immediately.
// The root key is generated using AES-256 encryption.
//
// If SPIKE has a root key, it is considered "bootstrapped", and this function
// will be a no-op. If SPIKE does not have a root key, it will generate one
// using AES-256 encryption.
//
// This function MUST be called ONCE during the application's startup.
//
// Note that this initialization is different from the initialization flow
// that is manually done by the admin through `spike init`.
//
// It is important to note that once the initialization is complete, the
// application is **guaranteed** to have a root key available for use.
//
// Returns:
//   - error: Any error encountered during initialization, nil on success
func Bootstrap(source *workloadapi.X509Source) error {
	log.Log().Info("boostrap", "msg", "bootstrapping")

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
		log.Log().Warn("boostrap", "msg", "already initialized. exiting")

		rootKeyMu.Lock()
		rootKey = existingRootKey
		rootKeyMu.Unlock()

		rootKeyMu.RLock()
		persist.InitializeBackend(existingRootKey)
		rootKeyMu.RUnlock()

		return ErrAlreadyInitialized
	}

	log.Log().Info("boostrap", "msg", "first time initialization: generating new root key")

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
