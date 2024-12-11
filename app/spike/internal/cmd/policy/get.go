//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// newPolicyGetCommand creates a new Cobra command for retrieving policy
// details. It fetches and displays the complete information about a specific
// policy in a formatted JSON output.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before retrieving policy information.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy retrieval
//
// Command usage:
//
//	get <policy-id>
//
// Arguments:
//   - policy-id: The unique identifier of the policy to retrieve (required)
//
// Example usage:
//
//	spike policy get policy-123
//
// Example output:
//
//	{
//	  "id": "policy-123",
//	  "name": "web-service-policy",
//	  "spiffe_id_pattern": "spiffe://example.org/web-service/*",
//	  "path_pattern": "/api/v1/*",
//	  "permissions": [
//	    "read",
//	    "write"
//	  ],
//	  "created_at": "2024-01-01T00:00:00Z",
//	  "created_by": "user-abc"
//	}
//
// The command will:
//  1. Check if the system is initialized
//  2. Retrieve the policy with the specified ID
//  3. Format the policy details as indented JSON
//  4. Display the formatted output
//
// Error conditions:
//   - Missing policy ID argument
//   - System not initialized (requires running 'spike init' first)
//   - Policy not found
//   - Insufficient permissions
//   - JSON formatting failure
func newPolicyGetCommand(source *workloadapi.X509Source) *cobra.Command {
	return &cobra.Command{
		Use:   "get <policy-id>",
		Short: "Get policy details",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			api := spike.NewWithSource(source)

			state, err := api.CheckInitState()
			if err != nil {
				fmt.Println("Failed to check initialization state:", err)
				return
			}

			if state == data.NotInitialized {
				fmt.Println("Please initialize first by running 'spike init'.")
				return
			}

			policyID := args[0]
			policy, err := api.GetPolicy(policyID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			output, err := json.MarshalIndent(policy, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting output: %v\n", err)
				return
			}

			fmt.Println(string(output))
		},
	}
}
