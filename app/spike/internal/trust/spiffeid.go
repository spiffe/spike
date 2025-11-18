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
)

// AuthenticateForPilot verifies if the provided SPIFFE ID belongs to a
// SPIKE Pilot instance. Logs a fatal error and exits if verification fails.
//
// SPIFFEID is the SPIFFE ID string to authenticate for pilot access.
func AuthenticateForPilot(SPIFFEID string) {
	const fName = "AuthenticateForPilot"
	if !svid.IsPilot(SPIFFEID) {
		log.Log().Error(
			fName,
			"message", "you need a 'super user' SPIFFE ID to use this command",
			"spiffe_id", SPIFFEID,
		)
		log.FatalLn(
			fName,
			"message", "you are not authorized to use this command",
			"spiffe_id", SPIFFEID,
		)
	}
}

// AuthenticateForPilotRecover validates the SPIFFE ID for the recover role
// and exits the application if it does not match the recover SPIFFE ID.
//
// SPIFFEID is the SPIFFE ID string to authenticate for pilot recover access.
func AuthenticateForPilotRecover(SPIFFEID string) {
	const fName = "AuthenticateForPilotRecover"

	if !svid.IsPilotRecover(SPIFFEID) {
		log.Log().Error(
			fName,
			"message", "You need a 'recover' SPIFFE ID to use this command",
			"spiffe_id", SPIFFEID,
		)
		log.FatalLn(
			fName,
			"message", "you are not authorized to use this command",
			"spiffe_id", SPIFFEID,
		)
	}
}

// AuthenticateForPilotRestore verifies if the given SPIFFE ID is valid for
// restoration. Logs a fatal error and exits if the SPIFFE ID validation fails.
//
// SPIFFEID is the SPIFFE ID string to authenticate for restore access.
func AuthenticateForPilotRestore(SPIFFEID string) {
	const fName = "AuthenticateForPilotRestore"

	if !svid.IsPilotRestore(SPIFFEID) {
		log.Log().Error(
			fName,
			"message", "you need a 'restore' SPIFFE ID to use this command",
			"spiffe_id", SPIFFEID,
		)
		log.FatalLn(
			fName,
			"message", "you are not authorized to use this command",
			"spiffe_id", SPIFFEID,
		)
	}
}
