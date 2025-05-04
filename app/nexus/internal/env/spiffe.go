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

func TrustRootForKeeper() string {
	tr := os.Getenv("SPIKE_TRUST_ROOT_KEEPER")
	if tr == "" {
		return "spike.ist"
	}
	return tr
}

func TrustRootForPilot() string {
	tr := os.Getenv("SPIKE_TRUST_ROOT_PILOT")
	if tr == "" {
		return "spike.ist"
	}
	return tr
}
