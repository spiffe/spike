//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"strings"

	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// validatePermissions validates policy permissions from a comma-separated string
// and returns a slice of PolicyPermission values. Only "read", "write", "list",
// and "super" are valid permissions. It returns an error if any permission is
// invalid or if the string contains no valid permissions.
//
// Parameters:
//   - permsStr: Comma-separated string of permissions (e.g., "read,write,list")
//
// Returns:
//   - []data.PolicyPermission: Validated policy permissions
//   - error: An error if any permission is invalid
func validatePermissions(permsStr string) ([]data.PolicyPermission, error) {
	validPerms := map[string]bool{
		"read":  true,
		"write": true,
		"list":  true,
		"super": true,
	}

	var permissions []string
	for _, p := range strings.Split(permsStr, ",") {
		perm := strings.TrimSpace(p)
		if perm != "" {
			permissions = append(permissions, perm)
		}
	}

	perms := make([]data.PolicyPermission, 0, len(permissions))
	for _, perm := range permissions {
		if _, ok := validPerms[perm]; !ok {
			validPermsList := "read, write, list, super"
			return nil, fmt.Errorf(
				"invalid permission '%s'. Valid permissions are: %s",
				perm, validPermsList)
		}
		perms = append(perms, data.PolicyPermission(perm))
	}

	if len(perms) == 0 {
		return nil,
			fmt.Errorf("no valid permissions specified. " +
				"Valid permissions are: read, write, list, super")
	}

	return perms, nil
}

// checkPolicyNameExists checks if a policy with the given name already exists.
//
// Parameters:
//   - api: The SPIKE API client
//   - name: The policy name to check
//
// Returns:
//   - bool: true if a policy with the name exists, false otherwise
//   - error: An error if there's an issue with the API call
func checkPolicyNameExists(api *spike.Api, name string) (bool, error) {
	policies, err := api.ListPolicies()
	if err != nil {
		return false, err
	}

	if policies != nil {
		for _, policy := range *policies {
			if policy.Name == name {
				return true, nil
			}
		}
	}

	return false, nil
}
