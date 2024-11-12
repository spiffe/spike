//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package trust

import (
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

// Authenticate validates the SPIFFE ID.
func Authenticate(spiffeid string) {
	if !config.IsNexus(spiffeid) {
		log.FatalF("Authenticate: SPIFFE ID %s is not valid.\n", spiffeid)
	}
}
