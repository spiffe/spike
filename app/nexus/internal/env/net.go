//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import "os"

// TlsPort returns the TLS port for the Spike Nexus service.
// It reads from the SPIKE_NEXUS_TLS_PORT environment variable.
// If the environment variable is not set, it returns the default port ":8553".
func TlsPort() string {
	p := os.Getenv("SPIKE_NEXUS_TLS_PORT")
	if p != "" {
		return p
	}

	return ":8553"
}

// KeepApiRoot returns the base URL for the Spike Keep API. It first checks for
// a custom URL in the SPIKE_KEEP_API_URL environment variable. If no custom URL
// is set, it returns the default URL "http://localhost:8443".
func KeepApiRoot() string {
	p := os.Getenv("SPIKE_KEEP_API_URL")
	if p != "" {
		return p
	}
	return "https://localhost:8443"
}
