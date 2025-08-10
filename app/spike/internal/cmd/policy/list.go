//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
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

// newPolicyListCommand creates a new Cobra command for listing policies.
// It retrieves and displays policies, optionally filtering by a resource path
// pattern or a SPIFFE ID pattern.
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
//	list [--format=<format>] [--path=<pattern> | --spiffeid=<pattern>]
//
// Flags:
//   - --format: Output format ("human" or "json", default is "human")
//   - --path: Filter policies by a resource path pattern (e.g., '/secrets/*')
//   - --spiffeid: Filter policies by a SPIFFE ID pattern (e.g., 'spiffe://example.org/service/*')
//
// Note: --path and --spiffeid flags cannot be used together.
//
// Example usage:
//
//	spike policy list
//	spike policy list --format=json
//	spike policy list --path="/api/v1/*"
//	spike policy list --spiffeid="spiffe://example.org/app"
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
//  2. Retrieve existing policies based on filters
//  3. Format the policies based on the format flag
//  4. Display the formatted output
//
// Error conditions:
//   - System not initialized (requires running 'spike init' first)
//   - An invalid format specified
//   - Using --path and --spiffeid flags together
//   - Insufficient permissions
//   - Policy retrieval failure
//
// Note: If no policies exist, it returns "No policies found" for human format
// or "[]" for JSON format.
func newPolicyListCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var (
		pathPattern     string
		SPIFFEIDPattern string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List policies, optionally filtering by path or SPIFFE ID",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			trust.Authenticate(SPIFFEID)
			api := spike.NewWithSource(source)

			policies, err := api.ListPolicies(SPIFFEIDPattern, pathPattern)
			if handleAPIError(err) {
				return
			}

			output := formatPoliciesOutput(cmd, policies)
			fmt.Println(output)
		},
	}

	cmd.Flags().StringVar(&pathPattern, "path", "",
		"Resource path pattern, e.g., '/secrets/*'")
	cmd.Flags().StringVar(&SPIFFEIDPattern, "spiffeid", "",
		"SPIFFE ID pattern, e.g., 'spiffe://example.org/service/*'")
	cmd.MarkFlagsMutuallyExclusive("path", "spiffeid")

	addFormatFlag(cmd)
	return cmd
}
