//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import "os"

// NexusBaseURL returns the base URL for SPIKE Nexus.
// It prefers environment overrides and otherwise falls back to the
// default in-cluster DNS name used by deployments.
//
// Environment variables:
//   - SPIKE_NEXUS_HOST: scheme and host (e.g., https://spire-spike-nexus.spire-server)
//   - SPIKE_NEXUS_TLS_PORT: port string (e.g., :8553). If empty, default port is used by server.
func NexusBaseURL() string {
	host := os.Getenv("SPIKE_NEXUS_HOST")
	if host == "" {
		host = "https://spire-spike-nexus.spire-server"
	}

	port := os.Getenv("SPIKE_NEXUS_TLS_PORT")
	if port == "" {
		return host
	}

	if port[0] != ':' {
		port = ":" + port
	}

	return host + port
}
