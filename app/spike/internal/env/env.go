//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import "os"

// TrustRoot returns the trust root domain(s) for the application.
//
// It retrieves the trust root from the SPIKE_TRUST_ROOT environment variable.
// If the environment variable is not set, it returns the default value
// "spike.ist". The return value can be a comma-delimited string of multiple
// trust root domains.
//
// Returns:
//   - A string containing one or more trust root domains, comma-delimited if
//     multiple
func TrustRoot() string {
	tr := os.Getenv("SPIKE_TRUST_ROOT")
	if tr == "" {
		return "spike.ist"
	}
	return tr
}
