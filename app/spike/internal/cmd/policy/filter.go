//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	spike "github.com/spiffe/spike-sdk-go/api"
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
//   - error: An error if the policy is not found or there's an API issue
func findPolicyByName(api *spike.Api, name string) (string, error) {
	policies, err := api.ListPolicies()
	if err != nil {
		return "", err
	}

	if policies != nil {
		for _, policy := range *policies {
			if policy.Name == name {
				return policy.Id, nil
			}
		}
	}

	return "", fmt.Errorf("no policy found with name '%s'", name)
}

const uuidRegex = `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`

func validUuid(uuid string) bool {
	r := regexp.MustCompile(uuidRegex)
	return r.MatchString(strings.ToLower(uuid))
}

// sendGetPolicyIdRequest gets the policy ID either from command arguments or
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
//   - error: An error if the policy cannot be found or if neither ID nor name
//     is provided
func sendGetPolicyIdRequest(cmd *cobra.Command,
	args []string, api *spike.Api,
) (string, error) {
	var policyId string

	name, _ := cmd.Flags().GetString("name")

	if len(args) > 0 {
		policyId = args[0]

		if !validUuid(policyId) {
			return "", fmt.Errorf("invalid policy ID '%s'", policyId)
		}

	} else if name != "" {
		id, err := findPolicyByName(api, name)
		if err != nil {
			return "", err
		}
		policyId = id
	} else {
		return "", fmt.Errorf(
			"either policy ID as argument or --name flag is required",
		)
	}

	return policyId, nil
}
