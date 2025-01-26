//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import "github.com/spiffe/spike-sdk-go/api/entity/data"

func contains(permissions []data.PolicyPermission,
	permission data.PolicyPermission) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func hasAllPermissions(
	haves []data.PolicyPermission,
	wants []data.PolicyPermission,
) bool {
	for _, want := range wants {
		if !contains(haves, want) {
			return false
		}
	}
	return true
}
