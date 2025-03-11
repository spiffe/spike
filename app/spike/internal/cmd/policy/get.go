//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newPolicyGetCommand creates a new Cobra command for retrieving policy
// details. It fetches and displays the complete information about a specific
// policy by ID or name.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before retrieving policy information.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication
//   - spiffeId: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy retrieval
//
// Command usage:
//
//	get [policy-id] [flags]
//
// Arguments:
//   - policy-id: The unique identifier of the policy to retrieve
//     (optional if --name is provided)
//
// Flags:
//   - --name: Policy name to look up (alternative to policy ID)
//   - --format: Output format ("human" or "json", default is "human")
//
// Example usage:
//
//	spike policy get abc123
//	spike policy get --name=web-service-policy
//	spike policy get abc123 --format=json
//
// Example output for human format:
//
//	POLICY DETAILS
//	=============
//
//	ID: policy-123
//	Name: web-service-policy
//	SPIFFE ID Pattern: spiffe://example.org/web-service/*
//	Path Pattern: /api/v1/*
//	Permissions: read, write
//	Created At: 2024-01-01T00:00:00Z
//	Created By: user-abc
//
// Example output for JSON format:
//
//	{
//	  "id": "policy-123",
//	  "name": "web-service-policy",
//	  "spiffeIdPattern": "spiffe://example.org/web-service/*",
//	  "pathPattern": "/api/v1/*",
//	  "permissions": ["read", "write"],
//	  "createdAt": "2024-01-01T00:00:00Z",
//	  "createdBy": "user-abc"
//	}
//
// The command will:
//  1. Check if the system is initialized
//  2. Get the policy ID either from arguments or by looking up the policy name
//  3. Retrieve the policy with the specified ID
//  4. Format the policy details based on the format flag
//  5. Display the formatted output
//
// Error conditions:
//   - Neither policy ID argument nor --name flag provided
//   - Policy not found by ID or name
//   - Invalid format specified
//   - System not initialized (requires running 'spike init' first)
//   - Insufficient permissions
func newPolicyGetCommand(
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [policy-id]",
		Short: "Get policy details",
		Long: `Get detailed information about a policy by ID or name.

        You can provide either:
        - A policy ID as an argument: spike policy get abc123
        - A policy name with the --name flag: spike policy get --name=mypolicy

        Use --format=json to get the output in JSON format.`,
		Run: func(cmd *cobra.Command, args []string) {
			trust.Authenticate(spiffeId)
			api := spike.NewWithSource(source)

			// If first argument is provided without --name flag, it could be
			// misinterpreted as trying to use policy name directly
			if len(args) > 0 && !cmd.Flags().Changed("name") {
				fmt.Println("Note: To look up a policy by name, use --name flag:")
				fmt.Printf("  spike policy get --name=%s\n\n", args[0])
				fmt.Printf("Attempting to use '%s' as policy ID...\n", args[0])
			}

			policyId, err := sendGetPolicyIdRequest(cmd, args, api)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			policy, err := api.GetPolicy(policyId)
			if handleAPIError(err) {
				return
			}

			if policy == nil {
				fmt.Println("Error: Got empty response from server")
				return
			}

			output := formatPolicy(cmd, policy)
			fmt.Println(output)
		},
	}

	addNameFlag(cmd)
	addFormatFlag(cmd)
	return cmd
}
