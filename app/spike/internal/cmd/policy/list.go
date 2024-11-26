//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/acl"
	"github.com/spiffe/spike/app/spike/internal/net/auth"
	"github.com/spiffe/spike/internal/entity/data"
)

// NewPolicyListCommand creates a new Cobra command for listing all policies.
// It retrieves and displays all existing policies in the system as formatted
// JSON output.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before listing policies.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy listing
//
// Command usage:
//
//	list
//
// Example usage:
//
//	spike policy list
//
// Example output:
//
//	[
//	  {
//	    "id": "policy-123",
//	    "name": "web-service-policy",
//	    "spiffe_id_pattern": "spiffe://example.org/web-service/*",
//	    "path_pattern": "/api/v1/*",
//	    "permissions": ["read", "write"],
//	    "created_at": "2024-01-01T00:00:00Z",
//	    "created_by": "user-abc"
//	  },
//	  {
//	    "id": "policy-456",
//	    "name": "db-service-policy",
//	    "spiffe_id_pattern": "spiffe://example.org/db/*",
//	    "path_pattern": "/data/*",
//	    "permissions": ["read"],
//	    "created_at": "2024-01-02T00:00:00Z",
//	    "created_by": "user-xyz"
//	  }
//	]
//
// The command will:
//  1. Check if the system is initialized
//  2. Retrieve all existing policies
//  3. Format the policy list as indented JSON
//  4. Display the formatted output
//
// Error conditions:
//   - System not initialized (requires running 'spike initialization' first)
//   - Insufficient permissions
//   - Policy retrieval failure
//   - JSON formatting failure
//
// Note: If no policies exist, an empty array ([]) will be displayed.
func NewPolicyListCommand(source *workloadapi.X509Source) *cobra.Command {
	return &cobra.Command{
		Use:   "policy list",
		Short: "List all policies",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			state, err := auth.CheckInitState(source)
			if err != nil {
				fmt.Println("Failed to check initialization state:", err)
				return
			}

			if state == data.NotInitialized {
				fmt.Println("Please initialize first by running 'spike initialization'.")
				return
			}

			policies, err := acl.ListPolicies(source)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			output, err := json.MarshalIndent(policies, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting output: %v\n", err)
				return
			}

			fmt.Println(string(output))
		},
	}
}
