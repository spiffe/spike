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

// Authenticate verifies if the provided SPIFFE ID belongs to a pilot instance.
// Logs a fatal error and exits if verification fails.
func Authenticate(spiffeid string) {
	if !auth.IsPilot(spiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", spiffeid)
	}
}

// AuthenticateRecover validates the SPIFFE ID for the recover role and exits
// the application if it does not match the recover SPIFFE ID.
func AuthenticateRecover(spiffeid string) {
	if !auth.IsPilotRecover(spiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", spiffeid)
	}
}

// AuthenticateRestore verifies if the given SPIFFE ID is valid for restoration.
// Logs a fatal error and exits if the SPIFFE ID validation fails.
func AuthenticateRestore(spiffeid string) {
	if !auth.IsPilotRestore(spiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", spiffeid)
	}
}
