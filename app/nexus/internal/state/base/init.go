//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// Initialize initializes the backend storage with the provided root key.
//
// Security behavior:
//   - For "in-memory" backends: The root key is not used; passing nil is safe.
//   - For persistent backends (sqlite, lite): A valid, non-zero root key is
//     required. The application will crash (via log.FatalErr) if the key is
//     nil or zeroed. This is intentional: operating without a valid root key
//     would leave all secrets unencrypted or encrypted with a predictable key,
//     which is a critical security failure.
//
// The called SetRootKey function also validates the key as a defense-in-depth
// measure. If somehow an invalid key bypasses this function's validation,
// SetRootKey will also crash the application.
//
// Parameters:
//   - r: Pointer to a 32-byte AES-256 root key. Must be non-nil and non-zero
//     for persistent backends.
func Initialize(r *[crypto.AES256KeySize]byte) {
	const fName = "Initialize"

	log.Info(
		fName,
		"message", "initializing state",
		"backendType", env.BackendStoreTypeVal(),
	)

	// Locks on a mutex; so only a single process can access it.
	persist.InitializeBackend(r)

	// The in-memory store does not use a root key to operate.
	if env.BackendStoreTypeVal() == env.Memory {
		log.Info(
			fName,
			"message", "state initialized (in-memory mode, root key not used)",
			"backendType", env.BackendStoreTypeVal(),
		)
		return
	}

	if r == nil || mem.Zeroed32(r) {
		failErr := *sdkErrors.ErrRootKeyEmpty.Clone()
		log.FatalErr(fName, failErr)
	}

	// Update the internal root key.
	// Locks on a mutex; so only a single process can modify the root key.
	SetRootKey(r)

	log.Info(
		fName,
		"message", "state initialized",
		"backendType", env.BackendStoreTypeVal(),
	)
}
