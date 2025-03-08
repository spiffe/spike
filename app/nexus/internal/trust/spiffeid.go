//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package trust provides functionalities for managing and validating
// trust relationships using SPIFFE IDs. It includes methods for
// handling authentication and other trust-related operations.
package trust

import (
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
)

// Authenticate validates the SPIFFE Id.
func Authenticate(spiffeid string) {
	if !auth.IsNexus(spiffeid) {
		log.FatalF("Authenticate: SPIFFE Id %s is not valid.\n", spiffeid)
	}
}
