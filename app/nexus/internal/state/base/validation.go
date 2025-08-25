//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

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
	// Super permission acts as a joker - grants all permissions
	if contains(haves, data.PermissionSuper) {
		return true
	}

	for _, want := range wants {
		if !contains(haves, want) {
			return false
		}
	}
	return true
}
