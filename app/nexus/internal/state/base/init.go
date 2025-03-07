//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"fmt"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// Initialize initializes the backend storage with the provided root key.
func Initialize(r *[32]byte) {
	fmt.Println("calling state.Initialize !!!!!!")

	// No need for a lock:
	// This method is called only once during initial bootstrapping.
	persist.InitializeBackend(r)
}
