//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// Initialize initializes the backend storage with the provided root key.
//
// For non-"in-memory" backing stores, if the root key is nil or empty,
// the application will crash for security.
//
// Parameters:
//   - r [32]byte: The root key to initialize the crypto state.
func Initialize(r *[crypto.AES256KeySize]byte) {
	const fName = "Initialize"

	// Locks on a mutex; so only a single process can access it.
	persist.InitializeBackend(r)

	// The in-memory store does not use a root key to operate.
	if env.BackendStoreType() == env.Memory {
		return
	}

	if r == nil || mem.Zeroed32(r) {
		log.FatalLn(fName, "message", "root key is nil or zeroed")
	}

	// Update the internal root key.
	// Locks on a mutex; so only a single process can modify the root key.
	SetRootKey(r)
}
