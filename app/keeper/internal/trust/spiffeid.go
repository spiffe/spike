//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package trust provides mechanisms for verifying and validating trust among
// entities using SPIFFE. It includes functionalities such as authenticating
// SPIFFE IDs to ensure secure communication and interaction.
package trust

import (
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
)

// Authenticate validates the SPIFFE ID.
func Authenticate(spiffeid string) {
	if !auth.IsKeeper(spiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", spiffeid)
	}
}
