//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"time"
)

// RecoveryOperationTimeout returns the recovery timeout duration.
// It reads from the SPIKE_NEXUS_RECOVERY_TIMEOUT environment variable
//
// If the environment variable is not set or is not a valid duration string
// then it defaults to `0` (no timeout limit).
func RecoveryOperationTimeout() time.Duration {
	e := os.Getenv("SPIKE_NEXUS_RECOVERY_TIMEOUT")
	if e != "" {
		if d, err := time.ParseDuration(e); err == nil {
			return d
		}
	}
	return 0
}
