//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package config

import "os"

func SpiffeEndpointSocket() string {
	p := os.Getenv("SPIFFE_ENDPOINT_SOCKET")
	if p != "" {
		return p
	}

	return "unix:///tmp/spire-agent/public/api.sock"
}
