//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newPolicyGetCommand creates a new Cobra command for retrieving policy
// details. It fetches and displays the complete information about a specific
// policy by name.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before retrieving policy information.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy retrieval
//
// Command usage:
//
//	get [policy-name] [flags]
//
// Arguments:
//   - policy-name: The name of the policy to retrieve
//     (optional if --name is provided)
//
// Flags:
//   - --name: Policy name to look up
//   - --format: Output format ("human" or "json", default is "human")
//
// Example usage:
//
//	spike policy get web-service-policy
//	spike policy get --name=web-service-policy
//	spike policy get web-service-policy --format=json
//
// Example output for human format:
//
//	POLICY DETAILS
//	=============
//
//	Name: web-service-policy
//	SPIFFE ID Pattern: ^spiffe://example\.org/web-service/.*$
//	Path Pattern: ^/secrets/db/.*$
//	Permissions: read, write
//	Created At: 2024-01-01T00:00:00Z
//	Created By: user-abc
//
// Example output for JSON format:
//
//	{
//	  "name": "web-service-policy",
//	  "spiffeIdPattern": "^spiffe://example\\.org/web-service/.*$",
//	  "pathPattern": "^tenants/demo/db$",
//	  "permissions": ["read", "write"],
//	  "createdAt": "2024-01-01T00:00:00Z",
//	  "createdBy": "user-abc"
//	}
//
// The command will:
//  1. Check if the system is initialized
//  2. Get the policy name from arguments or --name flag
//  3. Retrieve the policy with the specified name
//  4. Format the policy details based on the format flag
//  5. Display the formatted output
//
// Error conditions:
//   - Neither policy name argument nor --name flag provided
//   - Policy not found by name
//   - Invalid format specified
//   - System not initialized (requires running 'spike init' first)
//   - Insufficient permissions
func newPolicyGetCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [policy-name]",
		Short: "Get policy details",
		Long: `Get detailed information about a policy by name.

        You can provide either:
        - A policy name as an argument: spike policy get web-service-policy
        - A policy name with the --name flag: spike policy get --name=my-policy

        Use --format=json to get the output in JSON format.`,
		Run: func(c *cobra.Command, args []string) {
			trust.AuthenticateForPilot(SPIFFEID)

			if source == nil {
				c.PrintErrln("Error: SPIFFE X509 source is unavailable.")
				return
			}

			api := spike.NewWithSource(source)

			// TODO: Issue #250 - Using name as primary identifier.
			// The SDK still uses 'id' field in the API call.
			policyName, err := sendGetPolicyNameRequest(c, args, api)
			if stdout.HandleAPIError(c, err) {
				return
			}

			policy, apiErr := api.GetPolicy(policyName)
			if stdout.HandleAPIError(c, apiErr) {
				return
			}

			if policy == nil {
				c.PrintErrln("Error: Empty response from server.")
				return
			}

			output := formatPolicy(c, policy)
			c.Println(output)
		},
	}

	addNameFlag(cmd)
	addFormatFlag(cmd)

	return cmd
}
