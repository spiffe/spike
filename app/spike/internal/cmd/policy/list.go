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

// newPolicyListCommand creates a new Cobra command for listing all policies.
// It retrieves and displays all existing policies in the system.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before listing policies.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication
//   - spiffeId: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy listing
//
// Command usage:
//
//	list [--format=<format>]
//
// Flags:
//   - --format: Output format ("human" or "json", default is "human")
//
// Example usage:
//
//	spike policy list
//	spike policy list --format=json
//
// Example output for human format:
//
//	POLICIES
//	========
//
//	ID: policy-123
//	Name: web-service-policy
//	SPIFFE ID Pattern: spiffe://example.org/web-service/*
//	Path Pattern: /api/v1/*
//	Permissions: read, write
//	Created At: 2024-01-01T00:00:00Z
//	Created By: user-abc
//	--------
//
// Example output for JSON format:
//
//	[
//	  {
//	    "id": "policy-123",
//	    "name": "web-service-policy",
//	    "spiffeIdPattern": "spiffe://example.org/web-service/*",
//	    "pathPattern": "/api/v1/*",
//	    "permissions": ["read", "write"],
//	    "createdAt": "2024-01-01T00:00:00Z",
//	    "createdBy": "user-abc"
//	  }
//	]
//
// The command will:
//  1. Check if the system is initialized
//  2. Retrieve all existing policies
//  3. Format the policies based on the format flag
//  4. Display the formatted output
//
// Error conditions:
//   - System not initialized (requires running 'spike init' first)
//   - Invalid format specified
//   - Insufficient permissions
//   - Policy retrieval failure
//
// Note: If no policies exist, it returns "No policies found" for human format
// or "[]" for JSON format.
func newPolicyListCommand(
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all policies",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			trust.Authenticate(spiffeId)
			api := spike.NewWithSource(source)

			policies, err := api.ListPolicies()
			if handleAPIError(err) {
				return
			}

			output := formatPoliciesOutput(cmd, policies)
			fmt.Println(output)
		},
	}

	addFormatFlag(cmd)
	return cmd
}
