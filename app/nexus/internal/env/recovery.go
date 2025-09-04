//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"time"
)

// RecoveryOperationMaxInterval returns the maximum interval duration for
// recovery backoff retry algorithm. The interval is determined by the
// environment variable `SPIKE_NEXUS_RECOVERY_MAX_INTERVAL`.
//
// If the environment variable is not set or is not a valid duration
// string, then it defaults to 60 seconds.
func RecoveryOperationMaxInterval() time.Duration {
	e := os.Getenv("SPIKE_NEXUS_RECOVERY_MAX_INTERVAL")
	if e != "" {
		if d, err := time.ParseDuration(e); err == nil {
			return d
		}
	}
	return 60 * time.Second
}

// RecoveryKeeperUpdateInterval returns the duration between keeper updates for
// SPIKE Nexus. It first attempts to read the duration from the
// SPIKE_NEXUS_KEEPER_UPDATE_INTERVAL environment variable. If the environment
// variable is set and contains a valid duration string (as parsed by
// time.ParseDuration), that duration is returned. Otherwise, it returns a
// default value of 5 minutes.
func RecoveryKeeperUpdateInterval() time.Duration {
	e := os.Getenv("SPIKE_NEXUS_KEEPER_UPDATE_INTERVAL")
	if e != "" {
		if d, err := time.ParseDuration(e); err == nil {
			return d
		}
	}

	return 5 * time.Minute
}
