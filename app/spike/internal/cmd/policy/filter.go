//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	spike "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// findPolicyByName searches for a policy with the given name and returns its
// ID. It returns an error if the policy cannot be found or if there's an issue
// with the API call.
//
// Parameters:
//   - api: The SPIKE API client
//   - name: The policy name to search for
//
// Returns:
//   - string: The policy ID if found
//   - *sdkErrors.SDKError: An error if the policy is not found or there's an
//     API issue
func findPolicyByName(
	api *spike.API, name string,
) (string, *sdkErrors.SDKError) {
	ctx := context.Background()

	policies, err := api.ListAllPolicies(ctx)
	if err != nil {
		return "", err
	}

	if policies != nil {
		for _, policy := range *policies {
			if policy.Name == name {
				return policy.ID, nil
			}
		}
	}

	failErr := sdkErrors.ErrEntityNotFound.Clone()
	failErr.Msg = "no policy found with name: " + name
	return "", failErr
}

const uuidRegex = `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`

func validUUID(uuid string) bool {
	r := regexp.MustCompile(uuidRegex)
	return r.MatchString(strings.ToLower(uuid))
}

// sendGetPolicyIDRequest gets the policy ID either from command arguments or
// the name flag.
// If args contains a policy ID, it returns that. If the name flag is provided,
// it looks up the policy by name and returns its ID. If neither is provided,
// it returns an error.
//
// Parameters:
//   - cmd: The Cobra command containing the flags
//   - args: Command arguments that might contain the policy ID
//   - api: The SPIKE API client
//
// Returns:
//   - string: The policy ID
//   - *sdkErrors.SDKError: An error if the policy cannot be found or if neither
//     ID nor name is provided
func sendGetPolicyIDRequest(cmd *cobra.Command,
	args []string, api *spike.API,
) (string, *sdkErrors.SDKError) {
	var policyID string

	name, _ := cmd.Flags().GetString("name")

	if len(args) > 0 {
		policyID = args[0]

		if !validUUID(policyID) {
			failErr := sdkErrors.ErrDataInvalidInput.Clone()
			failErr.Msg = "invalid policy ID: " + policyID
			return "", failErr
		}

	} else if name != "" {
		id, err := findPolicyByName(api, name)
		if err != nil {
			return "", err
		}
		policyID = id
	} else {
		failErr := sdkErrors.ErrDataInvalidInput.Clone()
		failErr.Msg = "either policy ID as argument or --name flag is required"
		return "", failErr
	}

	return policyID, nil
}
