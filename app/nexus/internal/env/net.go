//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"

	"github.com/spiffe/spike-sdk-go/config/env"
)

// TLSPort returns the TLS port for the Spike Nexus service.
// It reads from the SPIKE_NEXUS_TLS_PORT environment variable.
// If the environment variable is not set, it returns the default port ":8553".
func TLSPort() string {
	p := os.Getenv(env.NexusTLSPort)
	if p != "" {
		return p
	}

	return ":8553"
}
