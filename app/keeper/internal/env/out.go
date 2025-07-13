//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"strings"
)

// BannerEnabled returns whether to show initial banner on app start based on
// the SPIKE_BANNER_ENABLED environment variable.
//
// The function reads the SPIKE_BANNER_ENABLED environment variable and returns:
//   - true if the variable is not set (default behavior)
//   - true if the variable is set to "true" (case-insensitive)
//   - false for any other value
//
// The environment variable value is trimmed of whitespace and converted to
// lowercase before comparison.
func BannerEnabled() bool {
	s := os.Getenv("SPIKE_BANNER_ENABLED")
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return true
	}
	return s == "true"
}
