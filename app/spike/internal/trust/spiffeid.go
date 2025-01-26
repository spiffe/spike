//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package trust provides functions and utilities to manage and validate trust
// relationships using the SPIFFE standard. This package includes methods for
// authenticating SPIFFE IDs, ensuring secure identity verification in
// distributed systems.
package trust

import (
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
)

// Authenticate validates the SPIFFE ID.
func Authenticate(spiffeid string) {
	if !auth.IsPilot(spiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", spiffeid)
	}
}
