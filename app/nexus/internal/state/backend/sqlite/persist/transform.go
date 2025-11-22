//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"strings"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// deserializePermissions converts a comma-separated string of permissions
// into a slice of PolicyPermission values.
//
// This function is used when loading policies from the SQLite database, where
// permissions are stored as a comma-separated string. Each permission value
// is trimmed of whitespace before being converted to a PolicyPermission type.
//
// Parameters:
//   - permissionsStr: A comma-separated string of permission values (e.g.,
//     "read,write,execute")
//
// Returns:
//   - []data.PolicyPermission: A slice of PolicyPermission values, or nil if
//     the input string is empty
func deserializePermissions(
	permissionsStr string,
) []data.PolicyPermission {
	if permissionsStr == "" {
		return nil
	}
	perms := strings.Split(permissionsStr, ",")
	permissions := make([]data.PolicyPermission, len(perms))
	for i, p := range perms {
		permissions[i] = data.PolicyPermission(strings.TrimSpace(p))
	}
	return permissions
}
