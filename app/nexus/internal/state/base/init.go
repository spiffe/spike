//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// Initialize initializes the backend storage with the provided root key.
func Initialize(r *[shardSize]byte) {
	// No need for a lock:
	// This method is called only once during initial bootstrapping.
	persist.InitializeBackend(r)
	// Update internal root key.
	SetRootKey(r)
}
