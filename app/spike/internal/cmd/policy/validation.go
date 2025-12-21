//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"

	spike "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike/internal/config"
)

// validatePermissions is a wrapper around config.ValidatePermissions that
// validates policy permissions from a comma-separated string.
// See config.ValidatePermissions for details.
var validatePermissions = config.ValidatePermissions

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	policies, err := api.ListPolicies(ctx, "", "")
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
