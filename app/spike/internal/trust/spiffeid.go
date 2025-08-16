//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package trust provides functions and utilities to manage and validate trust
// relationships using the SPIFFE standard. This package includes methods for
// authenticating SPIFFE IDs, ensuring secure identity verification in
// distributed systems.
package trust

import (
	"github.com/spiffe/spike-sdk-go/log"

	svid "github.com/spiffe/spike-sdk-go/spiffeid"
	"github.com/spiffe/spike/app/spike/internal/env"
)

// Authenticate verifies if the provided SPIFFE ID belongs to a pilot instance.
// Logs a fatal error and exits if verification fails.
func Authenticate(SPIFFEID string) {
	if !svid.IsPilot(env.TrustRoot(), SPIFFEID) {
		log.Log().Error(
			"Authenticate: You need a 'super user' SPIFFE ID to use this command.",
		)
		log.FatalF(
			"Authenticate: You are not authorized to use this command (%s).\n",
			SPIFFEID,
		)
	}
}

// AuthenticateRecover validates the SPIFFE ID for the recover role and exits
// the application if it does not match the recover SPIFFE ID.
func AuthenticateRecover(SPIFFEID string) {
	if !svid.IsPilotRecover(env.TrustRoot(), SPIFFEID) {
		log.Log().Error(
			"AuthenticateRecover: You need a 'recover' " +
				"SPIFFE ID to use this command.",
		)
		log.FatalF(
			"AuthenticateRecover: You are not authorized to use this command (%s).\n",
			SPIFFEID,
		)
	}
}

// AuthenticateRestore verifies if the given SPIFFE ID is valid for restoration.
// Logs a fatal error and exits if the SPIFFE ID validation fails.
func AuthenticateRestore(SPIFFEID string) {
	if !svid.IsPilotRestore(env.TrustRoot(), SPIFFEID) {
		log.Log().Error(
			"AuthenticateRestore: You need a 'restore' " +
				"SPIFFE ID to use this command.",
		)
		log.FatalF(
			"AuthenticateRecover: You are not authorized to use this command (%s).\n",
			SPIFFEID,
		)
	}
}
