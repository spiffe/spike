//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import "os"

func TrustRoot() string {
	tr := os.Getenv("SPIKE_TRUST_ROOT")
	if tr == "" {
		return "spike.ist"
	}
	return tr
}

func TrustRootForNexus() string {
	tr := os.Getenv("SPIKE_TRUST_ROOT_NEXUS")
	if tr == "" {
		return "spike.ist"
	}
	return tr
}
