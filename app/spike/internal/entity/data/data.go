//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package data

// Secret is the secret that returns from SPIKE Nexus mTLS REST API.
type Secret struct {
	Data map[string]string `json:"data"`
}
