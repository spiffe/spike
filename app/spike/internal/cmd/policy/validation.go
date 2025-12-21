//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	spike "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/validation"
)

// validatePermissions is a wrapper around validation.ValidatePermissions that
// validates policy permissions from a comma-separated string.
// See validation.ValidatePermissions for details.
var validatePermissions = validation.ValidatePermissions

// checkPolicyNameExists checks if a policy with the given name already exists.
//
// Parameters:
//   - api: The SPIKE API client
//   - name: The policy name to check
//
// Returns:
//   - bool: true if a policy with the name exists, false otherwise
//   - *sdkErrors.SDKError: An error if there is an issue with the API call
func checkPolicyNameExists(
	api *spike.API, name string,
) (bool, *sdkErrors.SDKError) {
	policies, err := api.ListPolicies("", "")
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
