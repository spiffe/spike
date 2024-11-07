//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"strconv"
	"time"
)

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

// PollInterval returns the polling interval duration for the Spike Nexus service.
// It reads from the SPIKE_NEXUS_POLL_INTERVAL environment variable which should
// contain a valid duration string (e.g. "1h", "30m", "1h30m").
// If the environment variable is not set or contains an invalid duration,
// it returns the default interval of 5 minutes.
func PollInterval() time.Duration {
	p := os.Getenv("SPIKE_NEXUS_POLL_INTERVAL")
	if p != "" {
		d, err := time.ParseDuration(p)
		if err == nil {
			return d
		}
	}

	return 5 * time.Minute
}

// MaxSecretVersions returns the maximum number of versions to retain for each secret.
// It reads from the SPIKE_NEXUS_MAX_SECRET_VERSIONS environment variable which should
// contain a positive integer value.
// If the environment variable is not set, contains an invalid integer, or specifies
// a non-positive value, it returns the default of 10 versions.
func MaxSecretVersions() int {
	p := os.Getenv("SPIKE_NEXUS_MAX_SECRET_VERSIONS")
	if p != "" {
		mv, err := strconv.Atoi(p)
		if err == nil && mv > 0 {
			return mv
		}
	}

	return 10
}
