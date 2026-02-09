//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/spf13/cobra"
	spike "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// findPolicyByName searches for a policy with the given name and returns true
// if found. It returns an error if the policy cannot be found or if there's an
// issue with the API call.
//
// Parameters:
//   - api: The SPIKE API client
//   - name: The policy name to search for
//
// Returns:
//   - bool: true if policy exists
//   - *sdkErrors.SDKError: An error if the policy is not found or there's an
//     API issue
func findPolicyByName(
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

	failErr := sdkErrors.ErrEntityNotFound.Clone()
	failErr.Msg = "no policy found with name: " + name
	return false, failErr
}

// sendGetPolicyNameRequest gets the policy name either from command arguments
// or the name flag. If neither is provided, it returns an error.
//
// Parameters:
//   - cmd: The Cobra command containing the flags
//   - args: Command arguments that might contain the policy name
//   - api: The SPIKE API client
//
// Returns:
//   - string: The policy name
//   - *sdkErrors.SDKError: An error if the policy cannot be found or if name
//     is not provided
func sendGetPolicyNameRequest(cmd *cobra.Command,
	args []string, api *spike.API,
) (string, *sdkErrors.SDKError) {
	var policyName string

	name, _ := cmd.Flags().GetString("name")

	if len(args) > 0 {
		policyName = args[0]
	} else if name != "" {
		policyName = name
	} else {
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "policy name as argument or --name flag is required"
		return "", failErr
	}

	// Verify the policy exists
	_, err := findPolicyByName(api, policyName)
	if err != nil {
		return "", err
	}

	return policyName, nil
}
