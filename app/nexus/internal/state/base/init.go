//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

func Initialize(r string) {
	// No need for a lock; this method is called only once during initial
	// bootstrapping.

	persist.InitializeBackend(r)
}
